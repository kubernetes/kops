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
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kubernetes/pkg/util/i18n"
)

type CreateClusterAzureOptions struct {
	// Inheritance in Go
	CreateClusterOptions
}

func NewCmdCreateClusterAzure(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateClusterAzureOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "azure",
		Short:   i18n.T("Create a Kubernetes cluster in Azure"),
		Long:    create_cluster_long,
		Example: create_cluster_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
				return
			}
			options.ClusterName = rootCommand.clusterName
			err = RunCreateClusterAzure(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}
	return cmd
}

func RunCreateClusterAzure(f *util.Factory, out io.Writer, c *CreateClusterAzureOptions) error {

	// All kinds of wonderful logic that only happens for azure clusters only

	return c.RunCreateCluster(f, out)
}
