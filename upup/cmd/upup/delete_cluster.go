package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/kutil"
)

type DeleteClusterCmd struct {
	ClusterID string
	Yes       bool
	Region    string
}

var deleteCluster DeleteClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Delete cluster",
		Long:  `Deletes a k8s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteCluster.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	deleteCmd.AddCommand(cmd)

	cmd.Flags().BoolVar(&deleteCluster.Yes, "yes", false, "Delete without confirmation")

	cmd.Flags().StringVar(&deleteCluster.ClusterID, "cluster-id", "", "cluster id")
	cmd.Flags().StringVar(&deleteCluster.Region, "region", "", "region")
}

func (c *DeleteClusterCmd) Run() error {
	if c.Region == "" {
		return fmt.Errorf("--region is required")
	}
	if c.ClusterID == "" {
		return fmt.Errorf("--cluster-id is required")
	}

	tags := map[string]string{"KubernetesCluster": c.ClusterID}
	cloud, err := awsup.NewAWSCloud(c.Region, tags)
	if err != nil {
		return fmt.Errorf("error initializing AWS client: %v", err)
	}

	d := &kutil.DeleteCluster{}

	d.ClusterID = c.ClusterID
	d.Region = c.Region
	d.Cloud = cloud

	glog.Infof("TODO: S3 bucket removal")

	resources, err := d.ListResources()
	if err != nil {
		return err
	}

	for k := range resources {
		fmt.Printf("%v\n", k)
	}

	if !c.Yes {
		return fmt.Errorf("Must specify --yes to delete")
	}

	return d.DeleteResources(resources)
}
