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

package util

import "testing"

func TestStringSlicesEqual(t *testing.T) {
	cases := []struct {
		l, r     []string
		expected bool
	}{
		{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
			true,
		},
		{
			[]string{"a", "b", "c"},
			[]string{"x", "y", "z"},
			false,
		},
		{
			[]string{"a", "b"},
			[]string{"a", "b", "c"},
			false,
		},
		{
			[]string{"", "", ""},
			[]string{"", "", ""},
			true,
		},
	}

	for _, c := range cases {
		if actual := StringSlicesEqual(c.l, c.r); actual != c.expected {
			t.Errorf("StringSlicesEqual(%#v, %#v) expected %#v, but got %#v", c.l, c.r, c.expected, actual)
		}
	}
}
