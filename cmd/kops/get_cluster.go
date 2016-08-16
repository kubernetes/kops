package main

import (
	"os"

	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
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
	clusterRegistry, err := rootCommand.ClusterRegistry()
	if err != nil {
		return err
	}

	var clusters []*api.Cluster

	clusterNames := args
	if len(args) == 0 {
		clusterNames, err = clusterRegistry.List()
		if err != nil {
			return err
		}
	}

	for _, clusterName := range clusterNames {
		cluster, err := clusterRegistry.Find(clusterName)
		if err != nil {
			return err
		}

		if cluster == nil {
			return fmt.Errorf("cluster not found %q", clusterName)
		}

		clusters = append(clusters, cluster)
	}

	if len(clusters) == 0 {
		fmt.Fprintf(os.Stderr, "No clusters found\n")
		return nil
	}

	output := getCmd.output
	if output == OutputTable {
		t := &Table{}
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
				spec, err := clusterRegistry.ReadCompletedConfig(cluster.Name)
				if err != nil {
					return fmt.Errorf("error reading full cluster spec for %q: %v", cluster.Name, err)
				}
				fullSpecs = append(fullSpecs, spec)
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
