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

type ApiGroup string

type ApiGroups []ApiGroup

func (a ApiGroup) String() string {
	return string(a)
}

func (a ApiGroups) Len() int      { return len(a) }
func (a ApiGroups) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ApiGroups) Less(i, j int) bool {
	// "apps" group APIs are newer than "extensions" group APIs
	if a[i].String() == "apps" && a[j].String() == "extensions" {
		return false
	}
	if a[j].String() == "apps" && a[i].String() == "extensions" {
		return true
	}
	return strings.Compare(a[i].String(), a[j].String()) < 0
}

type ApiVersions []ApiVersion

func (a ApiVersions) Len() int      { return len(a) }
func (a ApiVersions) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ApiVersions) Less(i, j int) bool {
	return a[i].LessThan(a[j])
}
