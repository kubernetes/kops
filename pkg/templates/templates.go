/*
Copyright 2017 The Kubernetes Authors.

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

package templates

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

type Templates struct {
	cluster           *kops.Cluster
	resources         map[string]fi.Resource
	TemplateFunctions template.FuncMap
}

func LoadTemplates(cluster *kops.Cluster, base vfs.Path) (*Templates, error) {
	t := &Templates{
		cluster:           cluster,
		resources:         make(map[string]fi.Resource),
		TemplateFunctions: make(template.FuncMap),
	}
	err := t.loadFrom(base)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Templates) Find(key string) fi.Resource {
	return t.resources[key]
}

func (t *Templates) loadFrom(base vfs.Path) error {
	files, err := base.ReadTree()
	if err != nil {
		return fmt.Errorf("error reading from %s", base)
	}

	for _, f := range files {
		contents, err := f.ReadFile()
		if err != nil {
			if os.IsNotExist(err) {
				// This is just an annoyance of gobindata - we can't tell the difference between files & directories.  Ignore.
				continue
			}
			return fmt.Errorf("error reading %s: %v", f, err)
		}

		key, err := vfs.RelativePath(base, f)
		if err != nil {
			return fmt.Errorf("error getting relative path for %s", f)
		}

		var resource fi.Resource
		if strings.HasSuffix(key, ".template") {
			key = strings.TrimSuffix(key, ".template")
			klog.V(6).Infof("loading (templated) resource %q", key)

			resource = &templateResource{
				template: string(contents),
				loader:   t,
				key:      key,
			}
		} else {
			klog.V(6).Infof("loading resource %q", key)
			resource = fi.NewBytesResource(contents)

		}

		t.resources[key] = resource
	}
	return nil
}

func (l *Templates) executeTemplate(key string, d string) (string, error) {
	t := template.New(key)

	funcMap := make(template.FuncMap)
	//funcMap["Args"] = func() []string {
	//	return args
	//}
	//funcMap["RenderResource"] = func(resourceName string, args []string) (string, error) {
	//	return l.renderResource(resourceName, args)
	//}
	for k, fn := range l.TemplateFunctions {
		funcMap[k] = fn
	}
	t.Funcs(funcMap)

	t.Option("missingkey=zero")

	spec := l.cluster.Spec

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

type templateResource struct {
	key      string
	loader   *Templates
	template string
}

var _ fi.Resource = &templateResource{}

func (a *templateResource) Open() (io.Reader, error) {
	var err error
	result, err := a.loader.executeTemplate(a.key, a.template)
	if err != nil {
		return nil, fmt.Errorf("error executing resource template %q: %v", a.key, err)
	}
	reader := bytes.NewReader([]byte(result))
	return reader, nil
}
