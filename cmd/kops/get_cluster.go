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
	"strings"

	"github.com/spf13/cobra"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/util/pkg/tables"
	k8sapi "k8s.io/kubernetes/pkg/api"
)

type GetClustersCmd struct {
	FullSpec bool
}

var getClustersCmd GetClustersCmd

func init() {
	cmd := &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   "get clusters",
		Long:    `List or get clusters.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getClustersCmd.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	getCmd.cobraCommand.AddCommand(cmd)

	cmd.Flags().BoolVar(&getClustersCmd.FullSpec, "full", false, "Show fully populated configuration")
}

func (c *GetClustersCmd) Run(args []string) error {
	client, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	clusterList, err := client.Clusters().List(k8sapi.ListOptions{})
	if err != nil {
		return err
	}

	var clusters []*api.Cluster
	if len(args) != 0 {
		m := make(map[string]*api.Cluster)
		for i := range clusterList.Items {
			c := &clusterList.Items[i]
			m[c.ObjectMeta.Name] = c
		}
		for _, arg := range args {
			ig := m[arg]
			if ig == nil {
				return fmt.Errorf("cluster not found %q", arg)
			}

			clusters = append(clusters, ig)
		}
	} else {
		for i := range clusterList.Items {
			c := &clusterList.Items[i]
			clusters = append(clusters, c)
		}
	}

	if len(clusters) == 0 {
		fmt.Fprintf(os.Stderr, "No clusters found\n")
		return nil
	}

	if c.FullSpec {
		var err error
		clusters, err = fullClusterSpecs(clusters)
		if err != nil {
			return err
		}
	}

	switch getCmd.output {
	case OutputTable:

		t := &tables.Table{}
		t.AddColumn("NAME", func(c *api.Cluster) string {
			return c.ObjectMeta.Name
		})
		t.AddColumn("CLOUD", func(c *api.Cluster) string {
			return c.Spec.CloudProvider
		})
		t.AddColumn("SUBNETS", func(c *api.Cluster) string {
			var subnetNames []string
			for _, s := range c.Spec.Subnets {
				subnetNames = append(subnetNames, s.Name)
			}
			return strings.Join(subnetNames, ",")
		})
		return t.Render(clusters, os.Stdout, "NAME", "CLOUD", "SUBNETS")

	case OutputYaml:
		for _, cluster := range clusters {
			if err := marshalToWriter(cluster, marshalYaml, os.Stdout); err != nil {
				return err
			}
		}
		return nil
	case OutputJSON:
		for _, cluster := range clusters {
			if err := marshalToWriter(cluster, marshalJSON, os.Stdout); err != nil {
				return err
			}
		}
		return nil

	default:
		return fmt.Errorf("Unknown output format: %q", getCmd.output)
	}
}

func fullClusterSpecs(clusters []*api.Cluster) ([]*api.Cluster, error) {
	var fullSpecs []*api.Cluster
	for _, cluster := range clusters {
		configBase, err := registry.ConfigBase(cluster)
		if err != nil {
			return nil, fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
		}
		fullSpec := &api.Cluster{}
		err = registry.ReadConfigDeprecated(configBase.Join(registry.PathClusterCompleted), fullSpec)
		if err != nil {
			return nil, fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
		}
		fullSpecs = append(fullSpecs, fullSpec)
	}
	return fullSpecs, nil
}
