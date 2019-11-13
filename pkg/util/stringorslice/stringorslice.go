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

package stringorslice

import (
	"encoding/json"
	"strings"
)

// StringOrSlice is a type that holds a []string, but marshals to a []string or a string.
type StringOrSlice struct {
	values             []string
	forceEncodeAsArray bool
}

// Slice will build a value that marshals to a JSON array
func Slice(v []string) StringOrSlice {
	return StringOrSlice{values: v, forceEncodeAsArray: true}
}

// Of will build a value that marshals to a JSON array if len(v) > 1,
// otherwise to a single string
func Of(v ...string) StringOrSlice {
	if v == nil {
		v = []string{}
	}
	return StringOrSlice{values: v}
}

// String will build a value that marshals to a single string
func String(v string) StringOrSlice {
	return StringOrSlice{values: []string{v}, forceEncodeAsArray: false}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (s *StringOrSlice) UnmarshalJSON(value []byte) error {
	if value[0] == '[' {
		s.forceEncodeAsArray = true
		if err := json.Unmarshal(value, &s.values); err != nil {
			return nil
		}
		return nil
	}
	s.forceEncodeAsArray = false
	var stringValue string
	if err := json.Unmarshal(value, &stringValue); err != nil {
		return err
	}
	s.values = []string{stringValue}
	return nil
}

// String returns the string value, or the Itoa of the int value.
func (s StringOrSlice) String() string {
	return strings.Join(s.values, ",")
}

func (v *StringOrSlice) Value() []string {
	return v.values
}

func (l StringOrSlice) Equal(r StringOrSlice) bool {
	if len(l.values) != len(r.values) {
		return false
	}
	for i := 0; i < len(l.values); i++ {
		if l.values[i] != r.values[i] {
			return false
		}
	}
	return true
}

// MarshalJSON implements the json.Marshaller interface.
func (v StringOrSlice) MarshalJSON() ([]byte, error) {
	encodeAsJsonArray := v.forceEncodeAsArray
	if len(v.values) > 1 {
		encodeAsJsonArray = true
	}
	values := v.values
	if values == nil {
		values = []string{}
	}
	if encodeAsJsonArray {
		return json.Marshal(values)
	} else if len(v.values) == 1 {
		s := v.values[0]
		return json.Marshal(&s)
	} else {
		return json.Marshal(values)
	}
}
