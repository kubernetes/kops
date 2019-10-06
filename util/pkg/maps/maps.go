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

package maps

import (
	"reflect"
	"sort"
)

// Keys returns the keys of a map
func Keys(m interface{}) []string {
	var list []string

	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Map {
		for _, x := range v.MapKeys() {
			list = append(list, x.String())
		}
	}

	return list
}

// SortedKeys returns a list of sorted keys
func SortedKeys(m interface{}) []string {
	list := Keys(m)
	sort.Strings(list)

	return list
}
