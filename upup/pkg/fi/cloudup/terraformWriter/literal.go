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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Literal represents a literal in terraform syntax
type Literal struct {
	// String is the Terraform representation.
	String string `cty:"string"`

	// FnArgs contains string representations of arguments to the function call.
	// Any string arguments must be quoted.
	FnArgs []string `cty:"fn_arg"`
}

var _ json.Marshaler = &Literal{}

func (l *Literal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&l.String)
}

func LiteralFunctionExpression(functionName string, args ...string) *Literal {
	return &Literal{
		String: fmt.Sprintf("%v(%v)", functionName, strings.Join(args, ", ")),
		FnArgs: args,
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
