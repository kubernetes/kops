package model

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/kops/upup/pkg/fi"
	"text/template"
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
