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

package urls

import (
	"testing"
)

func TestJoin(t *testing.T) {
	tests := []struct {
		base        string
		other1      string
		other2      string
		expectedStr string
	}{
		{
			base:        "/test",
			other1:      "z1",
			other2:      "/z2",
			expectedStr: "/test/z1/z2",
		},
		{
			base:        "test/",
			other1:      "z1",
			other2:      "/z2",
			expectedStr: "test/z1/z2",
		},
	}
	for _, test := range tests {
		result := Join(test.base, test.other1, test.other2)
		if test.expectedStr != result {
			t.Errorf("Expected %s, got %s", test.expectedStr, result)
		}
	}
}
