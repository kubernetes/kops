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
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"k8s.io/kops/pkg/diff"
)

func TestWriteLocalsOutputs(t *testing.T) {
	cases := []struct {
		name        string
		values      map[string]*terraformOutputVariable
		expected    string
		errExpected bool
	}{
		{
			name:     "empty map",
			expected: "",
		},
		{
			name: "single output",
			values: map[string]*terraformOutputVariable{
				"key1": {
					Key:   "key1",
					Value: LiteralFromStringValue("value1"),
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
			values: map[string]*terraformOutputVariable{
				"key1": {
					Key: "key1",
					ValueArray: []*Literal{
						LiteralFromStringValue("value2"),
						LiteralFromStringValue("value1"),
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
		{
			name: "duplicate names",
			values: map[string]*terraformOutputVariable{
				"key.1": {
					Key:   "key.1",
					Value: LiteralFromStringValue("value1"),
				},
				"key-1": {
					Key:   "key-1",
					Value: LiteralFromStringValue("value2"),
				},
			},
			errExpected: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := hclwrite.NewEmptyFile()
			root := f.Body()
			err := writeLocalsOutputs(root, tc.values)
			if tc.errExpected {
				if err == nil {
					t.Errorf("did not find expected error")
					t.FailNow()
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %e", err)
				t.FailNow()
			}
			actual := strings.TrimSpace(string(f.Bytes()))
			expected := strings.TrimSpace(tc.expected)
			if actual != expected {
				diffString := diff.FormatDiff(expected, string(actual))
				t.Logf("diff:\n%s\n", diffString)
				t.Errorf("expected: '%s', got: '%s'\n", expected, actual)
			}
		})
	}
}
