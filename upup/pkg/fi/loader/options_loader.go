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

package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/reflectutils"
)

const maxIterations = 10

type OptionsTemplate struct {
	Name     string
	Tags     []string
	Template *template.Template
}

type OptionsLoader struct {
	templates OptionsTemplateList

	TemplateFunctions template.FuncMap

	Builders []OptionsBuilder
}

type OptionsBuilder interface {
	BuildOptions(options interface{}) error
}

type OptionsTemplateList []*OptionsTemplate

func (a OptionsTemplateList) Len() int {
	return len(a)
}
func (a OptionsTemplateList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a OptionsTemplateList) Less(i, j int) bool {
	l := a[i]
	r := a[j]

	// First ordering criteria: Execute things with fewer tags first (more generic)
	if len(l.Tags) != len(r.Tags) {
		return len(l.Tags) < len(r.Tags)
	}

	// TODO: lexicographic sort on tags, for full determinism?

	// Final ordering criteria: order by name
	return l.Name < r.Name
}

func NewOptionsLoader(templateFunctions template.FuncMap, builders []OptionsBuilder) *OptionsLoader {
	l := &OptionsLoader{}
	l.TemplateFunctions = make(template.FuncMap)
	for k, v := range templateFunctions {
		l.TemplateFunctions[k] = v
	}
	l.Builders = builders
	return l
}

func (l *OptionsLoader) AddTemplate(t *OptionsTemplate) {
	l.templates = append(l.templates, t)
}

// iterate performs a single iteration of all the templates, executing each template in order
func (l *OptionsLoader) iterate(userConfig interface{}, current interface{}) (interface{}, error) {
	sort.Sort(l.templates)

	t := reflect.TypeOf(current).Elem()

	next := reflect.New(t).Interface()

	// Copy the current state before applying rules; they act as defaults
	reflectutils.JsonMergeStruct(next, current)

	for _, t := range l.templates {
		klog.V(2).Infof("executing template %s (tags=%s)", t.Name, t.Tags)

		var buffer bytes.Buffer
		err := t.Template.ExecuteTemplate(&buffer, t.Name, current)
		if err != nil {
			return nil, fmt.Errorf("error executing template %q: %v", t.Name, err)
		}

		yamlBytes := buffer.Bytes()

		jsonBytes, err := utils.YAMLToJSON(yamlBytes)
		if err != nil {
			// TODO: It would be nice if yaml returned us the line number here
			klog.Infof("error parsing yaml.  yaml follows:")
			for i, line := range strings.Split(string(yamlBytes), "\n") {
				fmt.Fprintf(os.Stderr, "%3d: %s\n", i, line)
			}
			return nil, fmt.Errorf("error parsing yaml %q: %v", t.Name, err)
		}

		err = json.Unmarshal(jsonBytes, next)
		if err != nil {
			return nil, fmt.Errorf("error parsing yaml (converted to JSON) %q: %v", t.Name, err)
		}
	}

	for _, t := range l.Builders {
		klog.V(2).Infof("executing builder %T", t)

		err := t.BuildOptions(next)
		if err != nil {
			return nil, err
		}
	}

	// Also copy the user-provided values after applying rules; they act as overrides now
	reflectutils.JsonMergeStruct(next, userConfig)

	return next, nil
}

// Build executes the options configuration templates, until they converge
// It bails out after maxIterations
func (l *OptionsLoader) Build(userConfig interface{}) (interface{}, error) {
	options := userConfig
	iteration := 0
	for {
		nextOptions, err := l.iterate(userConfig, options)
		if err != nil {
			return nil, err
		}

		if reflect.DeepEqual(options, nextOptions) {
			return options, nil
		}

		iteration++
		if iteration > maxIterations {
			return nil, fmt.Errorf("options did not converge after %d iterations", maxIterations)
		}

		options = nextOptions
	}
}

// HandleOptions is the file handler for options files
// It builds a template with the file, and adds it to the list of options templates
func (l *OptionsLoader) HandleOptions(i *TreeWalkItem) error {
	contents, err := i.ReadString()
	if err != nil {
		return err
	}

	t := template.New(i.RelativePath)
	t.Funcs(l.TemplateFunctions)

	_, err = t.Parse(contents)
	if err != nil {
		return fmt.Errorf("error parsing options template %q: %v", i.Path, err)
	}

	t.Option("missingkey=zero")

	l.AddTemplate(&OptionsTemplate{
		Name:     i.RelativePath,
		Tags:     i.Tags,
		Template: t,
	})
	return nil
}
