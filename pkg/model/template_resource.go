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

package model

import (
	"bytes"
	"fmt"
	"io"
	"text/template"

	"k8s.io/kops/upup/pkg/fi"
)

type templateResource struct {
	key        string
	definition string
	context    interface{}

	parsed *template.Template
}

var _ fi.Resource = &templateResource{}

func NewTemplateResource(key string, definition string, functions template.FuncMap, context interface{}) (*templateResource, error) {
	r := &templateResource{
		key:        key,
		definition: definition,
		context:    context,
	}

	t := template.New(key)

	t.Funcs(functions)
	_, err := t.Parse(definition)
	if err != nil {
		return nil, fmt.Errorf("error parsing template %q: %v", key, err)
	}

	t.Option("missingkey=zero")

	r.parsed = t

	return r, nil

}
func (r *templateResource) Open() (io.Reader, error) {
	buffer := &bytes.Buffer{}

	err := r.parsed.ExecuteTemplate(buffer, r.key, r.context)
	if err != nil {
		return nil, fmt.Errorf("error executing template %q: %v", r.key, err)
	}

	return buffer, nil
}
