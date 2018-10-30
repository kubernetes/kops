/*
Copyright 2017 The Kubernetes Authors.

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

package systemd

import (
	"testing"
)

func TestUnitFileExtensionValid(t *testing.T) {
	type testCase struct {
		name     string
		fileName string
		expected bool
	}

	// expectedWords maps boolean return value onto the word "valid" or
	// "invalid". If we expected the return value to be false, then we expected
	// the test name to be invalid; if we expected the return value to be true,
	// then we expected the test name to be valid.
	expectedWords := map[bool]string{
		false: "invalid",
		true:  "valid",
	}

	// Test all valid extensions, plus two invalid extensions.
	testcases := make([]testCase, 0, len(UnitExtensions)+2)

	// Add a test case for each valid extension.
	for _, ext := range UnitExtensions {
		testcases = append(testcases, testCase{
			"valid extension: " + ext,
			"my-unit" + ext,
			true,
		})
	}

	// Add two testcases for no extension and invalid extension.
	testcases = append(testcases, testCase{
		"invalid extension: (no extension)",
		"my-unit",
		false,
	})
	testcases = append(testcases, testCase{
		"invalid extension: .not-valid",
		"my-unit.not-valid",
		false,
	})

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if UnitFileExtensionValid(tc.fileName) != tc.expected {
				t.Errorf("expected %v to be %v, but was %v", tc.fileName,
					expectedWords[tc.expected], expectedWords[!tc.expected])
			}
		})
	}
}
