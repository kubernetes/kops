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
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	toolboxLong = templates.LongDesc(i18n.T(`
	Misc infrequently used commands.`))

	toolboxExample = templates.Examples(i18n.T(`
	# Dump cluster information
	kops toolbox dump --name k8s-cluster.example.com
	`))

	toolboxShort = i18n.T(`Misc infrequently used commands.`)
)

func NewCmdToolbox(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "toolbox",
		Short:   toolboxShort,
		Long:    toolboxLong,
		Example: toolboxExample,
	}

	cmd.AddCommand(NewCmdToolboxConvertImported(f, out))
	cmd.AddCommand(NewCmdToolboxDump(f, out))
	cmd.AddCommand(NewCmdToolboxBundle(f, out))
	cmd.AddCommand(NewCmdToolboxTemplate(f, out))

	return cmd
}
