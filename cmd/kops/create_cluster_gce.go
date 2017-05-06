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
	//"k8s.io/kops/pkg/client/simple/vfsclientset"
	//"k8s.io/kops/upup/pkg/fi"
)

type CreateClusterGceOptions struct {
	CreateClusterOptions // Global flags
	Project string       // Project for GCE clusters
}

func NewCmdCreateClusterGce(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateClusterGceOptions{}
	options.InitDefaults()
	cmd := &cobra.Command{
		Use:     "gce",
		Short:   i18n.T("Create a Kubernetes cluster in GCE"),
		Long:    create_cluster_long,
		Example: create_cluster_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
				return
			}
			options.ClusterName = rootCommand.clusterName
			err = RunCreateClusterGce(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	// Cloud specific overrides
	options.Cloud = "gce"
	cmd.Flags().StringVar(&options.Project, "project", options.Project, "Project to use (must be set on GCE)")

	// Global Flags
	options.createClusterGlobalFlags(cmd)
	return cmd
}

func RunCreateClusterGce(f *util.Factory, out io.Writer, c *CreateClusterGceOptions) error {

	// GCE Logic to go here

	return c.RunCreateCluster(f, out)
}
