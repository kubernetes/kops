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

package main

import (
	"io"
	"text/template"

	"k8s.io/kops/upup/tools/generators/pkg/codegen"
)

type FitaskGenerator struct {
	Package string
}

var _ codegen.Generator = &FitaskGenerator{}

const fileHeaderDef = `/*
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

`

var preambleDef = `
package {{.Package}}

import (
	"encoding/json"

	"k8s.io/kops/upup/pkg/fi"
)
`

const perTypeDef = `
// {{.Name}}

// JSON marshaling boilerplate
type real{{.Name}} {{.Name}}

// UnmarshalJSON implements conversion to JSON, supporting an alternate specification of the object as a string
func (o *{{.Name}}) UnmarshalJSON(data []byte) error {
	var jsonName string
	if err := json.Unmarshal(data, &jsonName); err == nil {
		o.Name = &jsonName
		return nil
	}

	var r real{{.Name}}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	*o = {{.Name}}(r)
	return nil
}

var _ fi.HasLifecycle = &{{.Name}}{}

// GetLifecycle returns the Lifecycle of the object, implementing fi.HasLifecycle
func (o *{{.Name}}) GetLifecycle() *fi.Lifecycle {
	return o.Lifecycle
}

// SetLifecycle sets the Lifecycle of the object, implementing fi.SetLifecycle
func (o *{{.Name}}) SetLifecycle(lifecycle fi.Lifecycle) {
	o.Lifecycle = &lifecycle
}

var _ fi.HasName = &{{.Name}}{}

// GetName returns the Name of the object, implementing fi.HasName
func (o *{{.Name}}) GetName() *string {
	return o.Name
}

// SetName sets the Name of the object, implementing fi.SetName
func (o *{{.Name}}) SetName(name string) {
	o.Name = &name
}

// String is the stringer function for the task, producing readable output using fi.TaskAsString
func (o *{{.Name}}) String() string {
	return fi.TaskAsString(o)
}
`

func (g *FitaskGenerator) Init(parser *codegen.GoParser) error {
	g.Package = parser.Package.Name

	return nil
}

func (g *FitaskGenerator) WriteFileHeader(w io.Writer) error {
	t := template.Must(template.New("FileHeader").Parse(fileHeaderDef))

	return t.Execute(w, g)
}

func (g *FitaskGenerator) WritePreamble(w io.Writer) error {
	t := template.Must(template.New("Preamble").Parse(preambleDef))

	return t.Execute(w, g)
}

type TypeData struct {
	Name string
}

func (g *FitaskGenerator) WriteType(w io.Writer, typeName string) error {
	t := template.Must(template.New("PerType").Parse(perTypeDef))

	d := &TypeData{}
	d.Name = typeName

	return t.Execute(w, d)
}
