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

package golden

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/kops/pkg/diff"
)

// AssertMatchesFile matches the actual value to a with expected file.
// If HACK_UPDATE_EXPECTED_IN_PLACE is set, it will write the actual value to the expected file,
// which is very handy when updating our tests.
func AssertMatchesFile(t *testing.T, actual string, p string) {
	actual = strings.TrimSpace(actual)

	expectedBytes, err := ioutil.ReadFile(p)
	if err != nil {
		if !os.IsNotExist(err) || os.Getenv("HACK_UPDATE_EXPECTED_IN_PLACE") == "" {
			t.Fatalf("error reading file %q: %v", p, err)
		}
	}
	expected := strings.TrimSpace(string(expectedBytes))

	//on windows, with git set to autocrlf, the reference files on disk have windows line endings
	expected = strings.Replace(expected, "\r\n", "\n", -1)

	if actual == expected {
		return
	}

	if os.Getenv("HACK_UPDATE_EXPECTED_IN_PLACE") != "" {
		t.Logf("HACK_UPDATE_EXPECTED_IN_PLACE: writing expected output %s", p)

		// Keep git happy with a trailing newline
		actual += "\n"

		if err := ioutil.WriteFile(p, []byte(actual), 0644); err != nil {
			t.Errorf("error writing expected output %s: %v", p, err)
		}

		// Keep going so we write all files in a test
		t.Errorf("output did not match expected for %q", p)
		return
	}

	diffString := diff.FormatDiff(expected, actual)
	t.Logf("diff:\n%s\n", diffString)

	abs, err := filepath.Abs(p)
	if err != nil {
		t.Errorf("unable to get absolute path for %q: %v", p, err)
	} else {
		p = abs
	}

	t.Logf("to update golden output automatically, run hack/update-expected.sh")

	t.Errorf("output did not match expected for %q", p)
}
