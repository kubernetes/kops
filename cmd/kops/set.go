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
	setLong = templates.LongDesc(i18n.T(`Set a configuration field.

        kops set does not update the cloud resources; to apply the changes use "kops update cluster".
    `))

	setExample = templates.Examples(i18n.T(`
    # Set cluster to run kubernetes version 1.10.0
    kops set cluster k8s-cluster.example.com spec.kubernetesVersion=1.10.0
	`))
)

func NewCmdSet(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set",
		Short:   i18n.T("Set fields on clusters and other resources."),
		Long:    setLong,
		Example: setExample,
	}

	// create subcommands
	cmd.AddCommand(NewCmdSetCluster(f, out))
	cmd.AddCommand(NewCmdSetInstancegroup(f, out))

	return cmd
}
