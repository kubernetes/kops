package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
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
	clusterNames, err := rootCommand.ListClusters()
	if err != nil {
		return err
	}

	var clusters []*api.Cluster

	for _, clusterName := range clusterNames {
		stateStore, err := rootCommand.StateStoreForCluster(clusterName)
		if err != nil {
			return err
		}

		// TODO: Faster if we don't read groups...
		// We probably can just have a command which directly reads all cluster config files
		cluster, _, err := api.ReadConfig(stateStore)
		clusters = append(clusters, cluster)
	}
	if len(clusters) == 0 {
		return nil
	}

	t := &Table{}
	t.AddColumn("NAME", func(c *api.Cluster) string {
		return c.Name
	})
	return t.Render(clusters, os.Stdout)
}
