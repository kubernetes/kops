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

package reflectutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"k8s.io/klog"
)

var SkipReflection = errors.New("skip this value")

type MethodNotFoundError struct {
	Name   string
	Target interface{}
}

func (e *MethodNotFoundError) Error() string {
	return fmt.Sprintf("method %s not found on %T", e.Name, e.Target)
}

func IsMethodNotFound(err error) bool {
	_, ok := err.(*MethodNotFoundError)
	return ok
}

// JsonMergeStruct merges src into dest
// It uses a JSON marshal & unmarshal, so only fields that are JSON-visible will be copied
func JsonMergeStruct(dest, src interface{}) {
	// Not the most efficient approach, but simple & relatively well defined
	j, err := json.Marshal(src)
	if err != nil {
		klog.Fatalf("error marshaling config: %v", err)
	}
	err = json.Unmarshal(j, dest)
	if err != nil {
		klog.Fatalf("error unmarshaling config: %v", err)
	}
}

// InvokeMethod calls the specified method by reflection
func InvokeMethod(target interface{}, name string, args ...interface{}) ([]reflect.Value, error) {
	v := reflect.ValueOf(target)

	method, found := v.Type().MethodByName(name)
	if !found {
		return nil, &MethodNotFoundError{
			Name:   name,
			Target: target,
		}
	}

	var argValues []reflect.Value
	for _, a := range args {
		argValues = append(argValues, reflect.ValueOf(a))
	}
	klog.V(12).Infof("Calling method %s on %T", method.Name, target)
	m := v.MethodByName(method.Name)
	rv := m.Call(argValues)
	return rv, nil
}

func BuildTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		return "*" + BuildTypeName(t.Elem())
	case reflect.Slice:
		return "[]" + BuildTypeName(t.Elem())
	case reflect.Struct, reflect.Interface:
		return t.Name()
	case reflect.String, reflect.Bool, reflect.Int64, reflect.Uint8:
		return t.Name()
	case reflect.Map:
		return "map[" + BuildTypeName(t.Key()) + "]" + BuildTypeName(t.Elem())
	default:
		klog.Errorf("cannot find type name for: %v, assuming %s", t, t.Name())
		return t.Name()
	}
}

type visitorFunc func(path string, field *reflect.StructField, v reflect.Value) error

// ReflectRecursive calls visitor with v and every recursive sub-value, skipping subtrees if SkipReflection is returned
func ReflectRecursive(v reflect.Value, visitor visitorFunc) error {
	return reflectRecursive("", v, visitor)
}

func reflectRecursive(path string, v reflect.Value, visitor visitorFunc) error {
	vType := v.Type()

	err := visitor(path, nil, v)
	if err != nil {
		if err == SkipReflection {
			return nil
		}
		return err
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			structField := vType.Field(i)
			if structField.PkgPath != "" {
				// Field not exported
				continue
			}

			f := v.Field(i)

			childPath := path + "." + structField.Name
			// TODO: I think we are double visiting here; we should instead pass down structField into reflectRecursive
			err := visitor(childPath, &structField, f)
			if err != nil && err != SkipReflection {
				return err
			}
			if err == nil {
				err = reflectRecursive(childPath, f, visitor)
				if err != nil {
					return err
				}
			}
		}

	case reflect.Map:
		keys := v.MapKeys()
		for _, key := range keys {
			mv := v.MapIndex(key)

			childPath := path + "[" + fmt.Sprintf("%s", key.Interface()) + "]"
			err := visitor(childPath, nil, mv)
			if err != nil && err != SkipReflection {
				return err
			}
			if err == nil {
				err = reflectRecursive(childPath, mv, visitor)
				if err != nil {
					return err
				}
			}
		}

	case reflect.Array, reflect.Slice:
		len := v.Len()
		for i := 0; i < len; i++ {
			av := v.Index(i)

			childPath := path + "[" + fmt.Sprintf("%d", i) + "]"
			err := visitor(childPath, nil, av)
			if err != nil && err != SkipReflection {
				return err
			}
			if err == nil {
				err = reflectRecursive(childPath, av, visitor)
				if err != nil {
					return err
				}
			}
		}

	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			e := v.Elem()
			err = reflectRecursive(path, e, visitor)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// IsPrimitiveValue returns true if passed a value of primitive type: int, bool, etc
// Note that string (like []byte) is not treated as a primitive type
func IsPrimitiveValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true

	// The less-obvious cases!
	case reflect.String, reflect.Slice, reflect.Array:
		return false

	case reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func, reflect.Map, reflect.Struct, reflect.UnsafePointer:
		return false

	default:
		klog.Fatalf("Unhandled kind: %v", v.Kind())
		return false
	}
}

// FormatValue returns a string representing the value
func FormatValue(value interface{}) string {
	// Based on code in k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/validation/field/errors.go
	valueType := reflect.TypeOf(value)
	if value == nil || valueType == nil {
		value = "null"
	} else if valueType.Kind() == reflect.Ptr {
		if reflectValue := reflect.ValueOf(value); reflectValue.IsNil() {
			value = "null"
		} else {
			value = reflectValue.Elem().Interface()
		}
	}
	switch t := value.(type) {
	case int64, int32, float64, float32, bool:
		// use simple printer for simple types
		return fmt.Sprintf("%v", value)
	case string:
		return fmt.Sprintf("%q", t)
	case fmt.Stringer:
		// anything that defines String() is better than raw struct
		return t.String()
	default:
		// fallback to raw struct
		// TODO: internal types have panic guards against json.Marshaling to prevent
		// accidental use of internal types in external serialized form.  For now, use
		// %#v, although it would be better to show a more expressive output in the future
		return fmt.Sprintf("%#v", value)
	}
}
