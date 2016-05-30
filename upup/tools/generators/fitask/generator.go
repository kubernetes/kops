package main

import (
	"io"
	"k8s.io/kube-deploy/upup/tools/generators/pkg/codegen"
	"text/template"
)

type FitaskGenerator struct {
	Package string
}

var _ codegen.Generator = &FitaskGenerator{}

const headerDef = `
package {{.Package}}

import (
	"encoding/json"

	"k8s.io/kube-deploy/upup/pkg/fi"
)
`

const perTypeDef = `
// {{.Name}}

// JSON marshalling boilerplate
type real{{.Name}} {{.Name}}

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

var _ fi.HasName = &{{.Name}}{}

func (e *{{.Name}}) GetName() *string {
	return e.Name
}

func (e *{{.Name}}) SetName(name string) {
	e.Name = &name
}

func (e *{{.Name}}) String() string {
	return fi.TaskAsString(e)
}
`

func (g *FitaskGenerator) Init(parser *codegen.GoParser) error {
	g.Package = parser.Package.Name

	return nil
}

func (g *FitaskGenerator) WriteHeader(w io.Writer) error {
	t := template.Must(template.New("Header").Parse(headerDef))

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
