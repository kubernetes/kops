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

package exec

import (
	"reflect"
	"testing"
)

func TestWithTee(t *testing.T) {
	tests := []struct {
		cmd      string
		args     []string
		logfile  string
		expected []string
	}{
		{
			cmd:     "ls",
			args:    []string{"-l", "-a"},
			logfile: "/var/log/ls.log",
			expected: []string{
				"/bin/sh",
				"-c",
				"mkfifo /tmp/pipe; (tee -a /var/log/ls.log < /tmp/pipe & ) ; exec ls -l -a > /tmp/pipe 2>&1",
			},
		},
	}

	for _, test := range tests {
		result := WithTee(test.cmd, test.args, test.logfile)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Actual result %v, expected %v", result, test.expected)
		}
	}
}
