/*
Copyright 2024 The Kubernetes Authors.

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

package predicates

// Predicate is a predicate function for a type.
type Predicate[T any] func(T) bool

// AllOf returns a predicate that is true if all of the given predicates are true.
func AllOf[T any](predicates ...Predicate[T]) Predicate[T] {
	return func(t T) bool {
		for _, predicate := range predicates {
			if !predicate(t) {
				return false
			}
		}
		return true
	}
}

// Filter returns a slice of elements that match the predicate.
func Filter[T any](slice []T, predicate Predicate[T]) []T {
	if predicate == nil {
		return slice
	}

	var result []T
	for _, element := range slice {
		if predicate(element) {
			result = append(result, element)
		}
	}
	return result
}
