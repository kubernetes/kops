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
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

func TestWriteLiteral(t *testing.T) {
	cases := []struct {
		name     string
		literal  *terraformWriter.Literal
		expected string
	}{
		{
			name:     "string",
			literal:  terraformWriter.LiteralFromStringValue("value"),
			expected: `foo = "value"`,
		},
		{
			name:     "traversal",
			literal:  terraformWriter.LiteralProperty("type", "name", "prop"),
			expected: "foo = type.name.prop",
		},
		{
			name:     "provider alias",
			literal:  terraformWriter.LiteralTokens("aws", "files"),
			expected: "foo = aws.files",
		},

		{
			name:     "file",
			literal:  terraformWriter.LiteralFunctionExpression("file", terraformWriter.LiteralFromStringValue("${path.module}/foo")),
			expected: `foo = file("${path.module}/foo")`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := hclwrite.NewEmptyFile()
			root := f.Body()
			writeLiteral(root, "foo", tc.literal)
			actual := strings.TrimSpace(string(f.Bytes()))
			expected := strings.TrimSpace(tc.expected)
			if actual != expected {
				t.Errorf("expected: '%s', got: '%s'\n", expected, actual)
			}
		})
	}
}

func TestWriteLiteralList(t *testing.T) {
	cases := []struct {
		name     string
		literals []*terraformWriter.Literal
		expected string
	}{
		{
			name:     "empty list",
			expected: "foo = []",
		},
		{
			name: "one literal",
			literals: []*terraformWriter.Literal{
				{
					String: "type.name.prop",
				},
			},
			expected: "foo = [type.name.prop]",
		},
		{
			name: "two literals",
			literals: []*terraformWriter.Literal{
				{
					String: "type1.name1.prop1",
				},
				{
					String: "type2.name2.prop2",
				},
			},
			expected: "foo = [type1.name1.prop1, type2.name2.prop2]",
		},
		{
			name: "one traversal literal, one string literal",
			literals: []*terraformWriter.Literal{
				{
					String: "type.name.prop",
				},
				{
					String: "\"foobar\"",
				},
			},
			expected: `foo = [type.name.prop, "foobar"]`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := hclwrite.NewEmptyFile()
			root := f.Body()
			writeLiteralList(root, "foo", tc.literals)
			actual := strings.TrimSpace(string(f.Bytes()))
			expected := strings.TrimSpace(tc.expected)
			if actual != expected {
				t.Errorf("expected: '%s', got: '%s'\n", expected, actual)
			}
		})
	}
}
