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
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	validateLong = templates.LongDesc(i18n.T(`
	This command validates a cluster.
	See:
	kops validate cluster -h
	`))

	validateExample = templates.Examples(i18n.T(`
	# Validate the cluster set as the current context of the kube config.
	# Kops will try for 10 minutes to validate the cluster 3 times.
	kops validate cluster --wait 10m --count 3`))

	validateShort = i18n.T(`Validate a kops cluster.`)
)

func NewCmdValidate(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate",
		Short:   validateShort,
		Long:    validateLong,
		Example: validateExample,
	}

	// create subcommands
	cmd.AddCommand(NewCmdValidateCluster(f, out))

	return cmd
}
