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
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	validate_long = templates.LongDesc(i18n.T(`
	This commands validates the following components:

	1. All k8s masters are running and have "Ready" status.
	2. All k8s nodes are running and have "Ready" status.
	3. Componentstatues returns healthly for all components.
	4. All pods in the kube-system namespace are running and healthy.
	`))

	validate_example = templates.Examples(i18n.T(`
	# Validate a cluster.
	# This command uses the currently selected kops cluster as
	# set by the kubectl config.
	kubernetes validate cluster`))

	validate_short = i18n.T(`Validate a kops cluster.`)
)

func NewCmdValidate(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate",
		Short:   validate_short,
		Long:    validate_long,
		Example: validate_example,
	}

	// create subcommands
	cmd.AddCommand(NewCmdValidateCluster(f, out))

	return cmd
}
