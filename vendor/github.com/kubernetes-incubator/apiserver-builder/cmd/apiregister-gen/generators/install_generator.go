/*
Copyright 2017 The Kubernetes Authors.

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

package generators

import (
	"io"
	"text/template"

	"path"

	"k8s.io/gengo/generator"
)

type installGenerator struct {
	generator.DefaultGen
	apigroup *APIGroup
}

var _ generator.Generator = &unversionedGenerator{}

func CreateInstallGenerator(apigroup *APIGroup, filename string) generator.Generator {
	return &installGenerator{
		generator.DefaultGen{OptionalName: filename},
		apigroup,
	}
}

func (d *installGenerator) Imports(c *generator.Context) []string {
	return []string{
		"k8s.io/apimachinery/pkg/apimachinery/announced",
		"k8s.io/apimachinery/pkg/apimachinery/registered",
		"k8s.io/apimachinery/pkg/runtime",
		path.Dir(d.apigroup.Pkg.Path),
	}
}

func (d *installGenerator) Finalize(context *generator.Context, w io.Writer) error {
	temp := template.Must(template.New("install-template").Parse(InstallAPITemplate))
	err := temp.Execute(w, d.apigroup)
	if err != nil {
		return err
	}
	return err
}

var InstallAPITemplate = `
func Install(
	groupFactoryRegistry announced.APIGroupFactoryRegistry,
	registry *registered.APIRegistrationManager,
	scheme *runtime.Scheme) {

	apis.Get{{ .GroupTitle }}APIBuilder().Install(groupFactoryRegistry, registry, scheme)
}
`
