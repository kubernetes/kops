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

type CreateClusterVsphereOptions struct {
	CreateClusterOptions // Global flags
	VSphereServer        string
	VSphereDatacenter    string
	VSphereResourcePool  string
	VSphereCoreDNSServer string
	// Note: We need open-vm-tools to be installed for vSphere Cloud Provider to work
	// We need VSphereDatastore to support Kubernetes vSphere Cloud Provider (v1.5.3)
	// We can remove this once we support higher versions.
	VSphereDatastore string
}

func NewCmdCreateClusterVsphere(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateClusterVsphereOptions{}
	options.InitDefaults()
	cmd := &cobra.Command{
		Use:     "vsphere",
		Short:   i18n.T("Create a Kubernetes cluster in vSphere"),
		Long:    create_cluster_long,
		Example: create_cluster_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
				return
			}
			options.ClusterName = rootCommand.clusterName
			err = RunCreateClusterVsphere(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	// Cloud specific overrides
	options.Cloud = "vsphere"
	cmd.Flags().StringVar(&options.VSphereServer, "vsphere-server", options.VSphereServer, "vsphere-server is required for vSphere. Set vCenter URL Ex: 10.192.10.30 or myvcenter.io (without https://)")
	cmd.Flags().StringVar(&options.VSphereDatacenter, "vsphere-datacenter", options.VSphereDatacenter, "vsphere-datacenter is required for vSphere. Set the name of the datacenter in which to deploy Kubernetes VMs.")
	cmd.Flags().StringVar(&options.VSphereResourcePool, "vsphere-resource-pool", options.VSphereDatacenter, "vsphere-resource-pool is required for vSphere. Set a valid Cluster, Host or Resource Pool in which to deploy Kubernetes VMs.")
	cmd.Flags().StringVar(&options.VSphereCoreDNSServer, "vsphere-coredns-server", options.VSphereCoreDNSServer, "vsphere-coredns-server is required for vSphere.")
	cmd.Flags().StringVar(&options.VSphereDatastore, "vsphere-datastore", options.VSphereDatastore, "vsphere-datastore is required for vSphere.  Set a valid datastore in which to store dynamic provision volumes.")

	// Global Flags
	options.createClusterGlobalFlags(cmd)
	return cmd
}

func RunCreateClusterVsphere(f *util.Factory, out io.Writer, c *CreateClusterVsphereOptions) error {

	// vSphere Logic to go here

	return c.RunCreateCluster(f, out)
}
