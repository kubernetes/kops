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

package fi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// PtrTo returns a pointer to a copy of any value.
func PtrTo[T any](v T) *T {
	return &v
}

// ValueOf returns the value of a pointer or its zero value
func ValueOf[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}

func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringSliceValue takes a slice of string pointers and returns a slice of strings
func StringSliceValue(stringSlice []*string) []string {
	var newSlice []string
	for _, value := range stringSlice {
		if value != nil {
			newSlice = append(newSlice, *value)
		}
	}
	return newSlice
}

func IsNilOrEmpty(s *string) bool {
	if s == nil {
		return true
	}
	return *s == ""
}

// String is a helper that builds a *string from a string value
// This is similar to aws.String, except that we use it for non-AWS values
func String(s string) *string {
	return &s
}

// StringSlice is a helper that builds a []*string from a slice of strings
func StringSlice(stringSlice []string) []*string {
	var newSlice []*string
	for i := range stringSlice {
		newSlice = append(newSlice, &stringSlice[i])
	}
	return newSlice
}

// Float32 returns a point to a float32
func Float32(v float32) *float32 {
	return &v
}

// Float32Value returns the value of the float
func Float32Value(v *float32) float32 {
	if v == nil {
		return 0.0
	}

	return *v
}

// Float64 returns a point to a float64
func Float64(v float64) *float64 {
	return &v
}

// Float64Value returns the value of the float
func Float64Value(v *float64) float64 {
	if v == nil {
		return 0.0
	}

	return *v
}

// Bool returns a pointer to a bool
func Bool(v bool) *bool {
	return &v
}

// BoolValue returns the value of bool pointer or false
func BoolValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func Int32(v int32) *int32 {
	return &v
}

func Int32Value(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

// Int64 is a helper that builds a *int64 from an int64 value
// This is similar to aws.Int64, except that we use it for non-AWS values
func Int64(v int64) *int64 {
	return &v
}

func Int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func Int(v int) *int {
	return &v
}

func IntValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func Uint64Value(v *uint64) uint64 {
	if v == nil {
		return 0
	}
	return *v
}

// ArrayContains is checking does array contain single word
func ArrayContains(array []string, word string) bool {
	for _, item := range array {
		if item == word {
			return true
		}
	}
	return false
}

func DebugPrint(o interface{}) string {
	if o == nil {
		return "<nil>"
	}
	if resource, ok := o.(Resource); ok {
		if resource == nil {
			// Avoid go nil vs interface problems
			return "<nil>"
		}

		s, err := ResourceAsString(resource)
		if err != nil {
			return fmt.Sprintf("error converting resource to string: %v", err)
		}
		if len(s) >= 256 {
			s = s[:256] + "... (truncated)"
		}
		return s
	}

	v := reflect.ValueOf(o)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "<nil>"
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return "<?>"
	}
	o = v.Interface()
	if stringer, ok := o.(fmt.Stringer); ok {
		if stringer == nil {
			// Avoid go nil vs interface problems
			return "<nil>"
		}
		return stringer.String()
	}

	return fmt.Sprint(o)
}

func DebugAsJsonString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error marshaling: %v", err)
	}
	return string(data)
}

func DebugAsJsonStringIndent(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshaling: %v", err)
	}
	return string(data)
}

func ToInt64(s *string) *int64 {
	if s == nil {
		return nil
	}
	v, err := strconv.ParseInt(*s, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}

func ToString(v *int64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatInt(*v, 10)
	return &s
}
