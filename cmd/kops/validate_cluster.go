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
)

// not used too much yet :)
type ValidateClusterCmd struct {
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
}

// Validate Your Kubernetes Cluster
func (c *ValidateClusterCmd) Run(args []string) error {

	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return fmt.Errorf("Process args failed %v", err)
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return fmt.Errorf("Cannot get cluster for %v", err)
	}

	clientSet, err := rootCommand.Clientset()
	if err != nil {
		return fmt.Errorf("Cannot get clientSet for %q: %v", cluster.Name, err)
	}

	list, err := clientSet.InstanceGroups(cluster.Name).List(k8sapi.ListOptions{})
	if err != nil {
		return fmt.Errorf("Cannot get nodes for %q: %v", cluster.Name, err)
	}

	fmt.Printf("Validating cluster %v\n\n", cluster.Name)

	var instanceGroups []*api.InstanceGroup
	for _, ig := range list.Items {
		instanceGroups = append(instanceGroups, &ig)
	}

	if len(instanceGroups) == 0 {
		return errors.New("No InstanceGroup objects found\n")
	}

	validationCluster, validationFailed := api.ValidateCluster(cluster.Name, list)

	if validationCluster.NodeList == nil {
		return fmt.Errorf("Cannot get nodes for %q: %v", cluster.Name, validationFailed)
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
	err = t.Render(instanceGroups, os.Stdout, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "ZONES")

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
	err = t.Render(validationCluster.NodeList.Items, os.Stdout, "NAME", "ROLE", "READY")

	if err != nil {
		return fmt.Errorf("Cannot render nodes for %q: %v", cluster.Name, err)
	}

	if validationFailed == nil {
		fmt.Printf("\nYour cluster %s is ready\n", cluster.Name)
		return nil
	} else {
		// do we need to print which instance group is not ready?
		// nodes are going to be a pain
		fmt.Printf("cluster - masters ready: %v, nodes ready: %v", validationCluster.MastersReady, validationCluster.NodesReady)
		fmt.Printf("mastersNotReady %v", len(validationCluster.MastersNotReadyArray))
		fmt.Printf("mastersCount %v, mastersReady %v", validationCluster.MastersCount, len(validationCluster.MastersReadyArray))
		fmt.Printf("nodesNotReady %v", len(validationCluster.NodesNotReadyArray))
		fmt.Printf("nodesCount %v, nodesReady %v", validationCluster.NodesCount, len(validationCluster.NodesReadyArray))
		return fmt.Errorf("\nYour cluster %s is NOT ready.", cluster.Name)
	}

}
