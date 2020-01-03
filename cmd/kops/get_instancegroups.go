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
	"strconv"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/formatter"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	getInstancegroupsLong = templates.LongDesc(i18n.T(`
	Display one or many instancegroup resources.`))

	getInstancegroupsExample = templates.Examples(i18n.T(`
	# Get all instancegroups in a state store
	kops get ig

	# Get a cluster's instancegroup
	kops get ig --name k8s-cluster.example.com nodes

	# Save a cluster's instancegroups desired configuration to YAML file
	kops get ig --name k8s-cluster.example.com -o yaml > instancegroups-desired-config.yaml
	`))

	getInstancegroupsShort = i18n.T(`Get one or many instancegroups`)
)

type GetInstanceGroupsOptions struct {
	*GetOptions
}

func NewCmdGetInstanceGroups(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := GetInstanceGroupsOptions{
		GetOptions: getOptions,
	}

	cmd := &cobra.Command{
		Use:     "instancegroups",
		Aliases: []string{"instancegroup", "ig"},
		Short:   getInstancegroupsShort,
		Long:    getInstancegroupsLong,
		Example: getInstancegroupsExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunGetInstanceGroups(&options, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunGetInstanceGroups(options *GetInstanceGroupsOptions, args []string) error {
	out := os.Stdout

	clusterName := rootCommand.ClusterName()
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(clusterName)
	if err != nil {
		return fmt.Errorf("error fetching cluster %q: %v", clusterName, err)
	}

	if cluster == nil {
		return fmt.Errorf("cluster %q was not found", clusterName)
	}

	list, err := clientset.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	instancegroups, err := filterInstanceGroupsByName(args, list.Items)
	if err != nil {
		return err
	}

	if len(instancegroups) == 0 {
		return fmt.Errorf("No InstanceGroup objects found")
	}

	var obj []runtime.Object
	if options.output != OutputTable {
		for _, c := range instancegroups {
			obj = append(obj, c)
		}
	}

	switch options.output {
	case OutputTable:
		return igOutputTable(cluster, instancegroups, out)
	case OutputYaml:
		return fullOutputYAML(out, obj...)
	case OutputJSON:
		return fullOutputJSON(out, obj...)
	default:
		return fmt.Errorf("Unknown output format: %q", options.output)
	}
}

func filterInstanceGroupsByName(instanceGroupNames []string, list []kopsapi.InstanceGroup) ([]*kopsapi.InstanceGroup, error) {
	var instancegroups []*kopsapi.InstanceGroup
	if len(instanceGroupNames) != 0 {
		// Build a map so we can return items in the same order
		m := make(map[string]*kopsapi.InstanceGroup)
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

func igOutputTable(cluster *kopsapi.Cluster, instancegroups []*kopsapi.InstanceGroup, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("NAME", func(c *kopsapi.InstanceGroup) string {
		return c.ObjectMeta.Name
	})
	t.AddColumn("ROLE", func(c *kopsapi.InstanceGroup) string {
		return string(c.Spec.Role)
	})
	t.AddColumn("MACHINETYPE", func(c *kopsapi.InstanceGroup) string {
		return c.Spec.MachineType
	})
	t.AddColumn("SUBNETS", formatter.RenderInstanceGroupSubnets(cluster))
	t.AddColumn("ZONES", formatter.RenderInstanceGroupZones(cluster))
	t.AddColumn("MIN", func(c *kopsapi.InstanceGroup) string {
		return int32PointerToString(c.Spec.MinSize)
	})
	t.AddColumn("MAX", func(c *kopsapi.InstanceGroup) string {
		return int32PointerToString(c.Spec.MaxSize)
	})
	// SUBNETS is not selected by default - not as useful as ZONES
	return t.Render(instancegroups, os.Stdout, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "ZONES")
}

func int32PointerToString(v *int32) string {
	if v == nil {
		return "-"
	}
	return strconv.Itoa(int(*v))
}
