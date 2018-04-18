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

package init_repo

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/util"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:     "repo",
	Short:   "Initialize a repo with the apiserver scaffolding and vendor/ dependencies",
	Long:    `Initialize a repo with the apiserver scaffolding and vendor/ dependencies`,
	Example: `apiserver-boot init repo --domain mydomain`,
	Run:     RunInitRepo,
}

var installDeps bool
var domain string
var copyright string

func AddInitRepo(cmd *cobra.Command) {
	cmd.AddCommand(repoCmd)
	repoCmd.Flags().StringVar(&domain, "domain", "", "domain the api groups live under")

	// Hide this flag by default
	repoCmd.Flags().StringVar(&copyright, "copyright", "boilerplate.go.txt", "Location of copyright boilerplate file.")
	repoCmd.Flags().
		BoolVar(&installDeps, "install-deps", true, "if true, install the vendored deps packaged with apiserver-boot.")
	repoCmd.Flags().
		BoolVar(&Update, "update", false, "if true, don't touch Gopkg.toml or Gopkg.lock, and replace versions of packages managed by apiserver-boot.")
	repoCmd.Flags().MarkHidden("install-deps")
}

func RunInitRepo(cmd *cobra.Command, args []string) {
	if len(domain) == 0 {
		log.Fatal("Must specify --domain")
	}
	cr := util.GetCopyright(copyright)

	createBazelWorkspace()
	createApiserver(cr)
	createControllerManager(cr)
	createAPIs(cr)

	createPackage(cr, filepath.Join("pkg"))
	createPackage(cr, filepath.Join("pkg", "controller"))
	createPackage(cr, filepath.Join("pkg", "controller", "sharedinformers"))
	createPackage(cr, filepath.Join("pkg", "openapi"))

	os.MkdirAll("bin", 0700)

	if installDeps {
		log.Printf("installing vendor/ directory.  To disable this, run with --install-deps=false.")
		RunVendorInstall(nil, nil)
	}
}

func createBazelWorkspace() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "WORKSPACE")
	util.WriteIfNotFound(path, "bazel-workspace-template", workspaceTemplate, nil)
	path = filepath.Join(dir, "BUILD.bazel")
	util.WriteIfNotFound(path, "bazel-build-template",
		buildTemplate, buildTemplateArguments{util.Repo})
}

type controllerManagerTemplateArguments struct {
	BoilerPlate string
	Repo        string
}

var controllerManagerTemplate = `
{{.BoilerPlate}}

package main

import (
	"flag"
	"log"

	controllerlib "github.com/kubernetes-incubator/apiserver-builder/pkg/controller"

	"{{ .Repo }}/pkg/controller"
)

var kubeconfig = flag.String("kubeconfig", "", "path to kubeconfig")

func main() {
	flag.Parse()
	config, err := controllerlib.GetConfig(*kubeconfig)
	if err != nil {
		log.Fatalf("Could not create Config for talking to the apiserver: %v", err)
	}

	controllers, _ := controller.GetAllControllers(config)
	controllerlib.StartControllerManager(controllers...)

	// Blockforever
	select {}
}
`

func createControllerManager(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "cmd", "controller-manager", "main.go")
	util.WriteIfNotFound(path, "main-template", controllerManagerTemplate,
		controllerManagerTemplateArguments{
			boilerplate,
			util.Repo,
		})
}

type apiserverTemplateArguments struct {
	Domain      string
	BoilerPlate string
	Repo        string
}

var apiserverTemplate = `
{{.BoilerPlate}}

package main

import (
	// Make sure dep tools picks up these dependencies
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "github.com/go-openapi/loads"

	"github.com/kubernetes-incubator/apiserver-builder/pkg/cmd/server"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Enable cloud provider auth

	"{{.Repo}}/pkg/apis"
	"{{.Repo}}/pkg/openapi"
)

func main() {
	version := "v0"
	server.StartApiServer("/registry/{{ .Domain }}", apis.GetAllApiBuilders(), openapi.GetOpenAPIDefinitions, "Api", version)
}
`

func createApiserver(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "cmd", "apiserver", "main.go")
	util.WriteIfNotFound(path, "apiserver-template", apiserverTemplate,
		apiserverTemplateArguments{
			domain,
			boilerplate,
			util.Repo,
		})

}

func createPackage(boilerplate, path string) {
	pkg := filepath.Base(path)
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path = filepath.Join(dir, path, "doc.go")
	util.WriteIfNotFound(path, "pkg-template", packageDocTemplate,
		packageDocTemplateArguments{
			boilerplate,
			pkg,
		})
}

type packageDocTemplateArguments struct {
	BoilerPlate string
	Package     string
}

var packageDocTemplate = `
{{.BoilerPlate}}


package {{.Package}}

`

func createAPIs(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "pkg", "apis", "doc.go")
	util.WriteIfNotFound(path, "apis-template", apisDocTemplate,
		apisDocTemplateArguments{
			boilerplate,
			domain,
		})
}

type apisDocTemplateArguments struct {
	BoilerPlate string
	Domain      string
}

var apisDocTemplate = `
{{.BoilerPlate}}


//
// +domain={{.Domain}}

package apis

`

var workspaceTemplate = `
http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.6.0/rules_go-0.6.0.tar.gz",
    sha256 = "ba6feabc94a5d205013e70792accb6cce989169476668fbaf98ea9b342e13b59",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()

load("@io_bazel_rules_go//proto:def.bzl", "proto_register_toolchains")
proto_register_toolchains()
`

type buildTemplateArguments struct {
	Repo string
}

var buildTemplate = `
# gazelle:proto disable
# gazelle:exclude vendor
load("@io_bazel_rules_go//go:def.bzl", "gazelle")

gazelle(
    name = "gazelle",
    command = "fix",
    prefix = "{{.Repo}}",
    external = "vendored",
    args = [
        "-build_file_name",
        "BUILD,BUILD.bazel",
    ],
)
`
