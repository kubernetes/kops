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

package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/build"
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/create"
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/init_repo"
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/run"
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/update"
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/util"
	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/version"
)

func main() {
	util.CheckInstall()
	gopath := os.Getenv("GOPATH")
	if len(gopath) == 0 {
		log.Fatal("GOPATH not defined")
	}
	util.GoSrc = filepath.Join(gopath, "src")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if !strings.HasPrefix(filepath.Dir(wd), util.GoSrc) {
		log.Fatalf("apiserver-boot must be run from the directory containing the go package to "+
			"bootstrap. This must be under $GOPATH/src/<package>. "+
			"\nCurrent GOPATH=%s.  \nCurrent directory=%s", gopath, wd)
	}
	util.Repo = strings.Replace(wd, util.GoSrc+string(filepath.Separator), "", 1)

	init_repo.AddInit(cmd)
	create.AddCreate(cmd)
	build.AddBuild(cmd)
	run.AddRun(cmd)
	update.AddUpdate(cmd)
	version.AddVersion(cmd)

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var cmd = &cobra.Command{
	Use:   "apiserver-boot",
	Short: "apiserver-boot development kit for building Kubernetes extensions in go.",
	Long:  `apiserver-boot development kit for building Kubernetes extensions in go.`,
	Example: `# Initialize your repository with scaffolding directories and go files.
apiserver-boot init repo --domain example.com

# Create new resource "Bee" in the "insect" group with version "v1beta"
apiserver-boot create group version kind --group insect --version v1beta --kind Bee

# Build the generated code, apiserver and controller-manager so they be run locally.
apiserver-boot build executables

# Run the tests that were created for your resources
# Requires generated code was already built by "build executables" or "build generated"
go test ./pkg/...

# Run locally by starting a local etcd, apiserver and controller-manager
# Produces a kubeconfig to talk to the local server
apiserver-boot run local

# Check the api versions of the locally running server
kubectl --kubeconfig kubeconfig api-versions

# Build an image and run in a cluster in the default namespace
# Note: after running this you should clear the discovery service
# cache before running kubectl with "rm -rf ~/.kube/cache/discovery/"
apiserver-boot run in-cluster --name creatures --namespace default --image repo/name:tag`,
	Run: RunMain,
}

func RunMain(cmd *cobra.Command, args []string) {
	cmd.Help()
}
