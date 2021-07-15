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

package templater

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"encoding/json"

	"k8s.io/kops/pkg/apis/kops"
	"sigs.k8s.io/yaml"
)

const (
	templateName = "mainTemplate"
)

// Templater is golang template renders
type Templater struct {
	channel *kops.Channel
}

// NewTemplater returns a new renderer implementation
func NewTemplater(channel *kops.Channel) *Templater {
	return &Templater{
		channel: channel,
	}
}

// Render is responsible for actually rendering the template
func (r *Templater) Render(content string, context map[string]interface{}, snippets map[string]string, failOnMissing bool) (rendered string, err error) {
	// @step: create the template
	tm := template.New(templateName)
	if _, err = tm.Funcs(r.templateFuncsMap(tm)).Parse(content); err != nil {
		return
	}
	if failOnMissing {
		tm.Option("missingkey=error")
	}

	// @step: add the snippits into the mix
	for filename, snippet := range snippets {
		if filename == templateName {
			return "", fmt.Errorf("snippet cannot have the same name as the template: %s", filename)
		}
		if _, err = tm.New(filename).Parse(snippet); err != nil {
			return rendered, fmt.Errorf("unable to parse snippet: %s, error: %s", filename, err)
		}
	}

	// @step: render the actual template
	writer := new(bytes.Buffer)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to parse template, error: %s", r)
		}
	}()
	if err = tm.ExecuteTemplate(writer, templateName, context); err != nil {
		return
	}

	return writer.String(), nil
}

// indentContent is responsible for indenting the string content
func indentContent(indent int, content string) string {
	var b bytes.Buffer
	length := len(strings.Split(content, "\n")) - 1
	for i, x := range strings.Split(content, "\n") {
		// @check if the length of the line is zero and set spacer
		spacer := indent
		if i == 0 || len(x) <= 0 {
			spacer = 0
		}
		// @step: write the line to the buffer
		line := fmt.Sprintf("%"+fmt.Sprintf("%d", spacer)+"s%s", "", x)
		b.WriteString(line)

		// @check if we need a newline
		if i < length {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// includeSnippet is responsible for including a snippet
func includeSnippet(tm *template.Template, name string, context map[string]interface{}) (string, error) {
	b := bytes.NewBufferString("")
	if err := tm.ExecuteTemplate(b, name, context); err != nil {
		return "", fmt.Errorf("snippet: %s, issue: %s", name, err)
	}

	return b.String(), nil
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// fromYAML converts a YAML document into a map[string]interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromYAML(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// fromYAMLArray converts a YAML array into a []interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string as
// the first and only item in the returned array.
func fromYAMLArray(str string) []interface{} {
	a := []interface{}{}

	if err := yaml.Unmarshal([]byte(str), &a); err != nil {
		a = []interface{}{err.Error()}
	}
	return a
}

// toJSON takes an interface, marshals it to json, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

// fromJSON converts a JSON document into a map[string]interface{}.
//
// This is not a general-purpose JSON parser, and will not parse all valid
// JSON documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromJSON(str string) map[string]interface{} {
	m := make(map[string]interface{})

	if err := json.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// fromJSONArray converts a JSON array into a []interface{}.
//
// This is not a general-purpose JSON parser, and will not parse all valid
// JSON documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string as
// the first and only item in the returned array.
func fromJSONArray(str string) []interface{} {
	a := []interface{}{}

	if err := json.Unmarshal([]byte(str), &a); err != nil {
		a = []interface{}{err.Error()}
	}
	return a
}
