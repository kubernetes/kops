/*
Copyright 2020 The Kubernetes Authors.

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

package utils

import (
	"testing"
)

func TestStringSlicesEqual(t *testing.T) {
	tests := []struct {
		l        []string
		r        []string
		expected bool
	}{
		{
			l:        []string{"a", "b"},
			r:        []string{"a"},
			expected: false,
		},
		{
			l:        []string{"a", "b"},
			r:        []string{"a", "c"},
			expected: false,
		},
		{
			l:        []string{"a", "b"},
			r:        []string{"a", "b"},
			expected: true,
		},
	}

	for _, test := range tests {
		result := StringSlicesEqual(test.l, test.r)
		if test.expected != result {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}
