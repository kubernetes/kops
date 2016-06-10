package terraform

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/hashicorp/hcl/hcl/ast"
	hcl_printer "github.com/hashicorp/hcl/hcl/printer"
	"strings"
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
		glog.Warningf(" unknown type: %T\n", n)
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
					if strings.IndexRune(safeChars, c) == -1 {
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

	return []byte(s), nil
}
