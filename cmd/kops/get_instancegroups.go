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
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	get_instancegroups_long = templates.LongDesc(i18n.T(`
	Display one or many instancegroup resources.`))

	get_instancegroups_example = templates.Examples(i18n.T(`
	# Get all instancegroups in a state store
	kops get ig

	# Get a cluster
	kops get ig --name k8s-cluster.example.com nodes`))

	get_instancegroups_short = i18n.T(`Get one or many instancegroups`)
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
		Short:   get_instancegroups_short,
		Long:    get_instancegroups_long,
		Example: get_instancegroups_example,
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

	instancegroups, err := buildInstanceGroups(args, list)
	if err != nil {
		return err
	}

	if len(instancegroups) == 0 {
		fmt.Fprintf(os.Stderr, "No InstanceGroup objects found\n")
		return nil
	}

	switch options.output {

	case OutputTable:
		return igOutputTable(instancegroups, out)
	case OutputYaml:
		return igOutputYAML(instancegroups, out)
	case OutputJSON:
		return igOutputJson(instancegroups, out)
	default:
		return fmt.Errorf("Unknown output format: %q", options.output)
	}
}

func buildInstanceGroups(args []string, list *api.InstanceGroupList) ([]*api.InstanceGroup, error) {
	var instancegroups []*api.InstanceGroup
	len := len(args)
	if len != 0 {
		m := make(map[string]*api.InstanceGroup)
		for i := range list.Items {
			ig := &list.Items[i]
			m[ig.ObjectMeta.Name] = ig
		}
		instancegroups = make([]*api.InstanceGroup, 0, len)
		for _, arg := range args {
			ig := m[arg]
			if ig == nil {
				return nil, fmt.Errorf("instancegroup not found %q", arg)
			}

			instancegroups = append(instancegroups, ig)
		}
	} else {
		for i := range list.Items {
			ig := &list.Items[i]
			instancegroups = append(instancegroups, ig)
		}
	}

	return instancegroups, nil
}

func igOutputTable(instancegroups []*api.InstanceGroup, out io.Writer) error {
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
	t.AddColumn("SUBNETS", func(c *api.InstanceGroup) string {
		return strings.Join(c.Spec.Subnets, ",")
	})
	t.AddColumn("MIN", func(c *api.InstanceGroup) string {
		return int32PointerToString(c.Spec.MinSize)
	})
	t.AddColumn("MAX", func(c *api.InstanceGroup) string {
		return int32PointerToString(c.Spec.MaxSize)
	})
	return t.Render(instancegroups, os.Stdout, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "SUBNETS")
}

func igOutputJson(instanceGroups []*api.InstanceGroup, out io.Writer) error {
	for _, ig := range instanceGroups {
		if err := marshalToWriter(ig, marshalJSON, out); err != nil {
			return err
		}
	}
	return nil
}

func igOutputYAML(instanceGroups []*api.InstanceGroup, out io.Writer) error {
	for i, ig := range instanceGroups {
		if i != 0 {
			if err := writeYAMLSep(out); err != nil {
				return fmt.Errorf("error writing to stdout: %v", err)
			}
		}
		if err := marshalToWriter(ig, marshalYaml, out); err != nil {
			return err
		}
	}
	return nil
}

func int32PointerToString(v *int32) string {
	if v == nil {
		return "-"
	}
	return strconv.Itoa(int(*v))
}
