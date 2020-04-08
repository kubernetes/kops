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

package systemd

import (
	"testing"
)

func TestEscapeCommand(t *testing.T) {
	tests := []struct {
		argv        []string
		expectedStr string
	}{
		{
			argv:        []string{`a`, `(b)`, `\c`, `\\d`, `"`, ` `},
			expectedStr: `a (b) \\c \\\\d \" " "`,
		},
		{
			argv:        []string{"/usr/bin/docker", "pull", "busybox:latest"},
			expectedStr: "/usr/bin/docker pull busybox:latest",
		},
	}
	for _, test := range tests {
		result := EscapeCommand(test.argv)
		if test.expectedStr != result {
			t.Errorf("Expected %s, got %s", test.expectedStr, result)
		}
	}
}
