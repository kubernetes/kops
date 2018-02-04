package imagebuilder

import (
	"bytes"
	"fmt"
	"text/template"
)

// ExpandTemplate executes a golang template
func ExpandTemplate(key string, templateString string, context interface{}) (string, error) {
	t := template.New(key)

	funcMap := make(template.FuncMap)

	t.Funcs(funcMap)

	_, err := t.Parse(templateString)
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
