package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kube-deploy/upup/pkg/kutil"
	"time"
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

	for _, r := range resources {
		fmt.Printf("%v\n", r)
	}

	if !c.Yes {
		return fmt.Errorf("Must specify --yes to delete")
	}

	for {
		// TODO: Parallel delete
		// TODO: Some form of ordering?
		// TODO: Give up eventually?

		var failed []kutil.DeletableResource
		for _, r := range resources {
			fmt.Printf("Deleting resource %s:  ", r)
			err := r.Delete(cloud)
			if err != nil {
				if kutil.IsDependencyViolation(err) {
					fmt.Printf("still has dependencies, will retry\n")
				} else {
					fmt.Printf("error deleting resource, will retry: %v\n", err)
				}
				failed = append(failed, r)
			} else {
				fmt.Printf(" ok\n")
			}
		}

		resources = failed
		if len(resources) == 0 {
			break
		}

		fmt.Printf("Not all resources deleted; waiting before reattempting deletion\n")
		for _, r := range resources {
			fmt.Printf("\t%s\n", r)
		}
		time.Sleep(10 * time.Second)
	}

	return nil
}
