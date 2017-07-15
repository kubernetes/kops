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

	"io"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	get_cluster_long = templates.LongDesc(i18n.T(`
	Display one or many cluster resources.`))

	get_cluster_example = templates.Examples(i18n.T(`
	# Get all clusters in a state store
	kops get clusters

	# Get a cluster
	kops get cluster k8s-cluster.example.com`))

	get_cluster_short = i18n.T(`Get one or many clusters.`)

	// Warning for --full.  Since we are not using the template from kubectl
	// we have to have zero white space before the comment characters otherwise
	// output to stdout is going to be off.
	get_cluster_full_warning = i18n.T(`
//
//   WARNING: Do not use a '--full' cluster specification to define a Kubernetes installation.
//   You may experience unexpected behavior and other bugs.  Use only the required elements
//   and any modifications that you require.
//
//   Use the following command to retrieve only the required elements:
//   $ kop get cluster -o yaml
//

`)
)

type GetClusterOptions struct {
	*GetOptions

	// FullSpec determines if we should output the completed (fully populated) spec
	FullSpec bool

	// ClusterNames is a list of cluster names to show; if not specified all clusters will be shown
	ClusterNames []string
}

func NewCmdGetCluster(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := GetClusterOptions{
		GetOptions: getOptions,
	}

	cmd := &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   get_cluster_short,
		Long:    get_cluster_long,
		Example: get_cluster_example,
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

	return cmd
}

func RunGetClusters(context Factory, out io.Writer, options *GetClusterOptions) error {
	client, err := context.Clientset()
	if err != nil {
		return err
	}

	clusterList, err := client.ListClusters(metav1.ListOptions{})
	if err != nil {
		return err
	}

	clusters, err := buildClusters(options.ClusterNames, clusterList)
	if err != nil {
		return err
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

		fmt.Fprint(out, get_cluster_full_warning)
	}

	switch options.output {
	case OutputTable:
		return clusterOutputTable(clusters, out)
	case OutputYaml:
		return clusterOutputYAML(clusters, out)
	case OutputJSON:
		return clusterOutputJson(clusters, out)

	default:
		return fmt.Errorf("Unknown output format: %q", options.output)
	}
}

func buildClusters(args []string, clusterList *api.ClusterList) ([]*api.Cluster, error) {
	var clusters []*api.Cluster
	if len(args) != 0 {
		m := make(map[string]*api.Cluster)
		for i := range clusterList.Items {
			c := &clusterList.Items[i]
			m[c.ObjectMeta.Name] = c
		}
		for _, clusterName := range args {
			c := m[clusterName]
			if c == nil {
				return nil, fmt.Errorf("cluster not found %q", clusterName)
			}

			clusters = append(clusters, c)
		}
	} else {
		for i := range clusterList.Items {
			c := &clusterList.Items[i]
			clusters = append(clusters, c)
		}
	}

	return clusters, nil
}

func clusterOutputTable(clusters []*api.Cluster, out io.Writer) error {
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
}

func clusterOutputJson(clusters []*api.Cluster, out io.Writer) error {
	for _, cluster := range clusters {
		if err := marshalToWriter(cluster, marshalJSON, out); err != nil {
			return err
		}
	}
	return nil
}

func clusterOutputYAML(clusters []*api.Cluster, out io.Writer) error {
	for i, cluster := range clusters {
		if i != 0 {
			if err := writeYAMLSep(out); err != nil {
				return fmt.Errorf("error writing to stdout: %v", err)
			}
		}
		if err := marshalToWriter(cluster, marshalYaml, out); err != nil {
			return err
		}
	}
	return nil
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
