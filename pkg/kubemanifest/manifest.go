/*
Copyright 2017 The Kubernetes Authors.

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

package kubemanifest

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/text"
)

// Object holds arbitrary untyped kubernetes objects; it is used when we don't have the type definitions for them
type Object struct {
	data map[string]interface{}
}

// ObjectList describes a list of objects, allowing us to add bulk-methods
type ObjectList []*Object

// LoadObjectsFrom parses multiple objects from a yaml file
func LoadObjectsFrom(contents []byte) (ObjectList, error) {
	var objects []*Object

	sections := text.SplitContentToSections(contents)

	for _, section := range sections {
		// We need this so we don't error on a section that is empty / commented out
		if !hasYAMLContent(section) {
			continue
		}

		data := make(map[string]interface{})
		err := yaml.Unmarshal(section, &data)
		if err != nil {
			return nil, fmt.Errorf("error parsing yaml: %v", err)
		}

		obj := &Object{
			//bytes: section,
			data: data,
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// hasYAMLContent determines if the byte slice has any content,
// because yaml parsing gives an error if called with no content.
// TODO: How does apimachinery avoid this problem?
func hasYAMLContent(yamlData []byte) bool {
	for _, line := range bytes.Split(yamlData, []byte("\n")) {
		l := bytes.TrimSpace(line)
		if len(l) != 0 && !bytes.HasPrefix(l, []byte("#")) {
			return true
		}
	}
	return false
}

// ToYAML serializes a list of objects back to bytes; it is the opposite of LoadObjectsFrom
func (l ObjectList) ToYAML() ([]byte, error) {
	var yamlSeparator = []byte("\n---\n\n")
	var yamls [][]byte
	for _, object := range l {
		// Don't serialize empty objects - they confuse yaml parsers
		if object.IsEmptyObject() {
			continue
		}

		y, err := object.ToYAML()
		if err != nil {
			return nil, fmt.Errorf("error re-marshaling manifest: %v", err)
		}

		yamls = append(yamls, y)
	}

	return bytes.Join(yamls, yamlSeparator), nil
}

func (m *Object) ToYAML() ([]byte, error) {
	b, err := yaml.Marshal(m.data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling manifest to yaml: %v", err)
	}
	return b, nil
}

func (m *Object) accept(visitor Visitor) error {
	err := visit(visitor, m.data, []string{}, func(v interface{}) {
		klog.Fatal("cannot mutate top-level data")
	})
	return err
}

// IsEmptyObject checks if the object has no keys set (i.e. `== {}`)
func (m *Object) IsEmptyObject() bool {
	return len(m.data) == 0
}

// Kind returns the kind field of the object, or "" if it cannot be found or is invalid
func (m *Object) Kind() string {
	v, found := m.data["kind"]
	if !found {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// APIVersion returns the apiVersion field of the object, or "" if it cannot be found or is invalid
func (m *Object) APIVersion() string {
	v, found := m.data["apiVersion"]
	if !found {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// Reparse parses a subfield from an object
func (m *Object) Reparse(obj interface{}, fields ...string) error {
	humanFields := strings.Join(fields, ".")

	current := m.data
	for _, field := range fields {
		v, found := current[field]
		if !found {
			return fmt.Errorf("field %q in %s not found", field, humanFields)
		}

		m, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field %q in %s was not an object, was %T", field, humanFields, v)
		}
		current = m
	}

	b, err := yaml.Marshal(current)
	if err != nil {
		return fmt.Errorf("error marshaling %s to yaml: %v", humanFields, err)
	}

	if err := yaml.Unmarshal(b, obj); err != nil {
		return fmt.Errorf("error unmarshaling subobject %s: %v", humanFields, err)
	}

	return nil
}

// Set mutates a subfield to the newValue
func (m *Object) Set(newValue interface{}, fieldPath ...string) error {
	humanFields := strings.Join(fieldPath, ".")

	current := m.data
	if len(fieldPath) >= 2 {
		for _, field := range fieldPath[:len(fieldPath)-1] {
			v, found := current[field]
			if !found {
				return fmt.Errorf("field %q in %s not found", field, humanFields)
			}

			m, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("field %q in %s was not an object, was %T", field, humanFields, v)
			}
			current = m
		}
	}

	current[fieldPath[len(fieldPath)-1]] = newValue

	return nil
}
