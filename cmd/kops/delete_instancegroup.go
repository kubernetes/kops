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
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/instancegroups"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/ui"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	deleteInstanceGroupLong = templates.LongDesc(i18n.T(`
		Delete an instance group configuration. kOps has the concept of "instance groups",
		which are a group of similar virtual machines. On AWS, they map to an
		AutoScalingGroup.`))

	deleteInstanceGroupExample = templates.Examples(i18n.T(`

		# Delete an instancegroup for the k8s-cluster.example.com cluster.
		# The --yes option runs the command immediately.
		# Note that the cloud resources will be deleted immediately, without running "kops update cluster"
		kops delete ig --name=k8s-cluster.example.com node-example --yes
		`))

	deleteInstanceGroupShort = i18n.T(`Delete instance group.`)
)

type DeleteInstanceGroupOptions struct {
	Yes         bool
	ClusterName string
	GroupName   string
}

func NewCmdDeleteInstanceGroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteInstanceGroupOptions{}

	cmd := &cobra.Command{
		Use:     "instancegroup INSTANCE_GROUP",
		Aliases: []string{"instancegroups", "ig"},
		Short:   deleteInstanceGroupShort,
		Long:    deleteInstanceGroupLong,
		Example: deleteInstanceGroupExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)

			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify the name of the instance group to delete")
			}

			options.GroupName = args[0]

			if len(args) != 1 {
				return fmt.Errorf("can only delete one instance group at a time")
			}

			return nil
		},
		ValidArgsFunction: completeInstanceGroup(f, nil, &[]string{strings.ToLower(string(kops.InstanceGroupRoleMaster))}),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()

			if !options.Yes {
				message := fmt.Sprintf("Do you really want to delete instance group %q? This action cannot be undone.", options.GroupName)

				c := &ui.ConfirmArgs{
					Out:     out,
					Message: message,
					Default: "no",
					Retries: 2,
				}

				confirmed, err := ui.GetConfirm(c)
				if err != nil {
					return err
				}
				if !confirmed {
					os.Exit(1)
				} else {
					options.Yes = true
				}
			}

			return RunDeleteInstanceGroup(ctx, f, out, options)
		},
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Specify --yes to immediately delete the instance group")

	return cmd
}

// RunDeleteInstanceGroup runs the deletion of an instance group
func RunDeleteInstanceGroup(ctx context.Context, f *util.Factory, out io.Writer, options *DeleteInstanceGroupOptions) error {
	// TODO make this drain and validate the ig?
	// TODO implement drain and validate logic
	groupName := options.GroupName
	if groupName == "" {
		return fmt.Errorf("GroupName is required")
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	group, err := clientset.InstanceGroupsFor(cluster).Get(ctx, groupName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error reading InstanceGroup %q: %v", groupName, err)
	}
	if group == nil {
		return fmt.Errorf("InstanceGroup %q not found", groupName)
	}

	fmt.Fprintf(out, "InstanceGroup %q found for deletion\n", groupName)

	if group.Spec.Role == kops.InstanceGroupRoleMaster {
		groups, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("listing InstanceGroups: %v", err)
		}

		onlyMaster := true
		for _, ig := range groups.Items {
			if ig.Name != groupName && ig.Spec.Role == kops.InstanceGroupRoleMaster {
				onlyMaster = false
				break
			}
		}

		if onlyMaster {
			return fmt.Errorf("cannot delete the only control plane instance group")
		}
	}

	if !options.Yes {
		fmt.Fprintf(out, "\nMust specify --yes to delete instancegroup\n")
		return nil
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
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
