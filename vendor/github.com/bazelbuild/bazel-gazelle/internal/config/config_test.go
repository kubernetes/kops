/* Copyright 2017 The Bazel Authors. All rights reserved.

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

package config

import "testing"

func TestPreprocessTags(t *testing.T) {
	c := &Config{
		GenericTags: map[string]bool{"a": true, "b": true},
	}
	c.PreprocessTags()
	expectedTags := []string{"a", "b", "gc"}
	for _, tag := range expectedTags {
		if !c.GenericTags[tag] {
			t.Errorf("tag %q not set", tag)
		}
	}
	unexpectedTags := []string{"x", "cgo", "go1.8", "go1.7"}
	for _, tag := range unexpectedTags {
		if c.GenericTags[tag] {
			t.Errorf("tag %q unexpectedly set", tag)
		}
	}
}
