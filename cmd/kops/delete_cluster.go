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
	"fmt"
	"github.com/spf13/cobra"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kops/util/pkg/vfs"
	"os"
)

type DeleteClusterCmd struct {
	Yes        bool
	Region     string
	External   bool
	Unregister bool
}

var deleteCluster DeleteClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster CLUSTERNAME [--yes]",
		Short: "Delete cluster",
		Long:  `Deletes a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteCluster.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	deleteCmd.AddCommand(cmd)

	cmd.Flags().BoolVar(&deleteCluster.Yes, "yes", false, "Delete without confirmation")
	cmd.Flags().BoolVar(&deleteCluster.Unregister, "unregister", false, "Don't delete cloud resources, just unregister the cluster")
	cmd.Flags().BoolVar(&deleteCluster.External, "external", false, "Delete an external cluster")

	cmd.Flags().StringVar(&deleteCluster.Region, "region", "", "region")
}

type getter func(o interface{}) interface{}

func (c *DeleteClusterCmd) Run(args []string) error {
	var configBase vfs.Path
	var clientset simple.Clientset

	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	clusterName := rootCommand.clusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required (for safety)")
	}

	var cloud fi.Cloud
	var cluster *api.Cluster

	if c.External {
		region := c.Region
		if region == "" {
			return fmt.Errorf("--region is required (when --external)")
		}

		tags := map[string]string{"KubernetesCluster": clusterName}
		cloud, err = awsup.NewAWSCloud(region, tags)
		if err != nil {
			return fmt.Errorf("error initializing AWS client: %v", err)
		}
	} else {
		clientset, err = rootCommand.Clientset()
		if err != nil {
			return err
		}

		cluster, err = clientset.Clusters().Get(clusterName)
		if err != nil {
			return err
		}

		if cluster == nil {
			return fmt.Errorf("cluster %q not found", clusterName)
		}

		if clusterName != cluster.ObjectMeta.Name {
			return fmt.Errorf("sanity check failed: cluster name mismatch")
		}

		configBase, err = registry.ConfigBase(cluster)
		if err != nil {
			return err
		}
	}

	wouldDeleteCloudResources := false

	if !c.Unregister {
		if cloud == nil {
			cloud, err = cloudup.BuildCloud(cluster)
			if err != nil {
				return err
			}
		}

		d := &kutil.DeleteCluster{}
		d.ClusterName = clusterName
		d.Cloud = cloud

		resources, err := d.ListResources()
		if err != nil {
			return err
		}

		if len(resources) == 0 {
			fmt.Printf("No cloud resources to delete\n")
		} else {
			wouldDeleteCloudResources = true

			t := &tables.Table{}
			t.AddColumn("TYPE", func(r *kutil.ResourceTracker) string {
				return r.Type
			})
			t.AddColumn("ID", func(r *kutil.ResourceTracker) string {
				return r.ID
			})
			t.AddColumn("NAME", func(r *kutil.ResourceTracker) string {
				return r.Name
			})
			var l []*kutil.ResourceTracker
			for _, v := range resources {
				l = append(l, v)
			}

			err := t.Render(l, os.Stdout, "TYPE", "NAME", "ID")
			if err != nil {
				return err
			}

			if !c.Yes {
				return fmt.Errorf("Must specify --yes to delete")
			}

			fmt.Fprintf(os.Stdout, "\n")

			err = d.DeleteResources(resources)
			if err != nil {
				return err
			}
		}
	}

	if !c.External {
		if !c.Yes {
			if wouldDeleteCloudResources {
				fmt.Printf("\nMust specify --yes to delete cloud resources & unregister cluster\n")
			} else {
				fmt.Printf("\nMust specify --yes to unregister the cluster\n")
			}
			return nil
		}
		err := registry.DeleteAllClusterState(configBase)
		if err != nil {
			return fmt.Errorf("error removing cluster from state store: %v", err)
		}
	}

	fmt.Printf("\nCluster deleted\n")
	return nil
}
