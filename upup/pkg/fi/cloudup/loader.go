/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudup

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/assettasks"
	"k8s.io/kops/util/pkg/reflectutils"
)

type Loader struct {
	Cluster *kopsapi.Cluster

	TemplateFunctions template.FuncMap

	Resources map[string]fi.Resource

	Builders []fi.ModelBuilder

	tasks map[string]fi.Task
}

func (l *Loader) Init() {
	l.tasks = make(map[string]fi.Task)
	l.Resources = make(map[string]fi.Resource)
	l.TemplateFunctions = make(template.FuncMap)
}

func (l *Loader) BuildTasks(assetBuilder *assets.AssetBuilder, lifecycle *fi.Lifecycle, lifecycleOverrides map[string]fi.Lifecycle) (map[string]fi.Task, error) {
	// Second pass: load everything else
	for _, builder := range l.Builders {
		context := &fi.ModelBuilderContext{
			Tasks:              l.tasks,
			LifecycleOverrides: lifecycleOverrides,
		}
		err := builder.Build(context)
		if err != nil {
			return nil, err
		}
		l.tasks = context.Tasks
	}

	if err := l.addAssetCopyTasks(assetBuilder.ContainerAssets, lifecycle); err != nil {
		return nil, err
	}

	if err := l.addAssetFileCopyTasks(assetBuilder.FileAssets, lifecycle); err != nil {
		return nil, err
	}
	if err := l.processDeferrals(); err != nil {
		return nil, err
	}
	return l.tasks, nil
}

func (l *Loader) addAssetCopyTasks(assets []*assets.ContainerAsset, lifecycle *fi.Lifecycle) error {
	for _, asset := range assets {
		if asset.CanonicalLocation != "" && asset.DockerImage != asset.CanonicalLocation {
			context := &fi.ModelBuilderContext{
				Tasks: l.tasks,
			}

			copyImageTask := &assettasks.CopyDockerImage{
				Name:        fi.String(asset.DockerImage),
				SourceImage: fi.String(asset.CanonicalLocation),
				TargetImage: fi.String(asset.DockerImage),
				Lifecycle:   lifecycle,
			}

			if err := context.EnsureTask(copyImageTask); err != nil {
				return fmt.Errorf("error adding asset-copy task: %v", err)
			}

			l.tasks = context.Tasks

		}
	}

	return nil
}

// addAssetFileCopyTasks creates the new tasks for copying files.
func (l *Loader) addAssetFileCopyTasks(assets []*assets.FileAsset, lifecycle *fi.Lifecycle) error {
	for _, asset := range assets {

		if asset.DownloadURL == nil {
			return fmt.Errorf("asset file url cannot be nil")
		}

		// test if the asset needs to be copied
		if asset.CanonicalURL != nil && asset.DownloadURL.String() != asset.CanonicalURL.String() {
			klog.V(10).Infof("processing asset: %q, %q", asset.DownloadURL.String(), asset.CanonicalURL.String())
			context := &fi.ModelBuilderContext{
				Tasks: l.tasks,
			}

			klog.V(10).Infof("adding task: %q", asset.DownloadURL.String())

			copyFileTask := &assettasks.CopyFile{
				Name:       fi.String(asset.CanonicalURL.String()),
				TargetFile: fi.String(asset.DownloadURL.String()),
				SourceFile: fi.String(asset.CanonicalURL.String()),
				SHA:        fi.String(asset.SHAValue),
				Lifecycle:  lifecycle,
			}

			context.AddTask(copyFileTask)
			l.tasks = context.Tasks

		}
	}

	return nil
}

func (l *Loader) processDeferrals() error {
	// TODO remove after proving it's not used
	for taskKey, task := range l.tasks {
		taskValue := reflect.ValueOf(task)

		err := reflectutils.ReflectRecursive(taskValue, func(path string, f *reflect.StructField, v reflect.Value) error {
			if reflectutils.IsPrimitiveValue(v) {
				return nil
			}

			if path == "" {
				// Don't process top-level value
				return nil
			}

			switch v.Kind() {
			case reflect.Interface, reflect.Ptr:
				if v.CanInterface() && !v.IsNil() {
					// TODO: Can we / should we use a type-switch statement
					intf := v.Interface()
					if hn, ok := intf.(fi.HasName); ok {
						name := hn.GetName()
						if name != nil {
							typeNameForTask := fi.TypeNameForTask(intf)
							primary := l.tasks[typeNameForTask+"/"+*name]
							if primary == nil {
								klog.Fatalf("task %q needed deferral resolution", typeNameForTask+"/"+*name)
								primary = l.tasks[*name]
							}
							if primary == nil {
								keys := sets.NewString()
								for k := range l.tasks {
									keys.Insert(k)
								}
								klog.Infof("Known tasks:")
								for _, k := range keys.List() {
									klog.Infof("  %s", k)
								}

								return fmt.Errorf("unable to find task %q, referenced from %s:%s", typeNameForTask+"/"+*name, taskKey, path)
							}

							klog.V(11).Infof("Replacing task %q at %s:%s", *name, taskKey, path)
							v.Set(reflect.ValueOf(primary))
						}
						return reflectutils.SkipReflection
					} else if rh, ok := intf.(*fi.ResourceHolder); ok {
						if rh.Resource == nil {
							klog.Fatalf("resource %s needed deferral resolution", rh.Name)
							//Resources can contain template 'arguments', separated by spaces
							// <resourcename> <arg1> <arg2>
							tokens := strings.Split(rh.Name, " ")
							match := tokens[0]
							args := tokens[1:]

							match = strings.TrimPrefix(match, "resources/")
							resource := l.Resources[match]

							if resource == nil {
								klog.Infof("Known resources:")
								for k := range l.Resources {
									klog.Infof("  %s", k)
								}
								return fmt.Errorf("unable to find resource %q, referenced from %s:%s", rh.Name, taskKey, path)
							}

							err := l.populateResource(rh, resource, args)
							if err != nil {
								return fmt.Errorf("error setting resource value: %v", err)
							}
						}
						return reflectutils.SkipReflection
					}
				}
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("unexpected error resolving task %q: %v", taskKey, err)
		}
	}

	return nil
}

func (l *Loader) populateResource(rh *fi.ResourceHolder, resource fi.Resource, args []string) error {
	if resource == nil {
		return nil
	}

	if len(args) != 0 {
		templateResource, ok := resource.(fi.TemplateResource)
		if !ok {
			return fmt.Errorf("cannot have arguments with resources of type %T", resource)
		}
		resource = templateResource.Curry(args)
	}
	rh.Resource = resource

	return nil
}
