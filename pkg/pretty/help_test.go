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

package pretty

import (
	"testing"
)

func TestLongDesc(t *testing.T) {
	tests := []struct {
		desc        string
		expectedStr string
	}{
		{
			desc:        "    test1    ",
			expectedStr: "test1",
		},
		{
			desc:        "test1\n   test2\n",
			expectedStr: "test1\ntest2",
		},
	}
	for _, test := range tests {
		result := LongDesc(test.desc)
		if test.expectedStr != result {
			t.Errorf("Expected %s, got %s", test.expectedStr, result)
		}
	}
}
