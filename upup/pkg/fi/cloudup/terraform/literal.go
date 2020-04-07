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

package terraform

import (
	"encoding/json"
	"fmt"
	"sort"

	"k8s.io/klog"
)

// Literal represents a literal in terraform syntax
type Literal struct {
	// Value is only used in terraform 0.11 and represents the literal string to use as a value.
	// "${}" interpolation is supported.
	Value string `cty:"value"`

	// The below fields are only used in terraform 0.12.
	// ResourceType represents the type of a resource in a literal reference
	ResourceType string `cty:"resource_type"`
	// ResourceName represents the name of a resource in a literal reference
	ResourceName string `cty:"resource_name"`
	// ResourceProp represents the property of a resource in a literal reference
	ResourceProp string `cty:"resource_prop"`
	// FilePath represents the path for a file() reference
	FilePath string `cty:"file_path"`
}

var _ json.Marshaler = &Literal{}

func (l *Literal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&l.Value)
}

func LiteralExpression(s string) *Literal {
	return &Literal{Value: s}
}

func LiteralFileExpression(modulePath string) *Literal {
	return &Literal{
		Value:    fmt.Sprintf("${file(%q)}", modulePath),
		FilePath: modulePath,
	}
}

func LiteralSelfLink(resourceType, resourceName string) *Literal {
	return LiteralProperty(resourceType, resourceName, "self_link")
}

func LiteralProperty(resourceType, resourceName, prop string) *Literal {
	tfName := tfSanitize(resourceName)
	expr := "${" + resourceType + "." + tfName + "." + prop + "}"
	return &Literal{
		Value:        expr,
		ResourceType: resourceType,
		ResourceName: tfName,
		ResourceProp: prop,
	}
}

func LiteralFromStringValue(s string) *Literal {
	return &Literal{Value: s}
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

// DedupLiterals removes any duplicate Literals before returning the slice.
// As a side-effect, it currently returns the Literals in sorted order.
func DedupLiterals(v []*Literal) ([]*Literal, error) {
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
