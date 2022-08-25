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
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
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
		literals := make([]*terraformWriter.Literal, 0)

		for _, val := range value.AsValueSlice() {
			refLiteral := reflect.New(reflect.TypeOf(terraformWriter.Literal{}))
			err := gocty.FromCtyValue(val, refLiteral.Interface())
			if literal, ok := refLiteral.Interface().(*terraformWriter.Literal); err == nil && ok {
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
		refLiteral := reflect.New(reflect.TypeOf(terraformWriter.Literal{}))
		err := gocty.FromCtyValue(value, refLiteral.Interface())
		if literal, ok := refLiteral.Interface().(*terraformWriter.Literal); err == nil && ok {
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
func writeLiteral(body *hclwrite.Body, key string, literal *terraformWriter.Literal) {
	if literal.FnName != "" {
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf("%v(%v)", literal.FnName, strings.Join(literal.FnArgs, ", "))),
			},
		}
		body.SetAttributeRaw(key, tokens)
	} else if literal.Index {
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenOQuote,
				Bytes: []byte(`"`),
			},
			{
				Type:  hclsyntax.TokenQuotedLit,
				Bytes: []byte(literal.Value),
			},
			{
				Type:  hclsyntax.TokenQuotedLit,
				Bytes: []byte(`-`),
			},
			{
				Type:  hclsyntax.TokenTemplateInterp,
				Bytes: []byte(`${`),
			},
			{
				Type:  hclsyntax.TokenQuotedLit,
				Bytes: []byte(`count.index`),
			},
			{
				Type:  hclsyntax.TokenTemplateSeqEnd,
				Bytes: []byte(`}`),
			},
			{
				Type:  hclsyntax.TokenCQuote,
				Bytes: []byte(`"`),
			},
			{
				Type:  hclsyntax.TokenEOF,
				Bytes: []byte{},
			},
		}
		body.SetAttributeRaw(key, tokens)
	} else if len(literal.Tokens) == 0 {
		body.SetAttributeValue(key, cty.StringVal(literal.Value))
	} else {
		traversal := hcl.Traversal{
			hcl.TraverseRoot{Name: literal.Tokens[0]},
		}
		for i := 1; i < len(literal.Tokens); i++ {
			token := literal.Tokens[i]
			traversal = append(traversal, hcl.TraverseAttr{Name: token})
		}
		body.SetAttributeTraversal(key, traversal)
	}
}

// literalListTokens returns the tokens of a list of literals
// Example:
// key = [type1.name1.attr1, type2.name2.attr2, "stringliteral"]
func literalListTokens(literals []*terraformWriter.Literal) hclwrite.Tokens {
	tokens := hclwrite.Tokens{
		{Type: hclsyntax.TokenOBrack, Bytes: []byte("["), SpacesBefore: 1},
	}
	for i, literal := range literals {
		if len(literal.Tokens) == 0 {
			tokens = append(tokens, []*hclwrite.Token{
				{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
				{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(literal.Value)},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
			}...)
		} else {
			tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenStringLit, Bytes: []byte(literal.Tokens[0]), SpacesBefore: 1})
			for i := 1; i < len(literal.Tokens); i++ {
				token := literal.Tokens[i]
				tokens = append(tokens, []*hclwrite.Token{
					{Type: hclsyntax.TokenDot, Bytes: []byte(".")},
					{Type: hclsyntax.TokenStringLit, Bytes: []byte(token)},
				}...)
			}
		}
		if i < len(literals)-1 {
			tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")})
		}
	}
	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte("]")})
	return tokens
}

// writeLiteralList writes a list of literals to a body
// Example:
// key = [type1.name1.attr1, type2.name2.attr2, "stringliteral"]
//
// The HCL2 library does not support this natively. See https://github.com/hashicorp/hcl/issues/347
func writeLiteralList(body *hclwrite.Body, key string, literals []*terraformWriter.Literal) {
	body.SetAttributeRaw(key, literalListTokens(literals))
}

// writeMap writes a map's key-value pairs to a body spready across multiple lines.
// Example:
//
//	key = {
//	  "key1" = "value1"
//	  "key2" = "value2"
//	}
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

		refLiteral := reflect.New(reflect.TypeOf(terraformWriter.Literal{}))
		errLiteral := gocty.FromCtyValue(v, refLiteral.Interface())

		refLiteralSlice := reflect.New(reflect.TypeOf([]*terraformWriter.Literal{}))
		errLiteralSlice := gocty.FromCtyValue(v, refLiteralSlice.Interface())
		// If this is a map of literals then do not surround the value with quotes
		if literal, ok := refLiteral.Interface().(*terraformWriter.Literal); errLiteral == nil && ok {
			if literal.FnName != "" {
				tokens = append(tokens, &hclwrite.Token{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(fmt.Sprintf("%v(%v)", literal.FnName, strings.Join(literal.FnArgs, ", "))),
				})
			} else if literal.Value != "" {
				tokens = append(tokens, []*hclwrite.Token{
					{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
					{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(literal.Value)},
					{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}, SpacesBefore: 1},
				}...)
			}
		} else if literals, ok := refLiteralSlice.Interface().(*[]*terraformWriter.Literal); errLiteralSlice == nil && ok {
			tokens = append(tokens, literalListTokens(*literals)...)
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
