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

package stringorset

import (
	"encoding/json"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// StringOrSet is a type that holds a []string, but marshals to a []string or a string.
type StringOrSet struct {
	values             sets.Set[string]
	forceEncodeAsArray bool
}

func (s *StringOrSet) IsEmpty() bool {
	return len(s.values) == 0
}

// Set will build a value that marshals to a JSON array
func Set(v []string) StringOrSet {
	values := sets.Set[string]{}
	values.Insert(v...)
	return StringOrSet{values: values, forceEncodeAsArray: true}
}

// Of will build a value that marshals to a JSON array if len(v) > 1,
// otherwise to a single string
func Of(v ...string) StringOrSet {
	if v == nil {
		v = []string{}
	}
	values := sets.Set[string]{}
	values.Insert(v...)
	return StringOrSet{values: values}
}

// String will build a value that marshals to a single string
func String(v string) StringOrSet {
	return StringOrSet{values: sets.New(v), forceEncodeAsArray: false}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (s *StringOrSet) UnmarshalJSON(value []byte) error {
	if value[0] == '[' {
		s.forceEncodeAsArray = true
		var vals []string
		if err := json.Unmarshal(value, &vals); err != nil {
			return nil
		}
		s.values = sets.New(vals...)
		return nil
	}
	s.forceEncodeAsArray = false
	var stringValue string
	if err := json.Unmarshal(value, &stringValue); err != nil {
		return err
	}
	s.values = sets.New(stringValue)
	return nil
}

// String returns the string value, or the Itoa of the int value.
func (s StringOrSet) String() string {
	return strings.Join(sets.List[string](s.values), ",")
}

func (v *StringOrSet) Value() []string {
	vals := sets.List[string](v.values)
	sort.Strings(vals)
	return vals
}

func (l StringOrSet) Equal(r StringOrSet) bool {
	return l.values.Equal(r.values)
}

// MarshalJSON implements the json.Marshaller interface.
func (v StringOrSet) MarshalJSON() ([]byte, error) {
	encodeAsJSONArray := v.forceEncodeAsArray
	if len(v.values) > 1 {
		encodeAsJSONArray = true
	}
	values := v.Value()
	if values == nil {
		values = []string{}
	}
	if encodeAsJSONArray {
		return json.Marshal(values)
	} else if len(values) == 1 {
		return json.Marshal(&values[0])
	} else {
		return json.Marshal(values)
	}
}
