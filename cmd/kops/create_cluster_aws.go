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

type CreateClusterAwsOptions struct {
	CreateClusterOptions // Global flags
	VPCID                string   // VPC ID
	DNSZone              string   // DNS hosted zone
	NodeSecurityGroups   []string // List of security group IDs for nodes
	MasterSecurityGroups []string // List of security group IDs for masters
	DNSType              string   // The DNS type to use (public/private)
	MasterTenancy        string   // Specify tenancy (default or dedicated)
	NodeTenancy          string   // Specify tenancy (default or dedicated)
}

func NewCmdCreateClusterAws(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateClusterAwsOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "aws",
		Short:   i18n.T("Create a Kubernetes cluster in AWS"),
		Long:    create_cluster_long,
		Example: create_cluster_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
				return
			}
			options.ClusterName = rootCommand.clusterName
			err = RunCreateClusterAws(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	// Cloud specific
	options.Cloud = "aws"
	cmd.Flags().StringVar(&options.VPCID, "vpc", options.VPCID, "Set to use a shared VPC")
	cmd.Flags().StringVar(&options.DNSZone, "dns-zone", options.DNSZone, "DNS hosted zone to use (defaults to longest matching zone)")
	cmd.Flags().StringSliceVar(&options.NodeSecurityGroups, "node-security-groups", options.NodeSecurityGroups, "Add precreated additional security groups to nodes.")
	cmd.Flags().StringSliceVar(&options.MasterSecurityGroups, "master-security-groups", options.MasterSecurityGroups, "Add precreated additional security groups to masters.")
	cmd.Flags().StringVar(&options.DNSType, "dns", options.DNSType, "DNS hosted zone to use: public|private. Default is 'public'.")
	cmd.Flags().StringVar(&options.MasterTenancy, "master-tenancy", options.MasterTenancy, "The tenancy of the master group on AWS. Can either be default or dedicated.")
	cmd.Flags().StringVar(&options.NodeTenancy, "node-tenancy", options.NodeTenancy, "The tenancy of the node group on AWS. Can be either default or dedicated.")


	// Global Flags
	options.createClusterGlobalFlags(cmd)

	return cmd
}

func RunCreateClusterAws(f *util.Factory, out io.Writer, c *CreateClusterAwsOptions) error {

	// AWS Logic to go here

	return c.RunCreateCluster(f, out)
}
