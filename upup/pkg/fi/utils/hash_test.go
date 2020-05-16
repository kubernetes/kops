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
			expectedStr: "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3",
		},
		{
			s:           "~!*/?",
			expectedStr: "783c2ea26549ce90df7cd90a27bf9094c226c406",
		},
		{
			s:           "测试1",
			expectedStr: "6d972f7f1450aba1f7496f685c3c656c4fca9624",
		},
		{
			s:           "-897668",
			expectedStr: "c8facb588b36948d5da0e3a7e16977a701331b0a",
		},
	}

	for _, test := range tests {
		result, _ := HashString(test.s)
		if test.expectedStr != result {
			t.Errorf("Expected %s, got %s", test.expectedStr, result)
		}
	}
}
