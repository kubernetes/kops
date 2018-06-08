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

package create

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/util"
)

var subresourceName string

var createSubresourceCmd = &cobra.Command{
	Use:   "subresource",
	Short: "Creates a subresource",
	Long:  `Creates a subresource.  Creates file pkg/apis/<group>/<version>/<subresourceName>_<kind>_types.go and updates pkg/apis/<group>/<version>/<kind>_types.go with the subresource comment directive.`,
	Example: `# Create new subresource "pollinate" of resource "Bee" in the "insect" group with version "v1beta"
apiserver-boot create subresource --subresource pollinate --group insect --version v1beta --kind Bee`,
	Run: RunCreateSubresource,
}

func AddCreateSubresource(cmd *cobra.Command) {
	RegisterResourceFlags(createSubresourceCmd)

	createSubresourceCmd.Flags().StringVar(&subresourceName, "subresource", "", "name of the subresource.  **Must be single lowercase word**")

	cmd.AddCommand(createSubresourceCmd)
}

func RunCreateSubresource(cmd *cobra.Command, args []string) {
	if _, err := os.Stat("pkg"); err != nil {
		log.Fatalf("could not find 'pkg' directory.  must run apiserver-boot init before creating resources")
	}

	util.GetDomain()
	if len(subresourceName) == 0 {
		log.Fatalf("Must specify --subresource")
	}
	ValidateResourceFlags()

	subresourceMatch := regexp.MustCompile("^[a-z]+$")
	if !subresourceMatch.MatchString(subresourceName) {
		log.Fatalf("--subresource must match regex ^[a-z]+$ but was (%s)", subresourceName)
	}

	cr := util.GetCopyright(copyright)

	ignoreGroupExists = true
	createGroup(cr)
	ignoreVersionExists = true
	createVersion(cr)

	createSubresource(cr)
}

func createSubresource(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	a := subresourceTemplateArgs{
		boilerplate,
		subresourceName,
		strings.Title(subresourceName) + strings.Title(kindName),
		util.Repo,
		groupName,
		versionName,
		kindName,
		resourceName,
	}

	found := false

	typesFileName := fmt.Sprintf("%s_%s_types.go", strings.ToLower(subresourceName), strings.ToLower(kindName))
	path := filepath.Join(dir, "pkg", "apis", groupName, versionName, typesFileName)
	created := util.WriteIfNotFound(path, "sub-resource-template", subresourceTemplate, a)
	if !created {
		if !found {
			log.Printf("API subresourceName %s for group version kind %s/%s/%s already exists.",
				subresourceName, groupName, versionName, kindName)
			found = true
		}
	}

	typesFileName = fmt.Sprintf("%s_%s_types_test.go", strings.ToLower(subresourceName), strings.ToLower(kindName))
	path = filepath.Join(dir, "pkg", "apis", groupName, versionName, typesFileName)
	created = util.WriteIfNotFound(path, "sub-resource-test-template", subresourceTestTemplate, a)
	if !created {
		if !found {
			log.Printf("API subresourceName %s for group version kind %s/%s/%s already exists.",
				subresourceName, groupName, versionName, kindName)
			found = true
		}
	}

	if !found {
		typesFileName = fmt.Sprintf("%s_types.go", strings.ToLower(kindName))
		path = filepath.Join(dir, "pkg", "apis", groupName, versionName, typesFileName)
		types, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		structName := fmt.Sprintf("type %s struct {", kindName)
		sub := fmt.Sprintf("// +subresource:request=%s,path=%s,rest=%s%sREST",
			strings.Title(subresourceName),
			strings.ToLower(subresourceName),
			strings.Title(subresourceName),
			kindName)
		result := strings.Replace(string(types),
			structName,
			sub+"\n"+structName, 1)
		ioutil.WriteFile(path, []byte(result), 0644)
	}

	if found {
		os.Exit(-1)
	}
}

type subresourceTemplateArgs struct {
	BoilerPlate     string
	Subresource     string
	SubresourceKind string
	Repo            string
	Group           string
	Version         string
	Kind            string
	Resource        string
}

var subresourceTemplate = `
{{.BoilerPlate}}

package {{.Version}}

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"{{ .Repo }}/pkg/apis/{{ .Group }}"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +subresource-request
type {{title .Subresource}} struct {
	metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
	metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
}

var _ rest.CreaterUpdater = &{{ .SubresourceKind }}REST{}
var _ rest.Patcher = &{{ .SubresourceKind }}REST{}

// +k8s:deepcopy-gen=false
type {{ .SubresourceKind }}REST struct {
	Registry {{ .Group }}.{{ .Kind }}Registry
}

func (r *{{ .SubresourceKind }}REST) Create(ctx request.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, includeUninitialized bool) (runtime.Object, error) {
	sub := obj.(*{{ title .Subresource }})
	rec, err := r.Registry.Get{{ title .Kind }}(ctx, sub.Name, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// Modify rec in someway before writing it back to storage

	r.Registry.Update{{ title .Kind }}(ctx, rec)
	return rec, nil
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *{{ .SubresourceKind }}REST) Get(ctx request.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return nil, nil
}

// Update alters the status subset of an object.
func (r *{{ .SubresourceKind }}REST) Update(ctx request.Context, name string, objInfo rest.UpdatedObjectInfo, createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc) (runtime.Object, bool, error) {
	return nil, false, nil
}

func (r *{{ .SubresourceKind }}REST) New() runtime.Object {
	return &{{ title .Subresource }}{}
}

`

var subresourceTestTemplate = `
{{.BoilerPlate}}

package {{.Version}}_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "{{.Repo}}/pkg/apis/{{.Group}}/{{.Version}}"
	. "{{.Repo}}/pkg/client/clientset_generated/clientset/typed/{{.Group}}/{{.Version}}"
)

var _ = Describe("{{.Kind}}", func() {
	var instance {{ .Kind}}
	var expected {{ .Kind}}
	var client {{ .Kind}}Interface

	BeforeEach(func() {
		instance = {{ .Kind}}{}
		instance.Name = "instance-1"

		expected = instance
	})

	AfterEach(func() {
		client.Delete(instance.Name, &metav1.DeleteOptions{})
	})

	Describe("when sending a {{ .Subresource }} request", func() {
		It("should return success", func() {
			client = cs.{{ title .Group }}{{ title .Version }}Client.{{ title .Resource }}("{{ lower .Kind }}-test-{{ lower .Subresource }}")
			_, err := client.Create(&instance)
			Expect(err).ShouldNot(HaveOccurred())

			{{ lower .Subresource }} := &{{ title .Subresource}}{}
			{{ lower .Subresource }}.Name = instance.Name
			restClient := cs.{{ title .Group }}{{ title .Version }}Client.RESTClient()
			err = restClient.Post().Namespace("{{ lower .Kind }}-test-{{ lower .Subresource}}").
				Name(instance.Name).
				Resource("{{ lower .Resource }}").
				SubResource("{{ lower .Subresource }}").
				Body({{ lower .Subresource }}).Do().Error()
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
`
