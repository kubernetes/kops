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
	"strconv"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/formatter"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getInstancegroupsLong = templates.LongDesc(i18n.T(`
	Display one or many instance group resources.`))

	getInstancegroupsExample = templates.Examples(i18n.T(`
	# Get all instance groups in a state store
	kops get instancegroups

	# Get a cluster's instancegroup
	kops get instancegroups --name k8s-cluster.example.com nodes

	# Save a cluster's instancegroups desired configuration to YAML file
	kops get instancegroups --name k8s-cluster.example.com -o yaml > instancegroups-desired-config.yaml
	`))

	getInstancegroupsShort = i18n.T(`Get one or many instance groups.`)
)

type GetInstanceGroupsOptions struct {
	*GetOptions
	InstanceGroupNames []string
}

func NewCmdGetInstanceGroups(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := GetInstanceGroupsOptions{
		GetOptions: getOptions,
	}

	cmd := &cobra.Command{
		Use:     "instancegroups [INSTANCE_GROUP]...",
		Aliases: []string{"instancegroup", "ig"},
		Short:   getInstancegroupsShort,
		Long:    getInstancegroupsLong,
		Example: getInstancegroupsExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)
			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			options.InstanceGroupNames = args
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeInstanceGroup(f, &args, nil)(cmd, nil, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGetInstanceGroups(context.TODO(), f, out, &options)
		},
	}

	return cmd
}

func RunGetInstanceGroups(ctx context.Context, f commandutils.Factory, out io.Writer, options *GetInstanceGroupsOptions) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return fmt.Errorf("error fetching cluster %q: %v", options.ClusterName, err)
	}

	if cluster == nil {
		return fmt.Errorf("cluster %q was not found", options.ClusterName)
	}

	list, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	instancegroups, err := filterInstanceGroupsByName(options.InstanceGroupNames, list.Items)
	if err != nil {
		return err
	}

	singleObject := false

	if len(instancegroups) == 0 {
		return fmt.Errorf("no InstanceGroup objects found")
	} else if len(instancegroups) == 1 {
		singleObject = true
	}

	var obj []runtime.Object
	if options.Output != OutputTable {
		for _, c := range instancegroups {
			obj = append(obj, c)
		}
	}

	switch options.Output {
	case OutputTable:
		return igOutputTable(cluster, instancegroups, out)
	case OutputYaml:
		return fullOutputYAML(out, obj...)
	case OutputJSON:
		return fullOutputJSON(out, singleObject, obj...)
	default:
		return fmt.Errorf("unknown output format: %q", options.Output)
	}
}

func filterInstanceGroupsByName(instanceGroupNames []string, list []api.InstanceGroup) ([]*api.InstanceGroup, error) {
	var instancegroups []*api.InstanceGroup
	if len(instanceGroupNames) != 0 {
		// Build a map so we can return items in the same order
		m := make(map[string]*api.InstanceGroup)
		for i := range list {
			ig := &list[i]
			m[ig.ObjectMeta.Name] = ig
		}
		for _, name := range instanceGroupNames {
			ig := m[name]
			if ig == nil {
				return nil, fmt.Errorf("instancegroup not found %q", name)
			}

			instancegroups = append(instancegroups, ig)
		}
	} else {
		for i := range list {
			ig := &list[i]
			instancegroups = append(instancegroups, ig)
		}
	}

	return instancegroups, nil
}

func igOutputTable(cluster *api.Cluster, instancegroups []*api.InstanceGroup, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("NAME", func(c *api.InstanceGroup) string {
		return c.ObjectMeta.Name
	})
	t.AddColumn("ROLE", func(c *api.InstanceGroup) string {
		return string(c.Spec.Role)
	})
	t.AddColumn("MACHINETYPE", func(c *api.InstanceGroup) string {
		return c.Spec.MachineType
	})
	t.AddColumn("SUBNETS", formatter.RenderInstanceGroupSubnets(cluster))
	t.AddColumn("ZONES", formatter.RenderInstanceGroupZones(cluster))
	t.AddColumn("MIN", func(c *api.InstanceGroup) string {
		return int32PointerToString(c.Spec.MinSize)
	})
	t.AddColumn("MAX", func(c *api.InstanceGroup) string {
		return int32PointerToString(c.Spec.MaxSize)
	})
	// SUBNETS is not selected by default - not as useful as ZONES
	return t.Render(instancegroups, out, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "ZONES")
}

func int32PointerToString(v *int32) string {
	if v == nil {
		return "-"
	}
	return strconv.Itoa(int(*v))
}
