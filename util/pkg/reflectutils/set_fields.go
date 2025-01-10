/*
Copyright 2025 The Kubernetes Authors.

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

package reflectutils

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// FindSetFields returns the set of fields that are set in the struct,
// using the json tags as the field names.
// It only considers the fields that are listed in the fields argument.
func FindSetFields[T any](v *T, fields ...string) (sets.Set[string], error) {
	val := reflect.ValueOf(v).Elem()
	valType := val.Type()

	fieldsByJsonName := make(map[string]reflect.StructField)

	for i := 0; i < val.NumField(); i++ {
		fd := valType.Field(i)
		jsonName := fd.Tag.Get("json")
		if jsonName == "" || jsonName == "-" {
			continue
		}
		jsonName = strings.TrimSuffix(jsonName, ",omitempty")
		fieldsByJsonName[jsonName] = fd
	}

	setFields := sets.New[string]()
	for _, field := range fields {
		fd, ok := fieldsByJsonName[field]
		if !ok {
			return nil, fmt.Errorf("field %s is not known", field)
		}

		fieldVal := val.FieldByIndex(fd.Index)
		switch fieldVal.Kind() {
		case reflect.Ptr:
			if !fieldVal.IsNil() {
				setFields.Insert(field)
			}
		default:
			return nil, fmt.Errorf("field %s is not a pointer", fd.Name)
		}

	}

	return setFields, nil
}
