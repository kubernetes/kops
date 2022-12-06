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

// writeMap writes a map's key-value pairs to a body spread across multiple lines.
// Example:
//
//	key = {
//	  "key1" = "value1"
//	  "key2" = "value2"
//	}
func writeMap(buf *bytes.Buffer, indent int, key string, values map[string]*terraformWriter.Literal) {
	if len(values) == 0 {
		return
	}
	writeIndent(buf, indent)
	buf.WriteString(key)
	buf.WriteString(" = {\n")
	keys := make([]string, 0, len(values))
	maxKeyLen := 0
	for k := range values {
		kLen := len(quote(k))
		if kLen > maxKeyLen {
			maxKeyLen = kLen
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		writeIndent(buf, indent+2)
		quoted := quote(k)
		buf.WriteString(quoted)
		writeIndent(buf, maxKeyLen-len(quoted))
		buf.WriteString(" = ")
		buf.WriteString(values[k].String)
		buf.WriteRune('\n')
	}
	writeIndent(buf, indent)
	buf.WriteString("}\n")
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
