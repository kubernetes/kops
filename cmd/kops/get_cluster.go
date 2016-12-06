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
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/util/pkg/tables"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"os"
	"strings"
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

	output := getCmd.output
	if output == OutputTable {
		t := &tables.Table{}
		t.AddColumn("NAME", func(c *api.Cluster) string {
			return c.ObjectMeta.Name
		})
		t.AddColumn("CLOUD", func(c *api.Cluster) string {
			return c.Spec.CloudProvider
		})
		t.AddColumn("ZONES", func(c *api.Cluster) string {
			var zoneNames []string
			for _, z := range c.Spec.Zones {
				zoneNames = append(zoneNames, z.Name)
			}
			return strings.Join(zoneNames, ",")
		})
		return t.Render(clusters, os.Stdout, "NAME", "CLOUD", "ZONES")
	} else if output == OutputYaml {
		if c.FullSpec {
			var fullSpecs []*api.Cluster
			for _, cluster := range clusters {
				configBase, err := registry.ConfigBase(cluster)
				if err != nil {
					return fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
				}
				fullSpec := &api.Cluster{}
				err = registry.ReadConfigDeprecated(configBase.Join(registry.PathClusterCompleted), fullSpec)
				if err != nil {
					return fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
				}
				fullSpecs = append(fullSpecs, fullSpec)
			}
			clusters = fullSpecs
		}

		for _, cluster := range clusters {
			y, err := api.ToYaml(cluster)
			if err != nil {
				return fmt.Errorf("error marshaling yaml for %q: %v", cluster.ObjectMeta.Name, err)
			}
			_, err = os.Stdout.Write(y)
			if err != nil {
				return fmt.Errorf("error writing to stdout: %v", err)
			}
		}
		return nil
	} else {
		return fmt.Errorf("Unknown output format: %q", output)
	}
}
