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
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	versionLong = templates.LongDesc(i18n.T(`
	Print the kops version and git SHA.`))

	versionExample = templates.Examples(i18n.T(`
	kops version`))

	versionShort = i18n.T(`Print the kops version information.`)
)

// NewCmdVersion builds a cobra command for the kops version command
func NewCmdVersion(f *util.Factory, out io.Writer) *cobra.Command {
	options := &commands.VersionOptions{}

	cmd := &cobra.Command{
		Use:     "version",
		Short:   versionShort,
		Long:    versionLong,
		Example: versionExample,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := commands.RunVersion(f, out, options)
		if err != nil {
			exitWithError(err)
		}
	}

	cmd.Flags().BoolVar(&options.Short, "short", options.Short, "only print the main kops version, useful for scripting")

	return cmd
}
