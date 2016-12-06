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
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/kutil"
)

type DeleteInstanceceGroupCmd struct {
}

var deleteInstanceceGroupCmd DeleteInstanceceGroupCmd

func init() {
	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   "Delete instancegroup",
		Long:    `Delete an instancegroup configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				exitWithError(fmt.Errorf("Specify name of instance group to delete"))
			}
			if len(args) != 1 {
				exitWithError(fmt.Errorf("Can only edit one instance group at a time!"))
			}
			err := deleteInstanceceGroupCmd.Run(args[0])
			if err != nil {
				exitWithError(err)
			}
		},
	}

	deleteCmd.AddCommand(cmd)
}

func (c *DeleteInstanceceGroupCmd) Run(groupName string) error {
	if groupName == "" {
		return fmt.Errorf("name is required")
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	group, err := clientset.InstanceGroups(cluster.ObjectMeta.Name).Get(groupName)
	if err != nil {
		return fmt.Errorf("error reading InstanceGroup %q: %v", groupName, err)
	}
	if group == nil {
		return fmt.Errorf("InstanceGroup %q not found", groupName)
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	d := &kutil.DeleteInstanceGroup{}
	d.Cluster = cluster
	d.Cloud = cloud
	d.Clientset = clientset

	err = d.DeleteInstanceGroup(group)
	if err != nil {
		return err
	}

	fmt.Printf("InstanceGroup %q deleted\n", group.ObjectMeta.Name)

	return nil
}
