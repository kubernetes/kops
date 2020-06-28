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
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/assettasks"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/util/pkg/reflectutils"
	"k8s.io/kops/util/pkg/vfs"
)

type Loader struct {
	Cluster *kopsapi.Cluster

	Tags              sets.String
	TemplateFunctions template.FuncMap

	Resources map[string]fi.Resource

	Builders []fi.ModelBuilder

	tasks map[string]fi.Task
}

type templateResource struct {
	key      string
	loader   *Loader
	template string
	args     []string
}

var _ fi.Resource = &templateResource{}
var _ fi.TemplateResource = &templateResource{}

func (a *templateResource) Open() (io.Reader, error) {
	var err error
	result, err := a.loader.executeTemplate(a.key, a.template, a.args)
	if err != nil {
		return nil, fmt.Errorf("error executing resource template %q: %v", a.key, err)
	}
	reader := bytes.NewReader([]byte(result))
	return reader, nil
}

func (a *templateResource) Curry(args []string) fi.TemplateResource {
	curried := &templateResource{}
	*curried = *a
	curried.args = append(curried.args, args...)
	return curried
}

func (l *Loader) Init() {
	l.tasks = make(map[string]fi.Task)
	l.Resources = make(map[string]fi.Resource)
	l.TemplateFunctions = make(template.FuncMap)
}

func (l *Loader) executeTemplate(key string, d string, args []string) (string, error) {
	t := template.New(key)

	funcMap := make(template.FuncMap)
	funcMap["Args"] = func() []string {
		return args
	}
	funcMap["RenderResource"] = func(resourceName string, args []string) (string, error) {
		return l.renderResource(resourceName, args)
	}
	for k, fn := range l.TemplateFunctions {
		funcMap[k] = fn
	}
	t.Funcs(funcMap)

	t.Option("missingkey=zero")

	spec := l.Cluster.Spec

	_, err := t.Parse(d)
	if err != nil {
		return "", fmt.Errorf("error parsing template %q: %v", key, err)
	}

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, key, spec)
	if err != nil {
		return "", fmt.Errorf("error executing template %q: %v", key, err)
	}

	return buffer.String(), nil
}

func ignoreHandler(i *loader.TreeWalkItem) error {
	// TODO remove after proving it's dead code
	klog.Fatalf("ignoreHandler called on %s", i.Path)
	return fmt.Errorf("ignoreHandler not implemented")
}

func (l *Loader) BuildTasks(modelStore vfs.Path, assetBuilder *assets.AssetBuilder, lifecycle *fi.Lifecycle, lifecycleOverrides map[string]fi.Lifecycle) (map[string]fi.Task, error) {
	// Second pass: load everything else
	tw := &loader.TreeWalker{
		DefaultHandler: l.objectHandler,
		Contexts: map[string]loader.Handler{
			"resources": l.resourceHandler,
		},
		Extensions: map[string]loader.Handler{
			".options": ignoreHandler,
		},
		Tags: l.Tags,
	}

	modelDir := modelStore.Join("cloudup")
	err := tw.Walk(modelDir)
	if err != nil {
		return nil, err
	}

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
	err = l.processDeferrals()
	if err != nil {
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

func (l *Loader) resourceHandler(i *loader.TreeWalkItem) error {
	contents, err := i.ReadBytes()
	if err != nil {
		return err
	}

	var a fi.Resource
	key := i.RelativePath
	if strings.HasSuffix(key, ".template") {
		key = strings.TrimSuffix(key, ".template")
		klog.V(2).Infof("loading (templated) resource %q", key)

		a = &templateResource{
			template: string(contents),
			loader:   l,
			key:      key,
		}
	} else {
		klog.V(2).Infof("loading resource %q", key)
		a = fi.NewBytesResource(contents)

	}

	l.Resources[key] = a
	return nil
}

func (l *Loader) objectHandler(i *loader.TreeWalkItem) error {
	// TODO remove after proving it's dead code
	klog.Fatalf("objectHandler called on %s", i.Path)
	return fmt.Errorf("objectHandler not implemented")
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

func (l *Loader) renderResource(resourceName string, args []string) (string, error) {
	resourceKey := strings.TrimSuffix(resourceName, ".template")
	resourceKey = strings.TrimPrefix(resourceKey, "resources/")
	configResource := l.Resources[resourceKey]
	if configResource == nil {
		return "", fmt.Errorf("cannot find resource %q", resourceName)
	}

	if tr, ok := configResource.(fi.TemplateResource); ok {
		configResource = tr.Curry(args)
	} else if len(args) != 0 {
		return "", fmt.Errorf("args passed when building node config, but config was not a template %q", resourceName)
	}

	data, err := fi.ResourceAsBytes(configResource)
	if err != nil {
		return "", fmt.Errorf("error reading resource %q: %v", resourceName, err)
	}

	return string(data), nil
}
