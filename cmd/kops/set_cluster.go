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
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands"
)

var (
	setClusterLong = templates.LongDesc(i18n.T(`Set a cluster field value.

        This command changes the desired cluster configuration in the registry.

        kops set does not update the cloud resources, to apply the changes use "kops update cluster".`))

	setClusterExample = templates.Examples(i18n.T(`
		# Set cluster to run kubernetes version 1.10.0
		kops set cluster k8s.cluster.site spec.kubernetesVersion=1.10.0
	`))
)

// NewCmdSetCluster builds a cobra command for the kops set cluster command
func NewCmdSetCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &commands.SetOptions{}

	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   i18n.T("Set cluster fields."),
		Long:    setClusterLong,
		Example: setClusterExample,
		Run: func(cmd *cobra.Command, args []string) {
			for i, arg := range args {
				index := strings.Index(arg, "=")

				if i == 0 && index == -1 {
					options.ClusterName = arg
				} else {
					if index == -1 {
						exitWithError(fmt.Errorf("unrecognized parameter %q (missing '=')", arg))
						return
					}
					options.Fields = append(options.Fields, arg)
				}
			}

			if options.ClusterName == "" {
				options.ClusterName = rootCommand.ClusterName()
			}

			if err := commands.RunSetCluster(f, cmd, out, options); err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}
