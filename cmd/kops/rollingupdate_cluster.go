package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/kutil"
)

type RollingUpdateClusterCmd struct {
	Yes   bool
	Force bool

	cobraCommand *cobra.Command
}

var rollingupdateCluster = RollingUpdateClusterCmd{
	cobraCommand: &cobra.Command{
		Use:   "cluster",
		Short: "rolling-update cluster",
		Long:  `rolling-updates a k8s cluster.`,
	},
}

func init() {
	cmd := rollingupdateCluster.cobraCommand
	rollingUpdateCommand.cobraCommand.AddCommand(cmd)

	cmd.Flags().BoolVar(&rollingupdateCluster.Yes, "yes", false, "perform rolling update without confirmation")
	cmd.Flags().BoolVar(&rollingupdateCluster.Force, "force", false, "Force rolling update, even if no changes")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := rollingupdateCluster.Run()
		if err != nil {
			exitWithError(err)
		}
	}
}

func (c *RollingUpdateClusterCmd) Run() error {
	_, cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	instanceGroupRegistry, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	instancegroups, err := instanceGroupRegistry.ReadAll()
	if err != nil {
		return err
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	d := &kutil.RollingUpdateCluster{}
	d.Cloud = cloud

	warnUnmatched := true
	groups, err := kutil.FindCloudInstanceGroups(cloud, cluster, instancegroups, warnUnmatched)
	if err != nil {
		return err
	}

	{
		t := &Table{}
		t.AddColumn("NAME", func(r *kutil.CloudInstanceGroup) string {
			return r.InstanceGroup.Name
		})
		t.AddColumn("STATUS", func(r *kutil.CloudInstanceGroup) string {
			return r.Status
		})
		t.AddColumn("NEEDUPDATE", func(r *kutil.CloudInstanceGroup) string {
			return strconv.Itoa(len(r.NeedUpdate))
		})
		t.AddColumn("READY", func(r *kutil.CloudInstanceGroup) string {
			return strconv.Itoa(len(r.Ready))
		})
		t.AddColumn("MIN", func(r *kutil.CloudInstanceGroup) string {
			return strconv.Itoa(r.MinSize())
		})
		t.AddColumn("MAX", func(r *kutil.CloudInstanceGroup) string {
			return strconv.Itoa(r.MaxSize())
		})
		var l []*kutil.CloudInstanceGroup
		for _, v := range groups {
			l = append(l, v)
		}

		err := t.Render(l, os.Stdout, "NAME", "STATUS", "NEEDUPDATE", "READY", "MIN", "MAX")
		if err != nil {
			return err
		}
	}

	needUpdate := false
	for _, group := range groups {
		if len(group.NeedUpdate) != 0 {
			needUpdate = true
		}
	}

	if !needUpdate && !c.Force {
		fmt.Printf("\nNo rolling-update required\n")
		return nil
	}

	if !c.Yes {
		fmt.Printf("\nMust specify --yes to rolling-update\n")
		return nil
	}

	return d.RollingUpdate(groups, c.Force)
}
