/*
Copyright 2021 The Kubernetes Authors.

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

package terraformWriter

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"
)

type TerraformWriter struct {
	// mutex protects the following items (resources & Files)
	mutex sync.Mutex
	// resources is a list of TF items that should be created
	resources []*terraformResource
	// outputs is a list of our TF output variables
	outputs map[string]*terraformOutputVariable
	// Files is a map of TF resource Files that should be created
	Files map[string][]byte
}

type OutputValue struct {
	Value      *Literal
	ValueArray []*Literal
}

type terraformResource struct {
	ResourceType string
	ResourceName string
	Item         interface{}
}

type terraformOutputVariable struct {
	Key        string
	Value      *Literal
	ValueArray []*Literal
}

// sanitizeName ensures terraform resource names don't start with digits or contain any invalid characters
func sanitizeName(name string) string {
	// Terraform resource names cannot start with a digit
	if _, err := strconv.Atoi(string(name[0])); err == nil {
		name = fmt.Sprintf("prefix_%v", name)
	}
	return strings.NewReplacer(".", "-", "/", "--", ":", "_").Replace(name)
}

func (t *TerraformWriter) InitTerraformWriter() {
	t.Files = make(map[string][]byte)
	t.outputs = make(map[string]*terraformOutputVariable)
}

func (t *TerraformWriter) AddFileBytes(resourceType string, resourceName string, key string, data []byte, base64 bool) (*Literal, error) {
	id := resourceType + "_" + resourceName + "_" + key

	t.mutex.Lock()
	defer t.mutex.Unlock()

	p := path.Join("data", id)
	t.Files[p] = data

	modulePath := fmt.Sprintf("%q", path.Join("${path.module}", p))
	fn := "file"
	if base64 {
		fn = "filebase64"
	}
	return LiteralFunctionExpression(fn, []string{modulePath}), nil
}

func (t *TerraformWriter) RenderResource(resourceType string, resourceName string, e interface{}) error {
	res := &terraformResource{
		ResourceType: resourceType,
		ResourceName: resourceName,
		Item:         e,
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.resources = append(t.resources, res)

	return nil
}

func (t *TerraformWriter) AddOutputVariable(key string, literal *Literal) error {
	v := &terraformOutputVariable{
		Key:   key,
		Value: literal,
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.outputs[key] != nil {
		return fmt.Errorf("duplicate variable: %q", key)
	}
	t.outputs[key] = v

	return nil
}

func (t *TerraformWriter) AddOutputVariableArray(key string, literal *Literal) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.outputs[key] == nil {
		v := &terraformOutputVariable{
			Key: key,
		}
		t.outputs[key] = v
	}
	if t.outputs[key].Value != nil {
		return fmt.Errorf("variable %q is both an array and a scalar", key)
	}

	t.outputs[key].ValueArray = append(t.outputs[key].ValueArray, literal)

	return nil
}

func (t *TerraformWriter) GetResourcesByType() (map[string]map[string]interface{}, error) {
	resourcesByType := make(map[string]map[string]interface{})

	for _, res := range t.resources {
		resources := resourcesByType[res.ResourceType]
		if resources == nil {
			resources = make(map[string]interface{})
			resourcesByType[res.ResourceType] = resources
		}

		tfName := sanitizeName(res.ResourceName)

		if resources[tfName] != nil {
			return nil, fmt.Errorf("duplicate resource found: %s.%s", res.ResourceType, tfName)
		}

		resources[tfName] = res.Item
	}

	return resourcesByType, nil
}

func (t *TerraformWriter) GetOutputs() (map[string]OutputValue, error) {
	values := map[string]OutputValue{}
	for _, v := range t.outputs {
		tfName := sanitizeName(v.Key)
		if _, found := values[tfName]; found {
			return nil, fmt.Errorf("duplicate variable found: %s", tfName)
		}
		deduped, err := dedupLiterals(v.ValueArray)
		if err != nil {
			return nil, err
		}
		values[tfName] = OutputValue{
			Value:      v.Value,
			ValueArray: deduped,
		}
	}
	return values, nil
}
