/*
Copyright 2019 The Kubernetes Authors.

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

	"io"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/instancegroups"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/ui"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	deleteIgLong = templates.LongDesc(i18n.T(`
		Delete an instancegroup configuration.  kops has the concept of "instance groups",
		which are a group of similar virtual machines. On AWS, they map to an
		AutoScalingGroup. An ig work either as a Kubernetes master or a node.`))

	deleteIgExample = templates.Examples(i18n.T(`

		# Delete an instancegroup for the k8s-cluster.example.com cluster.
		# The --yes option runs the command immediately.
		# Note that the cloud resources will be deleted immediately, without running "kops update cluster"
		kops delete ig --name=k8s-cluster.example.com node-example --yes
		`))

	deleteIgShort = i18n.T(`Delete instancegroup`)
)

type DeleteInstanceGroupOptions struct {
	Yes         bool
	ClusterName string
	GroupName   string
}

func NewCmdDeleteInstanceGroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteInstanceGroupOptions{}

	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   deleteIgShort,
		Long:    deleteIgLong,
		Example: deleteIgExample,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				exitWithError(fmt.Errorf("Specify name of instance group to delete"))
			}
			if len(args) != 1 {
				exitWithError(fmt.Errorf("Can only edit one instance group at a time!"))
			}

			groupName := args[0]
			options.GroupName = groupName

			options.ClusterName = rootCommand.ClusterName()

			if !options.Yes {
				message := fmt.Sprintf("Do you really want to delete instance group %q? This action cannot be undone.", groupName)

				c := &ui.ConfirmArgs{
					Out:     out,
					Message: message,
					Default: "no",
					Retries: 2,
				}

				confirmed, err := ui.GetConfirm(c)
				if err != nil {
					exitWithError(err)
				}
				if !confirmed {
					os.Exit(1)
				} else {
					options.Yes = true
				}
			}

			err := RunDeleteInstanceGroup(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Specify --yes to immediately delete the instance group")

	return cmd
}

// RunDeleteInstanceGroup runs the deletion of an instance group
func RunDeleteInstanceGroup(f *util.Factory, out io.Writer, options *DeleteInstanceGroupOptions) error {

	// TODO make this drain and validate the ig?
	// TODO implement drain and validate logic
	groupName := options.GroupName
	if groupName == "" {
		return fmt.Errorf("GroupName is required")
	}

	clusterName := options.ClusterName
	if clusterName == "" {
		return fmt.Errorf("ClusterName is required")
	}

	cluster, err := GetCluster(f, clusterName)
	if err != nil {
		return err
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	group, err := clientset.InstanceGroupsFor(cluster).Get(groupName, metav1.GetOptions{})
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

	fmt.Fprintf(out, "InstanceGroup %q found for deletion\n", groupName)

	if !options.Yes {
		fmt.Fprintf(out, "\nMust specify --yes to delete instancegroup\n")
		return nil
	}

	d := &instancegroups.DeleteInstanceGroup{}
	d.Cluster = cluster
	d.Cloud = cloud
	d.Clientset = clientset

	err = d.DeleteInstanceGroup(group)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "\nDeleted InstanceGroup: %q\n", group.ObjectMeta.Name)

	return nil
}
