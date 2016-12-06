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
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kops/util/pkg/tables"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"os"
	"strconv"
	"time"
)

type RollingUpdateClusterCmd struct {
	Yes            bool
	Force          bool
	CloudOnly      bool
	MasterInterval time.Duration
	NodeInterval   time.Duration

	cobraCommand *cobra.Command
}

var rollingupdateCluster = RollingUpdateClusterCmd{
	cobraCommand: &cobra.Command{
		Use:   "cluster",
		Short: "rolling-update cluster",
		Long:  `rolling-updates a k8s cluster.`,
	},
}

func init() {
	cmd := rollingupdateCluster.cobraCommand
	rollingUpdateCommand.cobraCommand.AddCommand(cmd)

	cmd.Flags().BoolVar(&rollingupdateCluster.Yes, "yes", false, "perform rolling update without confirmation")
	cmd.Flags().BoolVar(&rollingupdateCluster.Force, "force", false, "Force rolling update, even if no changes")
	cmd.Flags().BoolVar(&rollingupdateCluster.CloudOnly, "cloudonly", false, "Perform rolling update without confirming progress with k8s")
	cmd.Flags().DurationVar(&rollingupdateCluster.MasterInterval, "master-interval", 5*time.Minute, "Time to wait between restarting masters")
	cmd.Flags().DurationVar(&rollingupdateCluster.NodeInterval, "node-interval", 2*time.Minute, "Time to wait between restarting nodes")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := rollingupdateCluster.Run(args)
		if err != nil {
			exitWithError(err)
		}
	}
}

func (c *RollingUpdateClusterCmd) Run(args []string) error {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	contextName := cluster.ObjectMeta.Name
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: contextName}).ClientConfig()
	if err != nil {
		return fmt.Errorf("cannot load kubecfg settings for %q: %v", contextName, err)
	}

	var nodes []v1.Node
	var k8sClient *release_1_5.Clientset
	if !c.CloudOnly {
		k8sClient, err = release_1_5.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("cannot build kube client for %q: %v", contextName, err)
		}

		nodeList, err := k8sClient.Core().Nodes().List(v1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to reach the kubernetes API.\n")
			fmt.Fprintf(os.Stderr, "Use --cloudonly to do a rolling-update without confirming progress with the k8s API\n\n")
			return fmt.Errorf("error listing nodes in cluster: %v", err)
		}

		if nodeList != nil {
			nodes = nodeList.Items
		}
	}

	list, err := clientset.InstanceGroups(cluster.ObjectMeta.Name).List(k8sapi.ListOptions{})
	if err != nil {
		return err
	}
	var instancegroups []*api.InstanceGroup
	for i := range list.Items {
		instancegroups = append(instancegroups, &list.Items[i])
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	d := &kutil.RollingUpdateCluster{
		MasterInterval: c.MasterInterval,
		NodeInterval:   c.NodeInterval,
		Force:          c.Force,
	}
	d.Cloud = cloud

	warnUnmatched := true
	groups, err := kutil.FindCloudInstanceGroups(cloud, cluster, instancegroups, warnUnmatched, nodes)
	if err != nil {
		return err
	}

	{
		t := &tables.Table{}
		t.AddColumn("NAME", func(r *kutil.CloudInstanceGroup) string {
			return r.InstanceGroup.ObjectMeta.Name
		})
		t.AddColumn("STATUS", func(r *kutil.CloudInstanceGroup) string {
			return r.Status
		})
		t.AddColumn("NEEDUPDATE", func(r *kutil.CloudInstanceGroup) string {
			return strconv.Itoa(len(r.NeedUpdate))
		})
		t.AddColumn("READY", func(r *kutil.CloudInstanceGroup) string {
			return strconv.Itoa(len(r.Ready))
		})
		t.AddColumn("MIN", func(r *kutil.CloudInstanceGroup) string {
			return strconv.Itoa(r.MinSize())
		})
		t.AddColumn("MAX", func(r *kutil.CloudInstanceGroup) string {
			return strconv.Itoa(r.MaxSize())
		})
		t.AddColumn("NODES", func(r *kutil.CloudInstanceGroup) string {
			var nodes []*v1.Node
			for _, i := range r.Ready {
				if i.Node != nil {
					nodes = append(nodes, i.Node)
				}
			}
			for _, i := range r.NeedUpdate {
				if i.Node != nil {
					nodes = append(nodes, i.Node)
				}
			}
			return strconv.Itoa(len(nodes))
		})
		var l []*kutil.CloudInstanceGroup
		for _, v := range groups {
			l = append(l, v)
		}

		columns := []string{"NAME", "STATUS", "NEEDUPDATE", "READY", "MIN", "MAX"}
		if !c.CloudOnly {
			columns = append(columns, "NODES")
		}
		err := t.Render(l, os.Stdout, columns...)
		if err != nil {
			return err
		}
	}

	needUpdate := false
	for _, group := range groups {
		if len(group.NeedUpdate) != 0 {
			needUpdate = true
		}
	}

	if !needUpdate && !c.Force {
		fmt.Printf("\nNo rolling-update required\n")
		return nil
	}

	if !c.Yes {
		fmt.Printf("\nMust specify --yes to rolling-update\n")
		return nil
	}

	return d.RollingUpdate(groups, k8sClient)
}
