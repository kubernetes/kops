/*
Copyright 2020 The Kubernetes Authors.

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
	"strconv"
	"strings"
)

func SetString(target interface{}, targetPath string, newValue string) error {
	targetValue := reflect.ValueOf(target)

	targetFieldPath, err := ParseFieldPath(targetPath)
	if err != nil {
		return fmt.Errorf("cannot parse field path %q: %w", targetPath, err)
	}

	visitor := func(path *FieldPath, field *reflect.StructField, v reflect.Value) error {
		if !targetFieldPath.HasPrefixMatch(path) {
			return nil
		}

		if targetFieldPath.Matches(path) {
			if !v.CanSet() {
				return fmt.Errorf("cannot set field %q (marked immutable)", path)
			}

			if err := setPrimitive(v, newValue); err != nil {
				return fmt.Errorf("cannot set field %q: %v", path, err)
			}

			return nil
		}

		// Partial match, check for nil struct and auto-populate
		if v.Kind() == reflect.Ptr && v.IsNil() {
			if !v.CanSet() {
				return fmt.Errorf("cannot set field %q (marked immutable)", path)
			}

			t := v.Type().String()

			var newV reflect.Value

			switch v.Type().Elem().Kind() {
			case reflect.Struct:
				newV = reflect.New(v.Type().Elem())

			default:
				return fmt.Errorf("unhandled type %v %q", v.Type().Elem().Kind(), t)
			}

			v.Set(newV)
			return nil

		}

		return nil
	}

	return ReflectRecursive(targetValue, visitor, &ReflectOptions{JSONNames: true})
}

func setPrimitive(v reflect.Value, newValue string) error {
	if !v.CanSet() {
		return fmt.Errorf("cannot set value")
	}

	if v.Type().Kind() == reflect.Slice {
		// Because this function generally sets values, we overwrite instead of appending.
		// Then to support multiple values, we split on commas.
		// We have no way to escape a comma currently; but in general we prefer having a slice in the schema,
		// rather than having values that need to be parsed, so we may not need it.
		tokens := strings.Split(newValue, ",")
		valueArray := reflect.MakeSlice(v.Type(), 0, len(tokens))
		for _, s := range tokens {
			valueItem := reflect.New(v.Type().Elem())
			if err := setPrimitive(valueItem.Elem(), s); err != nil {
				return err
			}
			valueArray = reflect.Append(valueArray, valueItem.Elem())
		}
		reflect.New(v.Type().Elem())
		v.Set(valueArray)
		return nil
	}

	if v.Type().Kind() == reflect.Ptr {
		val := reflect.New(v.Type().Elem())
		if err := setPrimitive(val.Elem(), newValue); err != nil {
			return err
		}
		v.Set(val)
		return nil
	}

	t := v.Type().String()

	var newV reflect.Value

	switch t {
	case "string":
		newV = reflect.ValueOf(newValue)

	case "bool":
		b, err := strconv.ParseBool(newValue)
		if err != nil {
			return fmt.Errorf("cannot interpret %q value as bool", newValue)
		}
		newV = reflect.ValueOf(b)

	case "int64", "int32", "int":
		v, err := strconv.Atoi(newValue)
		if err != nil {
			return fmt.Errorf("cannot interpret %q value as integer", newValue)
		}

		switch t {
		case "int":
			newV = reflect.ValueOf(v)
		case "int32":
			v32 := int32(v)
			newV = reflect.ValueOf(v32)
		case "int64":
			v64 := int64(v)
			newV = reflect.ValueOf(v64)
		default:
			panic("missing case in switch")
		}

	default:
		// This handles enums and other simple conversions
		newV = reflect.ValueOf(newValue)
		if newV.Type().ConvertibleTo(v.Type()) {
			newV = newV.Convert(v.Type())
		} else {
			return fmt.Errorf("unhandled type %q", t)
		}
	}

	v.Set(newV)
	return nil
}
