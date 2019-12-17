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

	"github.com/Masterminds/sprig"
)

const (
	templateName = "mainTemplate"
)

// Templater is golang template renders
type Templater struct{}

// NewTemplater returns a new renderer implementation
func NewTemplater() *Templater {
	return &Templater{}
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

// templateFuncsMap returns a map if the template functions for this template
func (r *Templater) templateFuncsMap(tm *template.Template) template.FuncMap {
	// grab the template functions from sprig which are pretty awesome
	funcs := sprig.TxtFuncMap()

	funcs["indent"] = indentContent
	// @step: as far as i can see there's no native way in sprig in include external snippets of code
	funcs["include"] = func(name string, context map[string]interface{}) string {
		content, err := includeSnippet(tm, name, context)
		if err != nil {
			panic(err.Error())
		}

		return content
	}

	return funcs
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
