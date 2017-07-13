/*
Copyright 2016 The Kubernetes Authors.

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

	"github.com/spf13/cobra"
	"k8s.io/kops"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	version_long = templates.LongDesc(i18n.T(`
	Print the kops version and git SHA.`))

	version_example = templates.Examples(i18n.T(`
	kops version`))

	version_short = i18n.T(`Print the kops version information.`)
)

type VersionCmd struct {
	cobraCommand *cobra.Command
}

var versionCmd = VersionCmd{
	cobraCommand: &cobra.Command{
		Use:     "version",
		Short:   version_short,
		Long:    version_long,
		Example: version_example,
	},
}

func init() {
	cmd := versionCmd.cobraCommand
	rootCommand.cobraCommand.AddCommand(cmd)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := versionCmd.Run()
		if err != nil {
			exitWithError(err)
		}
	}
}

func (c *VersionCmd) Run() error {
	s := "Version " + kops.Version
	if kops.GitVersion != "" {
		s += " (git-" + kops.GitVersion + ")"
	}
	fmt.Println(s)

	return nil
}
