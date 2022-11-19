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

package model

import (
	"fmt"
	"sort"

	"k8s.io/kops/upup/pkg/fi"
)

// s is a helper that builds a *string from a string value
func s(v string) *string {
	return fi.PtrTo(v)
}

// b returns a pointer to a boolean
func b(v bool) *bool {
	return fi.PtrTo(v)
}

// buildContainerRuntimeEnvironmentVars just converts a series of keypairs to docker environment variables switches
func buildContainerRuntimeEnvironmentVars(env map[string]string) []string {
	var list []string
	for k, v := range env {
		list = append(list, []string{"--env", fmt.Sprintf("%s=%s", k, v)}...)
	}

	return list
}

// sortedStrings is just a one liner helper methods
func sortedStrings(list []string) []string {
	sort.Strings(list)

	return list
}
