package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/kutil"
	"os"
)

type DeleteClusterCmd struct {
	Yes      bool
	Region   string
	External bool
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

	cmd.Flags().BoolVar(&deleteCluster.External, "external", false, "Delete an external cluster")

	cmd.Flags().StringVar(&deleteCluster.Region, "region", "", "region")
}

type getter func(o interface{}) interface{}

func (c *DeleteClusterCmd) Run() error {
	var stateStore fi.StateStore
	var err error

	var cloud fi.Cloud
	clusterName := ""
	region := ""
	if c.External {
		region = c.Region
		if region == "" {
			return fmt.Errorf("--region is required")
		}
		clusterName := rootCommand.clusterName
		if clusterName == "" {
			return fmt.Errorf("--name is required (when --external)")
		}

		tags := map[string]string{"KubernetesCluster": clusterName}
		cloud, err = awsup.NewAWSCloud(c.Region, tags)
		if err != nil {
			return fmt.Errorf("error initializing AWS client: %v", err)
		}
	} else {
		stateStore, err = rootCommand.StateStore()
		if err != nil {
			return err
		}

		cluster, _, err := api.ReadConfig(stateStore)
		if err != nil {
			return err
		}

		if rootCommand.clusterName != cluster.Name {
			return fmt.Errorf("sanity check failed: cluster name mismatch")
		}
		clusterName = cluster.Name

		cloud, err = cloudup.BuildCloud(cluster)
		if err != nil {
			return err
		}
	}

	d := &kutil.DeleteCluster{}
	d.ClusterName = clusterName
	d.Region = region
	d.Cloud = cloud

	resources, err := d.ListResources()
	if err != nil {
		return err
	}

	if len(resources) == 0 {
		fmt.Printf("Nothing to delete\n")
	} else {
		t := &Table{}
		t.AddColumn("TYPE", func(r *kutil.ResourceTracker) string {
			return r.Type
		})
		t.AddColumn("ID", func(r *kutil.ResourceTracker) string {
			return r.ID
		})
		t.AddColumn("NAME", func(r *kutil.ResourceTracker) string {
			return r.Name
		})
		var l []*kutil.ResourceTracker
		for _, v := range resources {
			l = append(l, v)
		}
		err := t.Render(l, os.Stdout)
		if err != nil {
			return err
		}

		if !c.Yes {
			return fmt.Errorf("Must specify --yes to delete")
		}

		err = d.DeleteResources(resources)
		if err != nil {
			return err
		}
	}

	if stateStore != nil {
		if !c.Yes {
			return fmt.Errorf("Must specify --yes to delete")
		}
		err := api.DeleteConfig(stateStore)
		if err != nil {
			return fmt.Errorf("error removing cluster from state store: %v", err)
		}
	}

	fmt.Printf("\nCluster deleted\n")

	return nil
}
