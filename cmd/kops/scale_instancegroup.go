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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/instancegroups"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

type ScaleIgCmd struct {
	Yes      bool
	Replicas int64
}

const (
	defReplicas = -1
)

// TODO add ability to add-nodes rather than scale
var (
	//TODO add comments
	scale_instancegroup_long = templates.LongDesc(i18n.T(`
	long description...
	`))

	scale_instancegroup_example = templates.Examples(i18n.T(`
		# Scale a ig fixing it to 2 replicas
		kops scale ig --name cluster.kops.ddy.systems nodes --replicas=2

	`))
)

var scaleIg ScaleIgCmd

func init() {

	cmd := &cobra.Command{
		Use:     "ig",
		Aliases: []string{"instancegroup", "instancegroups"},
		Short:   i18n.T("Scale instances instancegroups"),
		Long:    scale_instancegroup_long,
		Example: scale_instancegroup_example,
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				exitWithError(fmt.Errorf("Specify name of instance group to edit"))
			}

			if len(args) != 1 {
				exitWithError(fmt.Errorf("Can only specify one instance group at a time"))
			}

			err := scaleIg.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().Int64Var(&scaleIg.Replicas, "replicas", defReplicas, i18n.T("The new desired number of replicas. Required."))

	scaleCmd.AddCommand(cmd)
}
func (c *ScaleIgCmd) Run(args []string) error {

	groupName := args[0]

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	if groupName == "" {
		return fmt.Errorf("name is required")
	}

	if c.Replicas == defReplicas {
		return fmt.Errorf("argument --replicas is required")
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	igGroup, err := clientset.InstanceGroupsFor(cluster).Get(groupName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error reading InstanceGroup %q: %v", groupName, err)
	}
	if igGroup == nil {
		return fmt.Errorf("InstanceGroup %q not found", groupName)
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	s := &instancegroups.ScaleInstanceGroup{Cluster: cluster, Cloud: cloud,
		Clientset: clientset, DesiredReplicas: c.Replicas}

	if err = s.ScaleInstanceGroup(igGroup); err != nil {
		return err
	}

	glog.Infof("Successful scaled! It will take few minutes to complete this operation...")

	return nil
}
