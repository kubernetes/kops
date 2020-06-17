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

package reflectutils

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"

	"k8s.io/klog/v2"
)

type FieldPath struct {
	elements []FieldPathElement
}

type FieldPathElementType int

const (
	FieldPathElementTypeField = iota
	FieldPathElementTypeMapKey
	FieldPathElementTypeArrayIndex
	FieldPathElementTypeWildcardIndex
)

type FieldPathElement struct {
	Type   FieldPathElementType
	token  string
	number int
}

func (f *FieldPath) Extend(el FieldPathElement) *FieldPath {
	n := len(f.elements)
	newElements := make([]FieldPathElement, n+1)
	copy(newElements, f.elements)
	newElements[n] = el
	return &FieldPath{elements: newElements}
}

func (f *FieldPath) String() string {
	var sb strings.Builder
	for i, el := range f.elements {
		switch el.Type {
		case FieldPathElementTypeField:
			if i != 0 {
				sb.WriteString(".")
			}
			sb.WriteString(el.token)
		case FieldPathElementTypeMapKey:
			sb.WriteString("[")
			sb.WriteString(el.token)
			sb.WriteString("]")
		case FieldPathElementTypeWildcardIndex:
			sb.WriteString("[*]")
		case FieldPathElementTypeArrayIndex:
			sb.WriteString("[")
			sb.WriteString(strconv.Itoa(el.number))
			sb.WriteString("]")
		default:
			klog.Fatalf("unknown token type %+v", el)
		}
	}
	return sb.String()
}

func ParseFieldPath(s string) (*FieldPath, error) {
	var elements []FieldPathElement

	var scan scanner.Scanner

	scan.Init(strings.NewReader(s))
	scan.Mode ^= scanner.SkipComments // don't skip comments

	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
		switch tok {
		case scanner.Ident:
			field := scan.TokenText()
			elements = append(elements, FieldPathElement{
				Type:  FieldPathElementTypeField,
				token: field,
			})

		case '.':
			// TODO: Validate that we don't have two dots?
			// Skip

		case '[':
			{
				tok := scan.Scan()
				switch tok {
				case '*':
					elements = append(elements, FieldPathElement{
						Type: FieldPathElementTypeWildcardIndex,
					})

				case scanner.Int:
					v := scan.TokenText()
					n, err := strconv.Atoi(v)
					if err != nil {
						return nil, fmt.Errorf("cannot parse token %q as array-index", v)
					}
					elements = append(elements, FieldPathElement{
						Type:   FieldPathElementTypeArrayIndex,
						number: n,
					})

				default:
					return nil, fmt.Errorf("unexpected token %v (%s)", tok, scan.TokenText())
				}

				tok = scan.Scan()
				switch tok {
				case ']':
					// ok
				default:
					return nil, fmt.Errorf("unexpected token %v (%s)", tok, scan.TokenText())
				}

			}

		default:
			return nil, fmt.Errorf("unexpected token %v (%s)", tok, scan.TokenText())
		}
	}

	return &FieldPath{elements: elements}, nil
}

func (p *FieldPath) IsEmpty() bool {
	return len(p.elements) == 0
}

func (p *FieldPath) Matches(r *FieldPath) bool {
	if len(p.elements) != len(r.elements) {
		return false
	}
	return p.HasPrefixMatch(r)
}

func (p *FieldPath) HasPrefixMatch(r *FieldPath) bool {
	if len(p.elements) < len(r.elements) {
		return false
	}
	for i := range r.elements {
		if p.elements[i] == r.elements[i] {
			continue
		}
		isMatch := false

		matcher := p.elements[i]
		target := r.elements[i]
		switch matcher.Type {
		case FieldPathElementTypeWildcardIndex:
			if target.Type == FieldPathElementTypeArrayIndex {
				isMatch = true
			}
		}
		if !isMatch {
			return false
		}
	}
	return true
}
