/*
Copyright 2026 The Kubernetes Authors.

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

package reflectutils

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

func TestValueAsString(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name:     "string set",
			value:    sets.New("udp:53", "tcp:443", "tcp:443"),
			expected: "[tcp:443, udp:53]",
		},
		{
			name:     "source range set",
			value:    sets.New("172.16.0.0/12", "10.0.0.0/8"),
			expected: "[10.0.0.0/8, 172.16.0.0/12]",
		},
		{
			name:     "empty string set",
			value:    sets.New[string](),
			expected: "[]",
		},
		{
			name:     "nil string set",
			value:    sets.Set[string](nil),
			expected: "<nil>",
		},
		{
			name:     "slice preserves order",
			value:    []string{"udp:53", "tcp:443"},
			expected: "[udp:53, tcp:443]",
		},
		{
			name:     "map preserves values",
			value:    map[string]string{"key": "value"},
			expected: "{key: value}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if actual := ValueAsString(reflect.ValueOf(tt.value)); actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
