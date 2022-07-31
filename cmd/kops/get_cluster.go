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
	"strings"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getClusterLong = templates.LongDesc(i18n.T(`
	Display one or many cluster resources.`))

	getClusterExample = templates.Examples(i18n.T(`
	# Get all clusters in a state store
	kops get clusters

	# Get a cluster
	kops get cluster k8s-cluster.example.com

	# Get a cluster YAML desired configuration
	kops get cluster k8s-cluster.example.com -o yaml

	# Save a cluster desired configuration to YAML file
	kops get cluster k8s-cluster.example.com -o yaml > cluster-desired-config.yaml
	`))

	getClusterShort = i18n.T(`Get one or many clusters.`)

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
//   $ kops get cluster -o yaml
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
		Use:     "clusters [CLUSTER]...",
		Aliases: []string{"cluster"},
		Short:   getClusterShort,
		Long:    getClusterLong,
		Example: getClusterExample,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				if rootCommand.clusterName != "" {
					return fmt.Errorf("cannot mix --name for cluster with positional arguments")
				}
				options.ClusterNames = append(options.ClusterNames, args...)
			} else if rootCommand.clusterName != "" {
				options.ClusterNames = append(options.ClusterNames, rootCommand.clusterName)
			}

			return nil
		},
		ValidArgsFunction: commandutils.CompleteClusterName(f, false, true),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGetClusters(context.TODO(), f, out, &options)
		},
	}

	cmd.Flags().BoolVar(&options.FullSpec, "full", options.FullSpec, "Show fully populated configuration")

	return cmd
}

func RunGetClusters(ctx context.Context, f commandutils.Factory, out io.Writer, options *GetClusterOptions) error {
	client, err := f.KopsClient()
	if err != nil {
		return err
	}

	singleClusterSelected := false
	var clusterList []*kopsapi.Cluster
	if len(options.ClusterNames) == 1 {
		// Optimization - avoid fetching all clusters if we're only querying one
		singleClusterSelected = true
		cluster, err := client.GetCluster(ctx, options.ClusterNames[0])
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return err
			}
		} else {
			clusterList = append(clusterList, cluster)
		}
	} else {
		list, err := client.ListClusters(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		for i := range list.Items {
			clusterList = append(clusterList, &list.Items[i])
		}
	}

	clusters, err := filterClustersByName(options.ClusterNames, clusterList)
	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		return fmt.Errorf("no clusters found")
	}

	if options.FullSpec {
		var err error
		clusters, err = fullClusterSpecs(clusters)
		if err != nil {
			return err
		}

		fmt.Fprint(out, get_cluster_full_warning)
	}

	var obj []runtime.Object
	if options.Output != OutputTable {
		for _, c := range clusters {
			obj = append(obj, c)
		}
	}

	switch options.Output {
	case OutputTable:
		return clusterOutputTable(clusters, out)
	case OutputYaml:
		return fullOutputYAML(out, obj...)
	case OutputJSON:
		// if singleClusterSelected is true, only a single object is returned
		// otherwise to keep it consistent, always returns an array.
		// Ex: kops get clusters -ojson should will always return an array (even if 1 cluster is available)
		// kops get cluster test.example.com -o json will return a single object (since a specific cluster is selected)
		return fullOutputJSON(out, singleClusterSelected, obj...)
	default:
		return fmt.Errorf("Unknown output format: %q", options.Output)
	}
}

// filterClustersByName returns the clusters matching the specified names.
// If names are specified and no cluster is found with a name, we return an error.
func filterClustersByName(clusterNames []string, clusters []*kopsapi.Cluster) ([]*kopsapi.Cluster, error) {
	if len(clusterNames) != 0 {
		// Build a map as we want to return them in the same order as args
		m := make(map[string]*kopsapi.Cluster)
		for _, c := range clusters {
			m[c.ObjectMeta.Name] = c
		}
		var filtered []*kopsapi.Cluster
		for _, clusterName := range clusterNames {
			c := m[clusterName]
			if c == nil {
				return nil, fmt.Errorf("cluster not found %q", clusterName)
			}

			filtered = append(filtered, c)
		}
		return filtered, nil
	}

	return clusters, nil
}

func clusterOutputTable(clusters []*kopsapi.Cluster, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("NAME", func(c *kopsapi.Cluster) string {
		return c.ObjectMeta.Name
	})
	t.AddColumn("CLOUD", func(c *kopsapi.Cluster) string {
		return string(c.Spec.GetCloudProvider())
	})
	t.AddColumn("ZONES", func(c *kopsapi.Cluster) string {
		zones := sets.NewString()
		for _, s := range c.Spec.Subnets {
			if s.Zone != "" {
				zones.Insert(s.Zone)
			}
		}
		return strings.Join(zones.List(), ",")
	})

	return t.Render(clusters, out, "NAME", "CLOUD", "ZONES")
}

// fullOutputJSON outputs the marshalled JSON of a list of clusters and instance groups.  It will handle
// nils for clusters and instanceGroups slices.
func fullOutputJSON(out io.Writer, singleObject bool, args ...runtime.Object) error {
	if !singleObject {
		if _, err := fmt.Fprint(out, "["); err != nil {
			return err
		}
	}

	for i, arg := range args {
		if i != 0 {
			if _, err := fmt.Fprint(out, ","); err != nil {
				return err
			}
		}
		if err := marshalToWriter(arg, marshalJSON, out); err != nil {
			return err
		}
	}

	if !singleObject {
		if _, err := fmt.Fprint(out, "]"); err != nil {
			return err
		}
	}

	return nil
}

// fullOutputYAML outputs the marshalled JSON of a list of clusters and instance groups.  It will handle
// nils for clusters and instanceGroups slices.
func fullOutputYAML(out io.Writer, args ...runtime.Object) error {
	for i, obj := range args {
		if i != 0 {
			if err := writeYAMLSep(out); err != nil {
				return fmt.Errorf("error writing to stdout: %v", err)
			}
		}
		if err := marshalToWriter(obj, marshalYaml, out); err != nil {
			return err
		}
	}
	return nil
}

func fullClusterSpecs(clusters []*kopsapi.Cluster) ([]*kopsapi.Cluster, error) {
	var fullSpecs []*kopsapi.Cluster
	for _, cluster := range clusters {
		configBase, err := registry.ConfigBase(cluster)
		if err != nil {
			return nil, fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
		}
		configPath := configBase.Join(registry.PathClusterCompleted)
		b, err := configPath.ReadFile()
		if err != nil {
			return nil, fmt.Errorf("error loading Cluster %q: %v", configPath, err)
		}

		o, _, err := kopscodecs.Decode(b, nil)
		if err != nil {
			return nil, fmt.Errorf("error parsing Cluster %q: %v", configPath, err)
		}
		if fullSpec, ok := o.(*kopsapi.Cluster); ok {
			fullSpecs = append(fullSpecs, fullSpec)
		} else {
			return nil, fmt.Errorf("unexpected object type for Cluster %q: %T", configPath, o)
		}
	}
	return fullSpecs, nil
}
