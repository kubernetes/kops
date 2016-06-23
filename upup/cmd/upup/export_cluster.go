package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/kutil"
)

type ExportClusterCmd struct {
	ClusterName string
	Region      string
}

var exportCluster ExportClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Export cluster",
		Long:  `Exports a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := exportCluster.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	exportCmd.AddCommand(cmd)

	cmd.Flags().StringVar(&exportCluster.ClusterName, "name", "", "cluster name")
	cmd.Flags().StringVar(&exportCluster.Region, "region", "", "region")
}

func (c *ExportClusterCmd) Run() error {
	if c.Region == "" {
		return fmt.Errorf("--region is required")
	}
	if c.ClusterName == "" {
		return fmt.Errorf("--name is required")
	}

	tags := map[string]string{"KubernetesCluster": c.ClusterName}
	cloud, err := awsup.NewAWSCloud(c.Region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	stateStore, err := rootCommand.StateStore()
	if err != nil {
		return fmt.Errorf("error state store: %v", err)
	}

	d := &kutil.ExportCluster{}
	d.ClusterName = c.ClusterName
	d.Cloud = cloud
	d.StateStore = stateStore

	err = d.ReverseAWS()
	if err != nil {
		return err
	}

	return nil
}
