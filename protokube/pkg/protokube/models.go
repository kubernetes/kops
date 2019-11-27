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
