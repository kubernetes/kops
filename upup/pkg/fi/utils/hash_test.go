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

func TestHashString(t *testing.T) {
	tests := []struct {
		s           string
		expectedStr string
	}{
		{
			s:           "test",
			expectedStr: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			s:           "~!*/?",
			expectedStr: "2ebef337aebca07992300b61b58f3581a3d0228e9eb17c0340eaa01bcb2abde2",
		},
		{
			s:           "测试1",
			expectedStr: "d7d16e0c2747b4dbe70a6f4977cf4f4ebb4a227d7ebc24c3a0e99acaae79b518",
		},
		{
			s:           "-897668",
			expectedStr: "e6909a82db1ad9349f016624af382f62226e9a1bffbd360f9ccc4ef8aaded232",
		},
	}

	for _, test := range tests {
		result, _ := HashString(test.s)
		if test.expectedStr != result {
			t.Errorf("Expected %s, got %s", test.expectedStr, result)
		}
	}
}
