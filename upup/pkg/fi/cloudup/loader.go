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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/assettasks"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/reflectutils"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	KEY_NAME = "name"
	KEY_TYPE = "_type"
)

type Loader struct {
	Cluster *api.Cluster

	WorkDir string

	ModelStore vfs.Path

	Tags              sets.String
	TemplateFunctions template.FuncMap

	typeMap map[string]reflect.Type

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
	l.typeMap = make(map[string]reflect.Type)
	l.Resources = make(map[string]fi.Resource)
	l.TemplateFunctions = make(template.FuncMap)
}

func (l *Loader) AddTypes(types map[string]interface{}) {
	for key, proto := range types {
		_, exists := l.typeMap[key]
		if exists {
			klog.Fatalf("duplicate type key: %q", key)
		}

		t := reflect.TypeOf(proto)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		l.typeMap[key] = t
	}
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
	return nil
}

func (l *Loader) BuildTasks(modelStore vfs.Path, models []string, assetBuilder *assets.AssetBuilder, lifecycle *fi.Lifecycle, lifecycleOverrides map[string]fi.Lifecycle) (map[string]fi.Task, error) {
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

	for _, model := range models {
		modelDir := modelStore.Join(model)
		err := tw.Walk(modelDir)
		if err != nil {
			return nil, err
		}
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
	err := l.processDeferrals()
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
	klog.V(8).Infof("Reading %s", i.Path)
	contents, err := i.ReadString()
	if err != nil {
		return err
	}

	data, err := l.executeTemplate(i.RelativePath, contents, nil)
	if err != nil {
		return err
	}

	objects, err := l.loadYamlObjects(i.RelativePath, data)
	if err != nil {
		return err
	}

	for k, v := range objects {
		_, found := l.tasks[k]
		if found {
			return fmt.Errorf("found duplicate object: %q", k)
		}
		l.tasks[k] = v.(fi.Task)
	}
	return nil
}

func (l *Loader) loadYamlObjects(key string, data string) (map[string]interface{}, error) {
	var o map[string]interface{}
	if strings.TrimSpace(data) != "" {
		err := utils.YamlUnmarshal([]byte(data), &o)
		if err != nil {
			// TODO: It would be nice if yaml returned us the line number here
			klog.Infof("error parsing yaml.  yaml follows:")
			for i, line := range strings.Split(string(data), "\n") {
				fmt.Fprintf(os.Stderr, "%3d: %s\n", i, line)
			}
			return nil, fmt.Errorf("error parsing yaml %q: %v", key, err)
		}
	}

	return l.loadObjectMap(key, o)
}

func (l *Loader) loadObjectMap(key string, data map[string]interface{}) (map[string]interface{}, error) {
	loaded := make(map[string]interface{})

	for k, v := range data {
		typeId := ""
		name := ""

		// If the name & type are not specified in the values,
		// we infer them from the key (first component -> typeid, last component -> name)
		if vMap, ok := v.(map[string]interface{}); ok {
			if s, ok := vMap[KEY_TYPE]; ok {
				typeId = s.(string)
			}
			if s, ok := vMap[KEY_NAME]; ok {
				name = s.(string)
			}
		}

		inferredName := false

		if name == "" {
			firstSlash := strings.Index(k, "/")
			name = k[firstSlash+1:]
			inferredName = true
		}

		if typeId == "" {
			firstSlash := strings.Index(k, "/")
			if firstSlash != -1 {
				typeId = k[:firstSlash]
			}

			if typeId == "" {
				return nil, fmt.Errorf("cannot determine type for %q", k)
			}
		}
		t, found := l.typeMap[typeId]
		if !found {
			return nil, fmt.Errorf("unknown type %q (in %q)", typeId, key)
		}

		o := reflect.New(t)

		// TODO replace with partial unmarshal...
		jsonValue, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("error marshaling to json: %v", err)
		}
		err = json.Unmarshal(jsonValue, o.Interface())
		if err != nil {
			klog.V(2).Infof("JSON was %q", string(jsonValue))
			return nil, fmt.Errorf("error parsing %q: %v", key, err)
		}
		klog.V(4).Infof("Built %s:%s => %v", key, k, o.Interface())

		if inferredName {
			hn, ok := o.Interface().(fi.HasName)
			if ok {
				hn.SetName(name)
			}
		}
		loaded[k] = o.Interface()
	}
	return loaded, nil
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
