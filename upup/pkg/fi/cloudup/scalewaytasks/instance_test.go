/*
Copyright 2023 The Kubernetes Authors.

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

package scalewaytasks

import (
	"strconv"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
)

func TestFindFirstFreeIndex(t *testing.T) {
	igName := "control-plane-1"
	type TestCase struct {
		Actual   []int
		Expected int
	}
	testCases := []TestCase{
		{
			Actual:   []int{},
			Expected: 0,
		},
		{
			Actual:   []int{0, 1, 2, 3},
			Expected: 4,
		},
		{
			Actual:   []int{0, 2, 1},
			Expected: 3,
		},
		{
			Actual:   []int{1, 2, 4},
			Expected: 0,
		},
		{
			Actual:   []int{4, 5, 2, 3, 0},
			Expected: 1,
		},
	}

	for _, testCase := range testCases {
		existing := []*instance.Server(nil)
		for _, i := range testCase.Actual {
			existing = append(existing, &instance.Server{Name: igName + "-" + strconv.Itoa(i)})
		}
		index := findFirstFreeIndex(existing)
		if index != testCase.Expected {
			t.Errorf("Expected %d, got %d", testCase.Expected, index)
		}
	}
}
