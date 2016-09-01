package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/kutil"
)

type ConvertImportedCmd struct {
	NewClusterName string
}

var convertImported ConvertImportedCmd

func init() {
	cmd := &cobra.Command{
		Use:   "convert-imported",
		Short: "Convert an imported cluster into a kops cluster",
		Run: func(cmd *cobra.Command, args []string) {
			err := convertImported.Run()
			if err != nil {
				exitWithError(err)
			}
		},
	}

	toolboxCmd.AddCommand(cmd)

	cmd.Flags().StringVar(&convertImported.NewClusterName, "newname", "", "new cluster name")
}

func (c *ConvertImportedCmd) Run() error {
	clusterRegistry, cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	instanceGroupRegistry, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	instanceGroups, err := instanceGroupRegistry.ReadAll()

	if cluster.Annotations[api.AnnotationNameManagement] != api.AnnotationValueManagementImported {
		return fmt.Errorf("cluster %q does not appear to be a cluster imported using kops import", cluster.Name)
	}

	if c.NewClusterName == "" {
		return fmt.Errorf("--newname is required for converting an imported cluster")
	}

	oldClusterName := cluster.Name
	if oldClusterName == "" {
		return fmt.Errorf("(Old) ClusterName must be set in configuration")
	}

	// TODO: Switch to cloudup.BuildCloud
	if len(cluster.Spec.Zones) == 0 {
		return fmt.Errorf("Configuration must include Zones")
	}

	region := ""
	for _, zone := range cluster.Spec.Zones {
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

	d := &kutil.ConvertKubeupCluster{}
	d.NewClusterName = c.NewClusterName
	d.OldClusterName = oldClusterName
	d.Cloud = cloud
	d.ClusterConfig = cluster
	d.InstanceGroups = instanceGroups
	d.ClusterRegistry = clusterRegistry

	err = d.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
