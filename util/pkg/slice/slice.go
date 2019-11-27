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

// Package slice provides utility methods for common operations on slices.
package slice

// GetUniqueStrings returns a slice of strings in the extra slice that are not
// present in the main slice
func GetUniqueStrings(main, extra []string) []string {
	unique := []string{}

	for _, item := range extra {
		found := false

		for _, s := range main {
			if item == s {
				found = true
			}
		}

		if !found {
			unique = append(unique, item)
		}
	}

	return unique
}

// Contains checks if a slice contains an element
func Contains(list []string, e string) bool {
	for _, x := range list {
		if x == e {
			return true
		}
	}

	return false
}
