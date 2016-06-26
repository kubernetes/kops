package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/kutil"
)

type UpgradeClusterCmd struct {
	NewClusterName string
}

var upgradeCluster UpgradeClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Upgrade cluster",
		Long:  `Upgrades a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := upgradeCluster.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	upgradeCmd.AddCommand(cmd)

	cmd.Flags().StringVar(&upgradeCluster.NewClusterName, "newname", "", "new cluster name")
}

func (c *UpgradeClusterCmd) Run() error {
	if c.NewClusterName == "" {
		return fmt.Errorf("--newname is required")
	}

	stateStore, err := rootCommand.StateStore()
	if err != nil {
		return fmt.Errorf("error state store: %v", err)
	}

	cluster, nodeSets, err := cloudup.ReadConfig(stateStore)
	if err != nil {
		return fmt.Errorf("error reading configuration: %v", err)
	}

	oldClusterName := cluster.ClusterName
	if oldClusterName == "" {
		return fmt.Errorf("(Old) ClusterName must be set in configuration")
	}

	if len(cluster.Zones) == 0 {
		return fmt.Errorf("Configuration must include Zones")
	}

	region := ""
	for _, zone := range cluster.Zones {
		if len(zone.Name) <= 2 {
			return fmt.Errorf("Invalid AWS zone: %q", zone.Name)
		}

		zoneRegion := zone.Name[:len(zone.Name)-1]
		if region != "" && zoneRegion != region {
			return fmt.Errorf("Clusters cannot span multiple regions")
		}

		region = zoneRegion
	}

	tags := map[string]string{"KubernetesCluster": oldClusterName}
	cloud, err := awsup.NewAWSCloud(region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	d := &kutil.UpgradeCluster{}
	d.NewClusterName = c.NewClusterName
	d.OldClusterName = oldClusterName
	d.Cloud = cloud
	d.ClusterConfig = cluster
	d.NodeSets = nodeSets
	d.StateStore = stateStore

	err = d.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
