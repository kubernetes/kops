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
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
	hcl_printer "github.com/hashicorp/hcl/hcl/printer"
	"k8s.io/klog"
	"k8s.io/kops/pkg/featureflag"
)

const safeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"

// sanitizer fixes up an invalid HCL AST, as produced by the HCL parser for JSON
type astSanitizer struct {
}

// output prints creates b printable HCL output and returns it.
func (v *astSanitizer) visit(n interface{}) {
	switch t := n.(type) {
	case *ast.File:
		v.visit(t.Node)
	case *ast.ObjectList:
		var index int
		for {
			if index == len(t.Items) {
				break
			}

			v.visit(t.Items[index])
			index++
		}
	case *ast.ObjectKey:
	case *ast.ObjectItem:
		v.visitObjectItem(t)
	case *ast.LiteralType:
	case *ast.ListType:
	case *ast.ObjectType:
		v.visit(t.List)
	default:
		klog.Warningf(" unknown type: %T\n", n)
	}

}

func (v *astSanitizer) visitObjectItem(o *ast.ObjectItem) {
	for i, k := range o.Keys {
		if i == 0 {
			text := k.Token.Text
			if text != "" && text[0] == '"' && text[len(text)-1] == '"' {
				v := text[1 : len(text)-1]
				safe := true
				for _, c := range v {
					if !strings.ContainsRune(safeChars, c) {
						safe = false
						break
					}
				}
				if safe {
					k.Token.Text = v
				}
			}

		}
	}

	// A hack so that Assign.IsValid is true, so that the printer will output =
	o.Assign.Line = 1

	v.visit(o.Val)
}

func hclPrint(node ast.Node) ([]byte, error) {
	var sanitizer astSanitizer
	sanitizer.visit(node)

	var b bytes.Buffer
	err := hcl_printer.Fprint(&b, node)
	if err != nil {
		return nil, fmt.Errorf("error writing HCL: %v", err)
	}
	s := b.String()

	// Remove extra whitespace...
	s = strings.Replace(s, "\n\n", "\n", -1)

	// ...but leave whitespace between resources
	s = strings.Replace(s, "}\nresource", "}\n\nresource", -1)

	// Workaround HCL insanity #6359: quotes are _not_ escaped in quotes (huh?)
	// This hits the file function
	s = strings.Replace(s, "(\\\"", "(\"", -1)
	s = strings.Replace(s, "\\\")", "\")", -1)

	// We don't need to escape > or <
	s = strings.Replace(s, "\\u003c", "<", -1)
	s = strings.Replace(s, "\\u003e", ">", -1)

	if featureflag.SkipTerraformFormat.Enabled() {
		klog.Infof("feature-flag SkipTerraformFormat was set; skipping terraform format")
		return []byte(s), nil
	}

	// Apply Terraform style (alignment etc.)
	formatted, err := hcl_printer.Format([]byte(s))
	if err != nil {
		klog.Errorf("Invalid HCL follows:")
		for i, line := range strings.Split(s, "\n") {
			klog.Errorf("%d\t%s", (i + 1), line)
		}
		return nil, fmt.Errorf("error formatting HCL: %v", err)
	}

	return formatted, nil
}
