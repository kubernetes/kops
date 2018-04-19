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
	"log"
	"os"
	"path/filepath"

	"strings"

	"github.com/kubernetes-incubator/apiserver-builder/cmd/apiserver-boot/boot/util"
	"github.com/spf13/cobra"
)

var createGroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Creates an API group",
	Long:  `Creates an API group.`,
	Run:   RunCreateGroup,
}

var groupName string
var ignoreGroupExists bool = false

func AddCreateGroup(cmd *cobra.Command) {
	createGroupCmd.Flags().StringVar(&groupName, "group", "", "name of the API group to create")

	cmd.AddCommand(createGroupCmd)
	createGroupCmd.AddCommand(createVersionCmd)
}

func RunCreateGroup(cmd *cobra.Command, args []string) {
	if _, err := os.Stat("pkg"); err != nil {
		log.Fatalf("could not find 'pkg' directory.  must run apiserver-boot init before creating resources")
	}

	util.GetDomain()
	if len(groupName) == 0 {
		log.Fatalf("Must specify --group")
	}

	if strings.ToLower(groupName) != groupName {
		log.Fatalf("--group must be lowercase was (%s)", groupName)
	}

	createGroup(util.GetCopyright(copyright))
}

func createGroup(boilerplate string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	a := groupTemplateArgs{
		boilerplate,
		util.Domain,
		groupName,
	}

	path := filepath.Join(dir, "pkg", "apis", groupName, "doc.go")
	created := util.WriteIfNotFound(path, "group-template", groupTemplate, a)

	path = filepath.Join(dir, "pkg", "apis", groupName, "install", "doc.go")
	created = util.WriteIfNotFound(path, "install-template", installTemplate, a)
	if !created && !ignoreGroupExists {
		log.Fatalf("API group %s already exists.", groupName)
	}
}

type groupTemplateArgs struct {
	BoilerPlate string
	Domain      string
	Name        string
}

var groupTemplate = `
{{.BoilerPlate}}


// +k8s:deepcopy-gen=package,register
// +groupName={{.Name}}.{{.Domain}}

// Package api is the internal version of the API.
package {{.Name}}

`

var installTemplate = `
{{.BoilerPlate}}

package install

`
