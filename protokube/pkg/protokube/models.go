package protokube

import (
	"bytes"
	"fmt"
	"text/template"
)

// ExecuteTemplate renders the specified template with the model
func ExecuteTemplate(key string, templateDefinition string, model interface{}) ([]byte, error) {
	t := template.New(key)

	_, err := t.Parse(templateDefinition)
	if err != nil {
		return nil, fmt.Errorf("error parsing template %q: %v", key, err)
	}

	t.Option("missingkey=zero")

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, key, model)
	if err != nil {
		return nil, fmt.Errorf("error executing template %q: %v", key, err)
	}

	return buffer.Bytes(), nil
}
