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

package jsonutils

import (
	"reflect"
	"testing"
)

func TestSortSlice(t *testing.T) {
	testCases := []struct {
		name      string
		input     []any
		expected  []any
		expectErr bool
	}{
		{
			name:     "slice of strings",
			input:    []any{"c", "a", "b"},
			expected: []any{"a", "b", "c"},
		},
		{
			name:     "slice of numbers",
			input:    []any{3.0, 1.0, 2.0},
			expected: []any{1.0, 2.0, 3.0},
		},
		{
			name: "slice of objects",
			input: []any{
				map[string]any{"id": 2},
				map[string]any{"id": 1},
			},
			expected: []any{
				map[string]any{"id": 1},
				map[string]any{"id": 2},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sorted, err := SortSlice(tc.input)

			if (err != nil) != tc.expectErr {
				t.Fatalf("unexpected error state: got err=%v, want err=%v", err, tc.expectErr)
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(sorted, tc.expected) {
				t.Errorf("unexpected result: got\n%v\nwant\n%v", sorted, tc.expected)
			}
		})
	}
}
