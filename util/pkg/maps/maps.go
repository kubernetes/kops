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
	"cmp"
	"slices"

	"golang.org/x/exp/maps"
)

// Keys returns the keys of a map
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	return maps.Keys(m)
}

// SortedKeys returns a list of sorted keys
func SortedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	list := maps.Keys(m)
	slices.Sort(list)

	return list
}
