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
	"context"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands"
)

var (
	unsetClusterLong = templates.LongDesc(i18n.T(`Unset a cluster field value.

        This command changes the desired cluster configuration in the registry.

        kops unset does not update the cloud resources; to apply the changes use "kops update cluster".`))

	unsetClusterExample = templates.Examples(i18n.T(`
	    kops unset cluster k8s-cluster.example.com spec.iam.allowContainerRegistry
	`))
)

// NewCmdUnsetCluster builds a cobra command for the kops unset cluster command
func NewCmdUnsetCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &commands.UnsetClusterOptions{}

	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   i18n.T("Unset cluster fields."),
		Long:    unsetClusterLong,
		Example: unsetClusterExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			for i, arg := range args {
				if i == 0 && !strings.HasPrefix(arg, "spec.") && !strings.HasPrefix(arg, "cluster.") {
					options.ClusterName = arg
				} else {
					options.Fields = append(options.Fields, arg)
				}
			}

			if options.ClusterName == "" {
				options.ClusterName = rootCommand.ClusterName(true)
			}

			if err := commands.RunUnsetCluster(ctx, f, cmd, out, options); err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}
