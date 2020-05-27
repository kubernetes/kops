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

package nodeup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/vfs"

	"k8s.io/klog"
)

type Loader struct {
	Builders []fi.ModelBuilder

	config  *nodeup.Config
	cluster *api.Cluster

	assets *fi.AssetStore
	tasks  map[string]fi.Task

	tags              sets.String
	TemplateFunctions template.FuncMap
}

func NewLoader(config *nodeup.Config, cluster *api.Cluster, assets *fi.AssetStore, tags sets.String) *Loader {
	l := &Loader{}
	l.assets = assets
	l.tasks = make(map[string]fi.Task)
	l.config = config
	l.cluster = cluster
	l.TemplateFunctions = make(template.FuncMap)
	l.tags = tags

	return l
}

func (l *Loader) executeTemplate(key string, d string) (string, error) {
	t := template.New(key)

	funcMap := make(template.FuncMap)
	for k, fn := range l.TemplateFunctions {
		funcMap[k] = fn
	}
	t.Funcs(funcMap)

	context := l.cluster.Spec

	_, err := t.Parse(d)
	if err != nil {
		return "", fmt.Errorf("error parsing template %q: %v", key, err)
	}

	t.Option("missingkey=zero")

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, key, context)
	if err != nil {
		return "", fmt.Errorf("error executing template %q: %v", key, err)
	}

	return buffer.String(), nil
}

func ignoreHandler(i *loader.TreeWalkItem) error {
	return nil
}

// Build is responsible for running the build tasks for nodeup
func (l *Loader) Build(baseDir vfs.Path) (map[string]fi.Task, error) {
	// First pass: load options
	tw := &loader.TreeWalker{
		DefaultHandler: ignoreHandler,
		Contexts: map[string]loader.Handler{
			"files":    ignoreHandler,
			"packages": ignoreHandler,
		},
		Tags: l.tags,
	}

	err := tw.Walk(baseDir)
	if err != nil {
		return nil, err
	}

	// Second pass: load everything else
	tw = &loader.TreeWalker{
		DefaultHandler: l.handleFile,
		Contexts: map[string]loader.Handler{
			"files":    l.handleFile,
			"packages": l.newTaskHandler("package/", nodetasks.NewPackage),
		},
		Tags: l.tags,
	}

	err = tw.Walk(baseDir)
	if err != nil {
		return nil, err
	}

	for _, builder := range l.Builders {
		context := &fi.ModelBuilderContext{
			Tasks: l.tasks,
		}
		err := builder.Build(context)
		if err != nil {
			return nil, err
		}
		l.tasks = context.Tasks
	}

	// If there is a package task, we need an update packages task
	for _, t := range l.tasks {
		if _, ok := t.(*nodetasks.Package); ok {
			klog.Infof("Package task found; adding UpdatePackages task")
			l.tasks["UpdatePackages"] = nodetasks.NewUpdatePackages()
			break
		}
	}
	if l.tasks["UpdatePackages"] == nil {
		klog.Infof("No package task found; won't update packages")
	}

	return l.tasks, nil
}

type TaskBuilder func(name string, contents string, meta string) (fi.Task, error)

func (l *Loader) newTaskHandler(prefix string, builder TaskBuilder) loader.Handler {
	return func(i *loader.TreeWalkItem) error {
		contents, err := i.ReadString()
		if err != nil {
			return err
		}
		name := i.Name
		if strings.HasSuffix(name, ".template") {
			name = strings.TrimSuffix(name, ".template")
			expanded, err := l.executeTemplate(name, contents)
			if err != nil {
				return fmt.Errorf("error executing template %q: %v", i.RelativePath, err)
			}

			contents = expanded
		}

		task, err := builder(name, contents, i.Meta)
		if err != nil {
			return fmt.Errorf("error building %s for %q: %v", i.Name, i.Path, err)
		}
		key := prefix + i.RelativePath

		if task != nil {
			l.tasks[key] = task
		}
		return nil
	}
}

func (l *Loader) handleFile(i *loader.TreeWalkItem) error {
	var task *nodetasks.File
	defaultFileType := nodetasks.FileType_File

	if strings.HasSuffix(i.RelativePath, ".template") {
		contents, err := i.ReadString()
		if err != nil {
			return err
		}

		// TODO: Use template resource here to defer execution?
		destPath := "/" + strings.TrimSuffix(i.RelativePath, ".template")
		name := strings.TrimSuffix(i.Name, ".template")
		expanded, err := l.executeTemplate(name, contents)
		if err != nil {
			return fmt.Errorf("error executing template %q: %v", i.RelativePath, err)
		}

		task, err = nodetasks.NewFileTask(name, fi.NewStringResource(expanded), destPath, i.Meta)
		if err != nil {
			return fmt.Errorf("error building task %q: %v", i.RelativePath, err)
		}
	} else if strings.HasSuffix(i.RelativePath, ".asset") {
		contents, err := i.ReadBytes()
		if err != nil {
			return err
		}

		destPath := "/" + strings.TrimSuffix(i.RelativePath, ".asset")
		name := strings.TrimSuffix(i.Name, ".asset")

		def := &nodetasks.AssetDefinition{}
		err = json.Unmarshal(contents, def)
		if err != nil {
			return fmt.Errorf("error parsing json for asset %q: %v", name, err)
		}

		asset, err := l.assets.Find(name, def.AssetPath)
		if err != nil {
			return fmt.Errorf("error trying to locate asset %q: %v", name, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", name)
		}

		task, err = nodetasks.NewFileTask(i.Name, asset, destPath, i.Meta)
		if err != nil {
			return fmt.Errorf("error building task %q: %v", i.RelativePath, err)
		}
	} else {
		var err error
		var contents fi.Resource
		if vfs.IsDirectory(i.Path) {
			defaultFileType = nodetasks.FileType_Directory
		} else {
			contents = fi.NewVFSResource(i.Path)
		}
		task, err = nodetasks.NewFileTask(i.Name, contents, "/"+i.RelativePath, i.Meta)
		if err != nil {
			return fmt.Errorf("error building task %q: %v", i.RelativePath, err)
		}
	}

	if task.Type == "" {
		task.Type = defaultFileType
	}

	klog.V(2).Infof("path %q -> task %v", i.Path, task)

	if task != nil {
		key := "file/" + i.RelativePath
		l.tasks[key] = task
	}

	return nil
}
