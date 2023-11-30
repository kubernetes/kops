/*
Copyright 2021 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

func NewCmdToolboxEnroll(f commandutils.Factory, out io.Writer) *cobra.Command {
	options := &commands.ToolboxEnrollOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:   "enroll [CLUSTER]",
		Short: i18n.T(`Add machine to cluster`),
		Long: templates.LongDesc(i18n.T(`
			Adds an individual machine to the cluster.`)),
		Example: templates.Examples(i18n.T(`
			kops toolbox enroll --name k8s-cluster.example.com
		`)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.RunToolboxEnroll(cmd.Context(), f, out, options)
		},
	}

	cmd.Flags().StringVar(&options.ClusterName, "cluster", options.ClusterName, "Name of cluster to join")
	cmd.Flags().StringVar(&options.InstanceGroup, "instance-group", options.InstanceGroup, "Name of instance-group to join")

	cmd.Flags().StringVar(&options.Host, "host", options.Host, "IP/hostname for machine to add")
	cmd.Flags().StringVar(&options.SSHUser, "ssh-user", options.SSHUser, "user for ssh")
	cmd.Flags().IntVar(&options.SSHPort, "ssh-port", options.SSHPort, "port for ssh")

	return cmd
}
