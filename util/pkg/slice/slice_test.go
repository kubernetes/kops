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

package slice

import (
	"reflect"
	"testing"
)

func TestGetUniqueStrings(t *testing.T) {
	tests := []struct {
		mainStr     []string
		extraStr    []string
		expectedStr []string
	}{
		{
			mainStr:     []string{"a", "b"},
			extraStr:    []string{"a", "b", "c", "d"},
			expectedStr: []string{"c", "d"},
		},
		{
			mainStr:     []string{"a", "b"},
			extraStr:    []string{"a", "b"},
			expectedStr: []string{},
		},
	}
	for _, test := range tests {
		uniqueStr := GetUniqueStrings(test.mainStr, test.extraStr)
		if !reflect.DeepEqual(test.expectedStr, uniqueStr) {
			t.Errorf("Expected %v, got %v", test.expectedStr, uniqueStr)
		}
	}
}
