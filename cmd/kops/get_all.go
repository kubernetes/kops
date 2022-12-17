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

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getAllLong = templates.LongDesc(i18n.T(`
	Display all resources for a cluster.`))

	getAllExample = templates.Examples(i18n.T(`
	# Get a cluster, its instance groups, and its addons
	kops get all k8s-cluster.example.com

	# Get a cluster, its instance groups, and its addons in YAML format
	kops get all k8s-cluster.example.com -o yaml
	`))

	getAllShort = i18n.T(`Display all resources for a cluster.`)
)

type GetAllOptions struct {
	*GetOptions
}

func NewCmdGetAll(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := &GetAllOptions{
		GetOptions: getOptions,
	}

	cmd := &cobra.Command{
		Use:               "all [CLUSTER]",
		Short:             getAllShort,
		Long:              getAllLong,
		Example:           getAllExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGetAll(cmd.Context(), f, out, options)
		},
	}

	return cmd
}

func RunGetAll(ctx context.Context, f commandutils.Factory, out io.Writer, options *GetAllOptions) error {
	client, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := client.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("no cluster found")
	}

	igList, err := client.InstanceGroupsFor(ctx, cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	if igList == nil || igList.Items == nil || len(igList.Items) == 0 {
		fmt.Fprintf(os.Stderr, "No instance groups found\n")
	}

	var instancegroups []*api.InstanceGroup
	for i := range igList.Items {
		instancegroups = append(instancegroups, &igList.Items[i])
	}

	var addonObjects []*unstructured.Unstructured
	{
		addons, err := client.AddonsFor(ctx, cluster).List()
		if err != nil {
			return err
		}
		for _, addon := range addons {
			addonObjects = append(addonObjects, addon.ToUnstructured())
		}
	}

	var allObjects []runtime.Object
	if options.Output != OutputTable {
		allObjects = append(allObjects, cluster)
		for _, group := range instancegroups {
			allObjects = append(allObjects, group)
		}
		for _, additionalObject := range addonObjects {
			allObjects = append(allObjects, additionalObject)
		}
	}

	switch options.Output {
	case OutputYaml:
		if err := fullOutputYAML(out, allObjects...); err != nil {
			return fmt.Errorf("error writing yaml to stdout: %v", err)
		}

		return nil

	case OutputJSON:
		if err := fullOutputJSON(out, false, allObjects...); err != nil {
			return fmt.Errorf("error writing json to stdout: %v", err)
		}
		return nil

	case OutputTable:
		fmt.Fprintf(out, "Cluster\n")
		err = clusterOutputTable([]*api.Cluster{cluster}, out)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "\nInstance Groups\n")
		err = igOutputTable(cluster, instancegroups, out)
		if err != nil {
			return err
		}
		if len(addonObjects) != 0 {
			fmt.Fprintf(out, "\nAddon Objects\n")
			err = addonsOutputTable(cluster, addonObjects, out)
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown output format: %q", options.Output)
	}

	return nil
}
