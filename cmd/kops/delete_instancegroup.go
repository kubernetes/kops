package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/kutil"
)

type DeleteInstanceceGroupCmd struct {
}

var deleteInstanceceGroupCmd DeleteInstanceceGroupCmd

func init() {
	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   "Delete instancegroup",
		Long:    `Delete an instancegroup configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				glog.Exitf("Specify name of instance group to delete")
			}
			if len(args) != 1 {
				glog.Exitf("Can only edit one instance group at a time!")
			}
			err := deleteInstanceceGroupCmd.Run(args[0])
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	deleteCmd.AddCommand(cmd)
}

func (c *DeleteInstanceceGroupCmd) Run(groupName string) error {
	if groupName == "" {
		return fmt.Errorf("name is required")
	}

	_, cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	registry, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	group, err := registry.Find(groupName)
	if err != nil {
		return fmt.Errorf("error reading InstanceGroup %q: %v", groupName, err)
	}
	if group == nil {
		return fmt.Errorf("InstanceGroup %q not found", groupName)
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	d := &kutil.DeleteInstanceGroup{}
	d.Cluster = cluster
	d.Cloud = cloud
	d.InstanceGroupRegistry = registry

	err = d.DeleteInstanceGroup(group)
	if err != nil {
		return err
	}

	fmt.Printf("InstanceGroup %q deleted\n", group.Name)

	return nil
}
