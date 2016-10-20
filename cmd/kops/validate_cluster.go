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
	"strings"

	"github.com/spf13/cobra"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/tables"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"

	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
)

// not used too much yet :)
type ValidateClusterCmd struct {
	FullSpec bool
}

var validateClusterCmd ValidateClusterCmd

// Init darn it
func init() {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"cluster"},
		Short:   "Validate cluster",
		Long:    `Validate a kubernetes cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			err := validateClusterCmd.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	validateCmd.cobraCommand.AddCommand(cmd)

	//cmd.Flags().BoolVar(&validateClusterCmd.FullSpec, "full", false, "Validate a cluster")
}

// Run a validation
func (c *ValidateClusterCmd) Run(args []string) error {

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

	nodeAA := &api.NodeAPIAdapter{}

	timeout, err := time.ParseDuration("30s")

	if err != nil {
		return fmt.Errorf("Cannot set timeout %q: %v", cluster.Name, err)
	}

	nodeAA.BuildNodeAPIAdapter(cluster.Name, timeout, "")

	nodes, err := nodeAA.GetAllNodes()

	if err != nil {
		return fmt.Errorf("Cannot get nodes for %q: %v", cluster.Name, err)
	}

	list, err := clientset.InstanceGroups(cluster.Name).List(k8sapi.ListOptions{})
	if err != nil {
		return fmt.Errorf("Cannot get nodes for %q: %v", cluster.Name, err)
	}

	fmt.Printf("Validating cluster %v\n\n", cluster.Name)

	var instancegroups []*api.InstanceGroup
	validationCluster := &api.ValidationCluster{}
	for i := range list.Items {
		ig := &list.Items[i]
		instancegroups = append(instancegroups, ig)
		if ig.Spec.Role == api.InstanceGroupRoleMaster {
			//validationCluster.MastersInstanceGroups = append(validationCluster.MastersInstanceGroups, ig)
			validationCluster.MastersCount += *ig.Spec.MinSize
		} else {
			//validationCluster.NodesInstanceGroups = append(validationCluster.NodesInstanceGroups, ig)
			validationCluster.NodesCount += *ig.Spec.MinSize
		}
	}

	if len(instancegroups) == 0 {
		return errors.New("No InstanceGroup objects found\n")
	}

	t := &tables.Table{}
	t.AddColumn("NAME", func(c *api.InstanceGroup) string {
		return c.Name
	})
	t.AddColumn("ROLE", func(c *api.InstanceGroup) string {
		return string(c.Spec.Role)
	})
	t.AddColumn("MACHINETYPE", func(c *api.InstanceGroup) string {
		return c.Spec.MachineType
	})
	t.AddColumn("ZONES", func(c *api.InstanceGroup) string {
		return strings.Join(c.Spec.Zones, ",")
	})
	t.AddColumn("MIN", func(c *api.InstanceGroup) string {
		return intPointerToString(c.Spec.MinSize)
	})
	t.AddColumn("MAX", func(c *api.InstanceGroup) string {
		return intPointerToString(c.Spec.MaxSize)
	})

	fmt.Println("INSTANCE GROUPS")
	err = t.Render(instancegroups, os.Stdout, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "ZONES")

	if err != nil {
		return fmt.Errorf("Cannot render nodes for %q: %v", cluster.Name, err)
	}

	t = &tables.Table{}

	t.AddColumn("NAME", func(n v1.Node) string {
		return n.Name
	})

	t.AddColumn("READY", func(n v1.Node) v1.ConditionStatus {
		return api.GetNodeConditionStatus(n.Status.Conditions)
	})

	t.AddColumn("ROLE", func(n v1.Node) string {
		role := "node"
		if val, ok := n.ObjectMeta.Labels["kubernetes.io/role"]; ok {
			role = val
		}
		return role
	})

	fmt.Println("\nNODE STATUS")
	err = t.Render(nodes.Items, os.Stdout, "NAME", "ROLE", "READY")

	if err != nil {
		return fmt.Errorf("Cannot render nodes for %q: %v", cluster.Name, err)
	}

	for _, node := range nodes.Items {

		role := "node"
		if val, ok := node.ObjectMeta.Labels["kubernetes.io/role"]; ok {
			role = val
		}

		n := &api.ValidationNode{
			Zone:     node.ObjectMeta.Labels["failure-domain.beta.kubernetes.io/zone"],
			Hostname: node.ObjectMeta.Labels["kubernetes.io/hostname"],
			Role:     role,
			Status:   api.GetNodeConditionStatus(node.Status.Conditions),
		}

		if n.Role == "master" {
			if n.Status == v1.ConditionTrue {
				validationCluster.MastersReady = append(validationCluster.MastersReady, n)
			} else {
				validationCluster.MastersNotReady = append(validationCluster.MastersNotReady, n)
			}
		} else if n.Role == "node" {
			if n.Status == v1.ConditionTrue {
				validationCluster.NodesReady = append(validationCluster.NodesReady, n)
			} else {
				validationCluster.NodesNotReady = append(validationCluster.NodesNotReady, n)
			}

		}

	}

	mastersReady := true
	nodesReady := true
	if len(validationCluster.MastersNotReady) != 0 || validationCluster.MastersCount !=
		len(validationCluster.MastersReady) {
		mastersReady = false
	}

	if len(validationCluster.NodesNotReady) != 0 || validationCluster.NodesCount !=
		len(validationCluster.NodesReady) {
		nodesReady = false
	}


	if mastersReady && nodesReady {
		fmt.Printf("\nYour cluster %s is ready\n", cluster.Name)
		return nil
	} else {
		// do we need to print which instance group is not ready?
		// nodes are going to be a pain
		glog.Infof("cluster - masters ready: %v, nodes ready: %v", mastersReady, nodesReady)
		glog.Infof("mastersNotReady %v", len(validationCluster.MastersNotReady))
		glog.Infof("mastersCount %v, mastersReady %v", validationCluster.MastersCount, len(validationCluster.MastersReady))
		glog.Infof("nodesNotReady %v", len(validationCluster.NodesNotReady))
		glog.Infof("nodesCount %v, nodesReady %v", validationCluster.NodesCount, len(validationCluster.NodesReady))
		return fmt.Errorf("You cluster is NOT ready %s", cluster.Name)
	}

}
