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
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/util/pkg/tables"
)

type GetClusterOptions struct {
	// FullSpec determines if we should output the completed (fully populated) spec
	FullSpec bool

	// ClusterNames is a list of cluster names to show; if not specified all clusters will be shown
	ClusterNames []string
}

func init() {
	var options GetClusterOptions

	cmd := &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   "get clusters",
		Long:    `List or get clusters.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				options.ClusterNames = append(options.ClusterNames, args...)
			}

			if rootCommand.clusterName != "" {
				if len(args) != 0 {
					exitWithError(fmt.Errorf("cannot mix --name for cluster with positional arguments"))
				}

				options.ClusterNames = append(options.ClusterNames, rootCommand.clusterName)
			}

			err := RunGetClusters(&rootCommand, os.Stdout, &options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVar(&options.FullSpec, "full", options.FullSpec, "Show fully populated configuration")

	getCmd.cobraCommand.AddCommand(cmd)
}

func RunGetClusters(context Factory, out io.Writer, options *GetClusterOptions) error {
	client, err := context.Clientset()
	if err != nil {
		return err
	}

	clusterList, err := client.Clusters().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	var clusters []*api.Cluster
	if len(options.ClusterNames) != 0 {
		m := make(map[string]*api.Cluster)
		for i := range clusterList.Items {
			c := &clusterList.Items[i]
			m[c.ObjectMeta.Name] = c
		}
		for _, clusterName := range options.ClusterNames {
			c := m[clusterName]
			if c == nil {
				return fmt.Errorf("cluster not found %q", clusterName)
			}

			clusters = append(clusters, c)
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

	if options.FullSpec {
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
		t.AddColumn("ZONES", func(c *api.Cluster) string {
			zones := sets.NewString()
			for _, s := range c.Spec.Subnets {
				zones.Insert(s.Zone)
			}
			return strings.Join(zones.List(), ",")
		})
		return t.Render(clusters, out, "NAME", "CLOUD", "ZONES")

	case OutputYaml:
		for i, cluster := range clusters {
			if i != 0 {
				_, err = out.Write([]byte("\n\n---\n\n"))
				if err != nil {
					return fmt.Errorf("error writing to stdout: %v", err)
				}
			}
			if err := marshalToWriter(cluster, marshalYaml, out); err != nil {
				return err
			}
		}
		return nil
	case OutputJSON:
		for _, cluster := range clusters {
			if err := marshalToWriter(cluster, marshalJSON, out); err != nil {
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
