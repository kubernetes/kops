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

	"k8s.io/klog/v2"
)

// Literal represents a literal in terraform syntax
type Literal struct {
	// Value is used to support Terraform's "${}" interpolation.
	Value string `cty:"value"`
	// Index to support the index of the count meta-argument.
	Index bool `cty:"index"`

	// Tokens are portions of a literal reference joined by periods.
	// example: {"aws_vpc", "foo", "id"}
	Tokens []string `cty:"tokens"`

	// FnName represents the name of a terraform function.
	FnName string `cty:"fn_name"`
	// FnArgs contains string representations of arguments to the function call.
	// Any string arguments must be quoted.
	FnArgs []string `cty:"fn_arg"`
}

var _ json.Marshaler = &Literal{}

func (l *Literal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&l.Value)
}

func LiteralFunctionExpression(functionName string, args []string) *Literal {
	return &Literal{
		Value:  fmt.Sprintf("${%v(%v)}", functionName, strings.Join(args, ", ")),
		FnName: functionName,
		FnArgs: args,
	}
}

func LiteralSelfLink(resourceType, resourceName string) *Literal {
	return LiteralProperty(resourceType, resourceName, "self_link")
}

func LiteralProperty(resourceType, resourceName, prop string) *Literal {
	tfName := sanitizeName(resourceName)
	expr := "${" + resourceType + "." + tfName + "." + prop + "}"
	return &Literal{
		Value:  expr,
		Tokens: []string{resourceType, tfName, prop},
	}
}

func LiteralTokens(tokens ...string) *Literal {
	expr := "${" + strings.Join(tokens, ".") + "}"
	return &Literal{
		Value:  expr,
		Tokens: tokens,
	}
}

func LiteralFromStringValue(s string) *Literal {
	return &Literal{Value: s}
}

func LiteralWithIndex(s string) *Literal {
	return &Literal{Value: s, Index: true}
}

type literalWithJSON struct {
	literal *Literal
	key     string
}

type byKey []*literalWithJSON

func (a byKey) Len() int      { return len(a) }
func (a byKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byKey) Less(i, j int) bool {
	return a[i].key < a[j].key
}

// buildSortProxies maps a list of Literals to a list of literalWithJSON
func buildSortProxies(v []*Literal) ([]*literalWithJSON, error) {
	var proxies []*literalWithJSON
	for _, l := range v {
		k, err := json.Marshal(l)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, &literalWithJSON{
			literal: l,
			key:     string(k),
		})
	}

	return proxies, nil
}

// SortLiterals sorts a list of Literal, by key.  It does so in-place
func SortLiterals(v []*Literal) {
	proxies, err := buildSortProxies(v)
	if err != nil {
		// Very unexpected
		klog.Fatalf("error processing terraform Literal: %v", err)
	}

	sort.Sort(byKey(proxies))

	for i := range proxies {
		v[i] = proxies[i].literal
	}
}

// dedupLiterals removes any duplicate Literals before returning the slice.
// As a side-effect, it currently returns the Literals in sorted order.
func dedupLiterals(v []*Literal) ([]*Literal, error) {
	if v == nil {
		return nil, nil
	}

	proxies, err := buildSortProxies(v)
	if err != nil {
		return nil, err
	}

	sort.Sort(byKey(proxies))

	var deduped []*Literal
	for i, p := range proxies {
		if i != 0 && proxies[i-1].key == proxies[i].key {
			continue
		}
		deduped = append(deduped, p.literal)
	}

	return deduped, nil
}
