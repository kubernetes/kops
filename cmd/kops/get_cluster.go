package main

import (
	"os"

	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/api/registry"
	"k8s.io/kops/util/pkg/tables"
	k8sapi "k8s.io/kubernetes/pkg/api"
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
	for i := range clusterList.Items {
		clusters = append(clusters, &clusterList.Items[i])
	}
	if len(clusters) == 0 {
		fmt.Fprintf(os.Stderr, "No clusters found\n")
		return nil
	}

	output := getCmd.output
	if output == OutputTable {
		t := &tables.Table{}
		t.AddColumn("NAME", func(c *api.Cluster) string {
			return c.Name
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
					return fmt.Errorf("error reading full cluster spec for %q: %v", cluster.Name, err)
				}
				fullSpec := &api.Cluster{}
				err = registry.ReadConfig(configBase.Join(registry.PathClusterCompleted), fullSpec)
				if err != nil {
					return fmt.Errorf("error reading full cluster spec for %q: %v", cluster.Name, err)
				}
				fullSpecs = append(fullSpecs, fullSpec)
			}
			clusters = fullSpecs
		}

		for _, cluster := range clusters {
			y, err := api.ToYaml(cluster)
			if err != nil {
				return fmt.Errorf("error marshaling yaml for %q: %v", cluster.Name, err)
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
