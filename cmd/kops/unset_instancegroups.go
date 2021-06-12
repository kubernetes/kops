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
	unsetInstancegroupLong = templates.LongDesc(i18n.T(`Unset an instance group field value.

		This command changes the desired instance group configuration in the registry.

        kops unset does not update the cloud resources; to apply the changes use "kops update cluster".`))

	unsetInstancegroupExample = templates.Examples(i18n.T(`
		# Set instance group to run default image
		kops unset instancegroup --name k8s-cluster.example.com nodes spec.image
	`))
)

// NewCmdUnsetInstancegroup builds a cobra command for the kops set instancegroup command.
func NewCmdUnsetInstancegroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &commands.UnsetInstanceGroupOptions{}

	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   i18n.T("Unset instancegroup fields."),
		Long:    unsetInstancegroupLong,
		Example: unsetInstancegroupExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			for i, arg := range args {
				if i == 0 && !strings.HasPrefix(arg, "spec.") && !strings.HasPrefix(arg, "instancegroup.") {
					options.InstanceGroupName = arg
				} else {
					options.Fields = append(options.Fields, arg)
				}
			}

			options.ClusterName = rootCommand.ClusterName()

			if err := commands.RunUnsetInstancegroup(ctx, f, cmd, out, options); err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}
