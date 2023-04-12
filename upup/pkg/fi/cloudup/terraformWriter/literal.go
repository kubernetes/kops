/*
Copyright 2019 The Kubernetes Authors.

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

package terraformWriter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/constraints"
)

// Literal represents a literal in terraform syntax
type Literal struct {
	// String is the Terraform representation.
	String string `cty:"string"`
}

var _ json.Marshaler = &Literal{}

func (l *Literal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&l.String)
}

func (l *Literal) IsSingleValue() bool {
	return true
}

func (l *Literal) Write(buffer *bytes.Buffer, indent int, key string) {
	buffer.WriteString("= ")
	buffer.WriteString(l.String)
	buffer.WriteString("\n")
}

// LiteralFunctionExpression constructs a Literal representing the result of
// calling the supplied functionName with the supplied args.
func LiteralFunctionExpression(functionName string, args ...*Literal) *Literal {
	s := functionName + "("
	for i, arg := range args {
		if i != 0 {
			s += ", "
		}
		s += arg.String
	}
	return &Literal{
		String: s + ")",
	}
}

func LiteralSelfLink(resourceType, resourceName string) *Literal {
	return LiteralProperty(resourceType, resourceName, "self_link")
}

func LiteralData(dataSourceType, dataSourceName, prop string) *Literal {
	tfName := sanitizeName(dataSourceName)
	return &Literal{
		String: "data." + dataSourceType + "." + tfName + "." + prop + "",
	}
}

func LiteralProperty(resourceType, resourceName, prop string) *Literal {
	tfName := sanitizeName(resourceName)
	return &Literal{
		String: resourceType + "." + tfName + "." + prop,
	}
}

func LiteralTokens(tokens ...string) *Literal {
	return &Literal{
		String: strings.Join(tokens, "."),
	}
}

func LiteralFromIntValue[T constraints.Integer](i T) *Literal {
	return &Literal{
		String: fmt.Sprintf("%d", i),
	}
}

func LiteralFromStringValue(s string) *Literal {
	return &Literal{
		String: "\"" + s + "\"",
	}
}

func LiteralWithIndex(s string) *Literal {
	return &Literal{
		String: fmt.Sprintf("\"%s-${count.index}\"", s),
	}
}

// LiteralBinaryExpression constructs a Literal with the result of a binary operator expression.
// It is the caller's responsibility to ensure the supplied parameters do not use operators
// with lower precedence than the supplied operator.
func LiteralBinaryExpression(lhs *Literal, operator string, rhs *Literal) *Literal {
	return &Literal{
		String: fmt.Sprintf("%s %s %s", lhs.String, operator, rhs.String),
	}
}

// LiteralIndexExpression constructs a Literal with the result of accessing the
// supplied collection using the supplied index.
// It is the caller's responsibility to ensure the supplied collection does not use operators
// with lower precedence than the index operator.
func LiteralIndexExpression(collection *Literal, index *Literal) *Literal {
	return &Literal{
		String: fmt.Sprintf("%s[%s]", collection.String, index.String),
	}
}

// LiteralListExpression constructs a Literal consisting of a list of supplied Literals.
func LiteralListExpression(args ...*Literal) *Literal {
	var b strings.Builder
	b.WriteRune('[')
	for i, arg := range args {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(arg.String)
	}
	b.WriteRune(']')
	return &Literal{
		String: b.String(),
	}
}

// LiteralEmptyStrConditionalExpression constructs a Literal which returns `null`
// if the supplied "empty" expression is an empty string, otherwise returns "value".
// It is the caller's responsibility to ensure the supplied parameters do not use operators
// with lower precedence than the conditional operator.
func LiteralEmptyStrConditionalExpression(empty, value *Literal) *Literal {
	return &Literal{
		String: fmt.Sprintf("%s == \"\" ? null : %s", empty.String, value.String),
	}
}

// SortLiterals sorts a list of Literal, by key.  It does so in-place
func SortLiterals(v []*Literal) {
	sort.Slice(v, func(i, j int) bool {
		return v[i].String < v[j].String
	})
}

// dedupLiterals removes any duplicate Literals before returning the slice.
// As a side-effect, it currently returns the Literals in sorted order.
func dedupLiterals(v []*Literal) ([]*Literal, error) {
	if v == nil {
		return nil, nil
	}

	SortLiterals(v)

	var deduped []*Literal
	for i, p := range v {
		if i != 0 && v[i-1].String == v[i].String {
			continue
		}
		deduped = append(deduped, p)
	}

	return deduped, nil
}
