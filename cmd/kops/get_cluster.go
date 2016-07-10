package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"strings"
)

type GetClustersCmd struct {
}

var getClustersCmd GetClustersCmd

func init() {
	cmd := &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   "get clusters",
		Long:    `List or get clusters.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getClustersCmd.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	getCmd.AddCommand(cmd)
}

func (c *GetClustersCmd) Run() error {
	clusterRegistry, err := rootCommand.ClusterRegistry()
	if err != nil {
		return err
	}

	clusterNames, err := clusterRegistry.List()
	if err != nil {
		return err
	}

	var clusters []*api.Cluster
	for _, clusterName := range clusterNames {
		cluster, err := clusterRegistry.Find(clusterName)
		if err != nil {
			return err
		}

		if cluster == nil {
			glog.Warningf("cluster was listed, but then not found %q", clusterName)
		}

		clusters = append(clusters, cluster)
	}
	if len(clusters) == 0 {
		return nil
	}

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
}
