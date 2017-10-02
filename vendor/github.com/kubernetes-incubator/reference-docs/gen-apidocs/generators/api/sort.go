/*
Copyright 2016 The Kubernetes Authors.

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

package api

import "strings"

type SortDefinitionsByName []*Definition

func (a SortDefinitionsByName) Len() int      { return len(a) }
func (a SortDefinitionsByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortDefinitionsByName) Less(i, j int) bool {
	if a[i].Name == a[j].Name {
		if a[i].Version.String() == a[j].Version.String() {
			return a[i].Group.String() < a[j].Group.String()
		}
		return a[i].Version.LessThan(a[j].Version)
	}
	return a[i].Name < a[j].Name
}

type SortDefinitionsByVersion []*Definition

func (a SortDefinitionsByVersion) Len() int      { return len(a) }
func (a SortDefinitionsByVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortDefinitionsByVersion) Less(i, j int) bool {
	switch {
	case a[i].Version == a[j].Version:
		return strings.Compare(a[i].Group.String(), a[j].Group.String()) < 0
	default:
		return a[i].Version.LessThan(a[j].Version)
	}
}
