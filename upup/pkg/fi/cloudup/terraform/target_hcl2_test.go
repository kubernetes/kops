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

package terraform

import (
	"bytes"
	"strings"
	"testing"

	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

func TestWriteLocalsOutputs(t *testing.T) {
	cases := []struct {
		name     string
		values   map[string]terraformWriter.OutputValue
		expected string
	}{
		{
			name:     "empty map",
			expected: "",
		},
		{
			name: "single output",
			values: map[string]terraformWriter.OutputValue{
				"key1": {
					Value: terraformWriter.LiteralFromStringValue("value1"),
				},
			},
			expected: `
locals {
  key1 = "value1"
}

output "key1" {
  value = "value1"
}`,
		},
		{
			name: "list output",
			values: map[string]terraformWriter.OutputValue{
				"key1": {
					ValueArray: []*terraformWriter.Literal{
						terraformWriter.LiteralFromStringValue("value1"),
						terraformWriter.LiteralFromStringValue("value2"),
					},
				},
			},
			expected: `
locals {
  key1 = ["value1", "value2"]
}

output "key1" {
  value = ["value1", "value2"]
}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			writeLocalsOutputs(buf, tc.values)
			actual := strings.TrimSpace(buf.String())
			expected := strings.TrimSpace(tc.expected)
			if actual != expected {
				diffString := diff.FormatDiff(expected, string(actual))
				t.Logf("diff:\n%s\n", diffString)
				t.Errorf("expected: '%s', got: '%s'\n", expected, actual)
			}
		})
	}
}
