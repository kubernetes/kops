package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/kutil"
	k8sapi "k8s.io/kubernetes/pkg/api"
)

type ConvertImportedCmd struct {
	NewClusterName string

	// Channel is the location of the api.Channel to use for our defaults
	Channel string
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
	cmd.Flags().StringVar(&convertImported.Channel, "channel", api.DefaultChannel, "Channel to use for upgrade")
}

func (c *ConvertImportedCmd) Run() error {
	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	list, err := clientset.InstanceGroups(cluster.Name).List(k8sapi.ListOptions{})
	if err != nil {
		return err
	}
	var instanceGroups []*api.InstanceGroup
	for i := range list.Items {
		instanceGroups = append(instanceGroups, &list.Items[i])
	}

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

	channel, err := api.LoadChannel(c.Channel)
	if err != nil {
		return err
	}

	d := &kutil.ConvertKubeupCluster{
		NewClusterName: c.NewClusterName,
		OldClusterName: oldClusterName,
		Cloud:          cloud,
		ClusterConfig:  cluster,
		InstanceGroups: instanceGroups,
		Clientset:      clientset,
		Channel:        channel,
	}

	err = d.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
