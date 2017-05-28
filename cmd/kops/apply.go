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
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"

	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	apply_long = templates.LongDesc(i18n.T(`
		Apply a cluster resource specification by filename or stdin.`))

	apply_example = templates.Examples(i18n.T(`
		# Apply a cluster resource specification using a file.
		kops apply -f my-cluster.yaml
		`))

	apply_short = i18n.T(`Apply a cluster resource specification.`)
)

func NewCmdApply(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ReplaceOptions{}

	cmd := &cobra.Command{
		Use:     "apply -f FILENAME",
		Short:   apply_short,
		Long:    apply_long,
		Example: apply_example,
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameEmpty(options.Filenames) {
				cmd.Help()
				return
			}

			cmdutil.CheckErr(RunReplace(f, cmd, out, options))
		},
	}

	ConfigureReplaceCmd(cmd, options)

	return cmd
}
