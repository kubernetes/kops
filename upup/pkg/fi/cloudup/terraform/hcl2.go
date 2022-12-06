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
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// writeLiteral will write a literal attribute to a body
// Examples:
// key = "value1"
// key = res_type.res_name.res_prop
// key = file("${module.path}/foo")
func writeLiteral(body *hclwrite.Body, key string, literal *terraformWriter.Literal) {
	body.SetAttributeRaw(key, hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(literal.String),
		},
	})
}

// literalListTokens returns the tokens of a list of literals
// Example:
// key = [type1.name1.attr1, type2.name2.attr2, "stringliteral"]
func literalListTokens(literals []*terraformWriter.Literal) hclwrite.Tokens {
	tokens := hclwrite.Tokens{
		{Type: hclsyntax.TokenOBrack, Bytes: []byte("["), SpacesBefore: 1},
	}
	for i, literal := range literals {
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(literal.String)})
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

type mapStringLiteral struct {
	members map[string]*terraformWriter.Literal
}

func (m *mapStringLiteral) IsSingleValue() bool {
	return false
}

func (m *mapStringLiteral) ToObject() element {
	o := &object{field: make(map[string]element, len(m.members))}
	for k, v := range m.members {
		o.field[k] = v
	}
	return o
}

// write writes a map's key-value pairs to a body spread across multiple lines.
// Example:
//
//	key = {
//	  "key1" = "value1"
//	  "key2" = "value2"
//	}
func (m *mapStringLiteral) Write(buffer *bytes.Buffer, indent int, key string) {
	if len(m.members) == 0 {
		return
	}
	writeIndent(buffer, indent)
	buffer.WriteString(key)
	buffer.WriteString(" = {\n")
	keys := make([]string, 0, len(m.members))
	maxKeyLen := 0
	for k := range m.members {
		kLen := len(quote(k))
		if kLen > maxKeyLen {
			maxKeyLen = kLen
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		writeIndent(buffer, indent+2)
		quoted := quote(k)
		buffer.WriteString(quoted)
		writeIndent(buffer, maxKeyLen-len(quoted))
		buffer.WriteString(" = ")
		buffer.WriteString(m.members[k].String)
		buffer.WriteRune('\n')
	}
	writeIndent(buffer, indent)
	buffer.WriteString("}\n")
}

func mapToElement(item interface{}) *mapStringLiteral {
	v := reflect.ValueOf(item)
	if v.Kind() != reflect.Map {
		panic(fmt.Sprintf("not a map type %s", v.Kind()))
	}
	if v.Type().Key().Kind() != reflect.String {
		panic(fmt.Sprintf("unhandled map key type %s", v.Type().Key().Kind()))
	}
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Pointer && elemType.Elem() == literalType {
		o := &mapStringLiteral{members: make(map[string]*terraformWriter.Literal, v.Len())}
		for _, key := range v.MapKeys() {
			o.members[key.String()] = v.MapIndex(key).Interface().(*terraformWriter.Literal)
		}
		return o
	}
	if elemType.Kind() != reflect.String {
		panic(fmt.Sprintf("unhandled map value type %s", elemType.Kind()))
	}
	o := &mapStringLiteral{members: make(map[string]*terraformWriter.Literal, v.Len())}
	for _, key := range v.MapKeys() {
		o.members[key.String()] = terraformWriter.LiteralFromStringValue(v.MapIndex(key).String())
	}
	return o
}

func writeIndent(buf *bytes.Buffer, indent int) {
	for i := 0; i < indent; i++ {
		buf.WriteString(" ")
	}
}

func quote(s string) string {
	var b strings.Builder
	b.WriteRune('"')
	for _, r := range s {
		if r == '\\' || r == '"' {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	b.WriteRune('"')
	return b.String()
}
