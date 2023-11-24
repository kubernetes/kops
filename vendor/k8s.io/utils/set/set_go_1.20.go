//go:build !go1.21

/*
Copyright 2023 The Kubernetes Authors.

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

package set

// Clear empties the set.
// It is preferable to replace the set with a newly constructed set,
// but not all callers can do that (when there are other references to the map).
// In some cases the set *won't* be fully cleared, e.g. a Set[float32] containing NaN
// can't be cleared because NaN can't be removed.
// For sets containing items of a type that is reflexive for ==,
// this is optimized to a single call to runtime.mapclear().
func (s Set[T]) Clear() Set[T] {
	for key := range s {
		delete(s, key)
	}
	return s
}
