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

package ui

import (
	"reflect"
	"testing"
)

type testWriter struct {
	buffer string
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.buffer += string(p)
	return len(p), nil
}

func TestGetConfirm(t *testing.T) {
	var writer testWriter
	var args *ConfirmArgs = &ConfirmArgs{
		Out:     &writer,
		Message: "test",
		Default: "yes",
	}

	cases := []struct {
		input       string
		expected    bool
		expectedOut string
	}{
		{
			input:       "yes",
			expected:    true,
			expectedOut: "test (Y/n)\n",
		},
		{
			input:       "no",
			expected:    false,
			expectedOut: "test (Y/n)\n",
		},
		{
			input:       "invalid",
			expected:    false,
			expectedOut: "test (Y/n)\n",
		},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			writer.buffer = ""
			args.TestVal = c.input

			got, _ := GetConfirm(args)

			if !reflect.DeepEqual(c.expectedOut, writer.buffer) {
				t.Errorf("expectedOut: %v, got: %v", c.expectedOut, writer.buffer)
			}
			if !reflect.DeepEqual(c.expected, got) {
				t.Errorf("expected: %v, got: %v", c.expected, got)
			}
		})
	}
}
