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

package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/gengo/args"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
	"k8s.io/klog/v2"
)

// These are the comment tags that carry parameters for fitask generation.
const tagName = "kops:fitask"

func extractTag(comments []string) []string {
	return types.ExtractCommentTags("+", comments)[tagName]
}

const perTypeDef = `
// {{.Name}}

// JSON marshaling boilerplate
type real{{.Name}} {{.Name}}

// UnmarshalJSON implements conversion to JSON, supporting an alternate specification of the object as a string
func (o *{{.Name}}) UnmarshalJSON(data []byte) error {
	var jsonName string
	if err := json.Unmarshal(data, &jsonName); err == nil {
		o.Name = &jsonName
		return nil
	}

	var r real{{.Name}}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	*o = {{.Name}}(r)
	return nil
}

var _ fi.HasLifecycle = &{{.Name}}{}

// GetLifecycle returns the Lifecycle of the object, implementing fi.HasLifecycle
func (o *{{.Name}}) GetLifecycle() *fi.Lifecycle {
	return o.Lifecycle
}

// SetLifecycle sets the Lifecycle of the object, implementing fi.SetLifecycle
func (o *{{.Name}}) SetLifecycle(lifecycle fi.Lifecycle) {
	o.Lifecycle = &lifecycle
}

var _ fi.HasName = &{{.Name}}{}

// GetName returns the Name of the object, implementing fi.HasName
func (o *{{.Name}}) GetName() *string {
	return o.Name
}

// SetName sets the Name of the object, implementing fi.SetName
func (o *{{.Name}}) SetName(name string) {
	o.Name = &name
}

// String is the stringer function for the task, producing readable output using fi.TaskAsString
func (o *{{.Name}}) String() string {
	return fi.TaskAsString(o)
}
`

// NameSystems returns the name system used by the generators in this package.
func NameSystems() namer.NameSystems {
	return namer.NameSystems{
		"public":  namer.NewPublicNamer(0),
		"private": namer.NewPrivateNamer(0),
		"raw":     namer.NewRawNamer("", nil),
	}
}

// DefaultNameSystem returns the default name system for ordering the types to be
// processed by the generators in this package.
func DefaultNameSystem() string {
	return "public"
}

// Packages makes the sets package definition.
func Packages(context *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	boilerplate, err := arguments.LoadGoBoilerplate()
	if err != nil {
		klog.Fatalf("Failed loading boilerplate: %v", err)
	}

	inputs := sets.NewString(context.Inputs...)
	packages := generator.Packages{}
	header := append([]byte(fmt.Sprintf("// +build !%s\n\n", arguments.GeneratedBuildTag)), boilerplate...)

	for i := range inputs {
		klog.V(5).Infof("considering pkg %q", i)
		pkg := context.Universe[i]
		if pkg == nil {
			// If the input had no Go files, for example.
			continue
		}

		fitasks := map[*types.Type]bool{}
		for _, t := range pkg.Types {
			if t.Kind == types.Struct && len(extractTag(t.CommentLines)) > 0 {
				fitasks[t] = true
			}
		}

		packages = append(packages, &generator.DefaultPackage{
			PackageName: filepath.Base(pkg.Path),
			PackagePath: strings.TrimPrefix(pkg.Path, "k8s.io/kops/"),
			HeaderText:  header,
			GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
				for t := range fitasks {
					generators = append(generators, NewGenFitask(t))
				}
				return generators
			},
			FilterFunc: func(c *generator.Context, t *types.Type) bool {
				return fitasks[t]
			},
		})
	}

	return packages
}

type genFitask struct {
	generator.DefaultGen
	typeToMatch *types.Type
}

func NewGenFitask(t *types.Type) generator.Generator {
	return &genFitask{
		DefaultGen: generator.DefaultGen{
			OptionalName: strings.ToLower(t.Name.Name) + "_fitask",
		},
		typeToMatch: t,
	}
}

// Filter ignores all but one type because we're making a single file per type.
func (g *genFitask) Filter(c *generator.Context, t *types.Type) bool { return t == g.typeToMatch }

func (g *genFitask) Imports(c *generator.Context) (imports []string) {
	return []string{
		"encoding/json",
		"k8s.io/kops/upup/pkg/fi",
	}
}

type TypeData struct {
	Name string
}

func (g *genFitask) GenerateType(_ *generator.Context, t *types.Type, w io.Writer) error {
	tmpl := template.Must(template.New("PerType").Parse(perTypeDef))

	d := &TypeData{}
	d.Name = t.Name.Name

	return tmpl.Execute(w, d)
}
