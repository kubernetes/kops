/*
Copyright 2020 The Kubernetes Authors.

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
	setInstancegroupLong = templates.LongDesc(i18n.T(`Set an instance group field value.

		This command changes the desired instance group configuration in the registry.

        kops set does not update the cloud resources, to apply the changes use "kops update cluster".
		
		Valid Instance Group Settings:

		%s`))

	setInstancegroupExample = templates.Examples(i18n.T(`
		# Set instance group to run image custom-ami-image
		kops set instancegroup --name k8s-cluster.example.com nodes spec.image=custom-ami-image
	`))

	setInstancegroupShort = i18n.T("Set instancegroup fields.")
)

// NewCmdSetInstancegroup builds a cobra command for the kops set instancegroup command
func NewCmdSetInstancegroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &commands.SetOptions{}
	kts := commands.ValidInstanceGroupKeysToSetters()

	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   setInstancegroupShort,
		Long:    fmt.Sprintf(setInstancegroupLong, kts.PrettyPrintKeysWithNewlines()),
		Example: setInstancegroupExample,
		Run: func(cmd *cobra.Command, args []string) {
			for i, arg := range args {
				index := strings.Index(arg, "=")

				if i == 0 {
					if index != -1 {
						exitWithError(fmt.Errorf("Specify name of instance group to edit"))
					}
					options.InstanceGroupName = arg
				} else {
					if index == -1 {
						exitWithError(fmt.Errorf("unrecognized parameter %q (missing '=')", arg))
						return
					}
					options.Fields = append(options.Fields, arg)
				}
			}

			options.ClusterName = rootCommand.ClusterName()

			if err := commands.RunSetInstancegroup(f, cmd, out, options); err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}
