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
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"k8s.io/kops/pkg/diff"
)

func TestWriteValue(t *testing.T) {
	cases := []struct {
		name     string
		value    cty.Value
		expected string
	}{
		{
			name:     "null",
			value:    cty.NullVal(cty.String),
			expected: "",
		},
		{
			name:     "empty list",
			value:    cty.ListValEmpty(cty.String),
			expected: "",
		},

		{
			name:     "string",
			value:    cty.StringVal("value"),
			expected: `foo = "value"`,
		},
		{
			name:     "list",
			value:    cty.ListVal([]cty.Value{cty.StringVal("val1"), cty.StringVal("val2")}),
			expected: `foo = ["val1", "val2"]`,
		},
		{
			name: "list of objects",
			value: cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"key1": cty.StringVal("val1"),
					"key2": cty.NumberIntVal(10),
					"key3": cty.ListVal([]cty.Value{cty.StringVal("val2"), cty.StringVal("val3")}),
					"key4": cty.BoolVal(true),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"key1": cty.StringVal("val4"),
					"key2": cty.NumberIntVal(100),
					"key3": cty.ListVal([]cty.Value{cty.StringVal("val5")}),
					"key4": cty.BoolVal(false),
				}),
			}),
			expected: `
foo {
  key1 = "val1"
  key2 = 10
  key3 = ["val2", "val3"]
  key4 = true
}
foo {
  key1 = "val4"
  key2 = 100
  key3 = ["val5"]
  key4 = false
}`,
		},
		{
			name: "object block",
			value: cty.ObjectVal(map[string]cty.Value{
				"attr2": cty.StringVal("val1"),
				"attr1": cty.BoolVal(true),
			}),
			expected: `
foo {
  attr1 = true
  attr2 = "val1"
}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := hclwrite.NewEmptyFile()
			root := f.Body()
			writeValue(root, "foo", tc.value)
			actual := strings.TrimSpace(string(f.Bytes()))
			expected := strings.TrimSpace(tc.expected)
			if actual != expected {
				t.Errorf("expected: '%s', got: '%s'\n", expected, actual)
			}
		})
	}
}

func TestWriteLiteral(t *testing.T) {
	cases := []struct {
		name     string
		literal  *Literal
		expected string
	}{
		{
			name:     "string",
			literal:  &Literal{Value: "value"},
			expected: `foo = "value"`,
		},
		{
			name: "traversal",
			literal: &Literal{
				ResourceType: "type",
				ResourceName: "name",
				ResourceProp: "prop",
			},
			expected: "foo = type.name.prop",
		},
		{
			name: "file",
			literal: &Literal{
				FilePath: "${path.module}/foo",
			},
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
		literals []*Literal
		expected string
	}{
		{
			name:     "empty list",
			expected: "foo = []",
		},
		{
			name: "one literal",
			literals: []*Literal{
				{
					ResourceType: "type",
					ResourceName: "name",
					ResourceProp: "prop",
				},
			},
			expected: "foo = [type.name.prop]",
		},
		{
			name: "two literals",
			literals: []*Literal{
				{
					ResourceType: "type1",
					ResourceName: "name1",
					ResourceProp: "prop1",
				},
				{
					ResourceType: "type2",
					ResourceName: "name2",
					ResourceProp: "prop2",
				},
			},
			expected: "foo = [type1.name1.prop1, type2.name2.prop2]",
		},
		{
			name: "one traversal literal, one string literal",
			literals: []*Literal{
				{
					ResourceType: "type",
					ResourceName: "name",
					ResourceProp: "prop",
				},
				{
					Value: "foobar",
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

func TestWriteMap(t *testing.T) {
	cases := []struct {
		name     string
		values   map[string]cty.Value
		expected string
	}{
		{
			name:     "empty map",
			expected: "",
		},
		{
			name: "simple map",
			values: map[string]cty.Value{
				"key1": cty.StringVal("value1"),
			},
			expected: `
tags = {
  "key1" = "value1"
}`,
		},
		{
			name: "complex keys",
			values: map[string]cty.Value{
				"key1.k8s.local/foo": cty.StringVal("value1"),
			},
			expected: `
tags = {
  "key1.k8s.local/foo" = "value1"
}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := hclwrite.NewEmptyFile()
			root := f.Body()
			writeMap(root, "tags", tc.values)
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

func TestWriteMapLiterals(t *testing.T) {
	cases := []struct {
		name     string
		values   map[string]Literal
		expected string
	}{
		{
			name: "literal values",
			values: map[string]Literal{
				"key1": {FilePath: "${module.path}/path/to/value1"},
				"key2": {FilePath: "${module.path}/path/to/value2"},
			},
			expected: `
metadata = {
  "key1" = file("${module.path}/path/to/value1")
  "key2" = file("${module.path}/path/to/value2")
}
			`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			literalMap := make(map[string]cty.Value)
			for k, v := range tc.values {
				literalType, err := gocty.ImpliedType(v)
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
				literalVal, err := gocty.ToCtyValue(v, literalType)

				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
				literalMap[k] = literalVal
			}

			f := hclwrite.NewEmptyFile()
			root := f.Body()
			writeMap(root, "metadata", literalMap)
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
