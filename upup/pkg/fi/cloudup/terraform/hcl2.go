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
	"fmt"
	"reflect"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// writeValue writes a cty Value to a body with a given key
// It detects the value type and uses the more specific functions below
func writeValue(body *hclwrite.Body, key string, value cty.Value) {
	if value.IsNull() {
		return
	}
	if value.Type().IsListType() {
		if value.LengthInt() == 0 {
			return
		}
		literals := make([]*Literal, 0)

		for _, val := range value.AsValueSlice() {
			refLiteral := reflect.New(reflect.TypeOf(Literal{}))
			err := gocty.FromCtyValue(val, refLiteral.Interface())
			if literal, ok := refLiteral.Interface().(*Literal); err == nil && ok {
				literals = append(literals, literal)
			}
		}
		if len(literals) > 0 {
			writeLiteralList(body, key, literals)
		} else {
			// We assume that if the first element of a list is an object
			// then all of the elements are objects
			firstElement := value.Index(cty.NumberIntVal(0))
			if firstElement.Type().IsObjectType() {
				it := value.ElementIterator()
				for it.Next() {
					_, v := it.Element()
					childBlock := body.AppendNewBlock(key, nil)
					childBody := childBlock.Body()
					keys := make([]string, 0)
					for k := range v.Type().AttributeTypes() {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						writeValue(childBody, k, v.GetAttr(k))
					}
				}
			} else {
				body.SetAttributeValue(key, value)
			}
		}
	} else {
		refLiteral := reflect.New(reflect.TypeOf(Literal{}))
		err := gocty.FromCtyValue(value, refLiteral.Interface())
		if literal, ok := refLiteral.Interface().(*Literal); err == nil && ok {
			writeLiteral(body, key, literal)
		} else if value.Type().IsObjectType() {
			childBlock := body.AppendNewBlock(key, nil)
			childBody := childBlock.Body()
			keys := make([]string, 0)
			for k := range value.Type().AttributeTypes() {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				writeValue(childBody, k, value.GetAttr(k))
			}
		} else if value.Type().IsMapType() {
			writeMap(body, key, value.AsValueMap())
		} else {
			body.SetAttributeValue(key, value)
		}
	}
}

// writeLiteral will write a literal attribute to a body
// Examples:
// key = "value1"
// key = res_type.res_name.res_prop
// key = file("${module.path}/foo")
func writeLiteral(body *hclwrite.Body, key string, literal *Literal) {
	if literal.FilePath != "" {
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf("file(%q)", literal.FilePath)),
			},
		}
		body.SetAttributeRaw(key, tokens)
	} else if literal.ResourceType == "" || literal.ResourceName == "" || literal.ResourceProp == "" {
		body.SetAttributeValue(key, cty.StringVal(literal.Value))
	} else {
		traversal := hcl.Traversal{
			hcl.TraverseRoot{Name: literal.ResourceType},
			hcl.TraverseAttr{Name: literal.ResourceName},
			hcl.TraverseAttr{Name: literal.ResourceProp},
		}
		body.SetAttributeTraversal(key, traversal)
	}
}

// writeLiteralList writes a list of literals to a body
// Example:
// key = [type1.name1.attr1, type2.name2.attr2, "stringliteral"]
//
// The HCL2 library does not support this natively. See https://github.com/hashicorp/hcl/issues/347
func writeLiteralList(body *hclwrite.Body, key string, literals []*Literal) {
	tokens := hclwrite.Tokens{
		{Type: hclsyntax.TokenOBrack, Bytes: []byte("["), SpacesBefore: 1},
	}
	for i, literal := range literals {
		if literal.ResourceType == "" || literal.ResourceName == "" || literal.ResourceProp == "" {
			tokens = append(tokens, []*hclwrite.Token{
				{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(literal.Value)},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
			}...)
		} else {
			tokens = append(tokens, []*hclwrite.Token{
				{Type: hclsyntax.TokenStringLit, Bytes: []byte(literal.ResourceType), SpacesBefore: 1},
				{Type: hclsyntax.TokenDot, Bytes: []byte(".")},
				{Type: hclsyntax.TokenStringLit, Bytes: []byte(literal.ResourceName)},
				{Type: hclsyntax.TokenDot, Bytes: []byte(".")},
				{Type: hclsyntax.TokenStringLit, Bytes: []byte(literal.ResourceProp)},
			}...)
		}
		if i < len(literals)-1 {
			tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")})
		}
	}
	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte("]")})
	body.SetAttributeRaw(key, tokens)
}

// writeMap writes a map's key-value pairs to a body spready across multiple lines.
// Example:
// key = {
//   "key1" = "value1"
//   "key2" = "value2"
// }
//
// The HCL2 library does not support this natively. See https://github.com/hashicorp/hcl/issues/356
func writeMap(body *hclwrite.Body, key string, values map[string]cty.Value) {
	if len(values) == 0 {
		return
	}
	tokens := hclwrite.Tokens{
		{Type: hclsyntax.TokenOBrace, Bytes: []byte("{"), SpacesBefore: 1},
		{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		tokens = append(tokens, []*hclwrite.Token{
			{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
			{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(k)},
			{Type: hclsyntax.TokenCQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
			{Type: hclsyntax.TokenEqual, Bytes: []byte("="), SpacesBefore: 1},
		}...)

		v := values[k]

		refLiteral := reflect.New(reflect.TypeOf(Literal{}))
		err := gocty.FromCtyValue(v, refLiteral.Interface())
		// If this is a map of literals then do not surround the value with quotes
		if literal, ok := refLiteral.Interface().(*Literal); err == nil && ok {
			// For maps of literals we currently only support file references
			// If we ever need to support a map of strings to resource property references that can be added here
			if literal.FilePath != "" {
				tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("file(%q)", literal.FilePath))})
			} else if literal.Value != "" {
				tokens = append(tokens, []*hclwrite.Token{
					{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
					{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(literal.Value)},
					{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
				}...)
			}
		} else {
			tokens = append(tokens, []*hclwrite.Token{
				{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(v.AsString())},
				{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
			}...)
		}
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
	}
	tokens = append(tokens,
		&hclwrite.Token{Type: hclsyntax.TokenCBrace, Bytes: []byte("}")},
	)
	body.SetAttributeRaw(key, tokens)
}
