package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/kutil"
)

type ImportClusterCmd struct {
	Region string
}

var importCluster ImportClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Import existing cluster into the state store",
		Long:  `Imports the settings of an existing k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := importCluster.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	importCmd.AddCommand(cmd)

	cmd.Flags().StringVar(&importCluster.Region, "region", "", "region")
}

func (c *ImportClusterCmd) Run() error {
	if c.Region == "" {
		return fmt.Errorf("--region is required")
	}
	clusterName := rootCommand.clusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	tags := map[string]string{"KubernetesCluster": clusterName}
	cloud, err := awsup.NewAWSCloud(c.Region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	stateStore, err := rootCommand.StateStore()
	if err != nil {
		return fmt.Errorf("error state store: %v", err)
	}

	d := &kutil.ImportCluster{}
	d.ClusterName = clusterName
	d.Cloud = cloud
	d.StateStore = stateStore

	err = d.ImportAWSCluster()
	if err != nil {
		return err
	}

	fmt.Printf("\nImported settings for cluster %q\n", clusterName)

	return nil
}
