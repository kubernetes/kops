package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
)

type GetInstanceGroupsCmd struct {
}

var getInstanceGroupsCmd GetInstanceGroupsCmd

func init() {
	cmd := &cobra.Command{
		Use:     "instancegroups",
		Aliases: []string{"instancegroup", "ig"},
		Short:   "get instancegroups",
		Long:    `List or get InstanceGroups.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getInstanceGroupsCmd.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	getCmd.AddCommand(cmd)
}

func (c *GetInstanceGroupsCmd) Run() error {
	stateStore, err := rootCommand.StateStore()
	if err != nil {
		return err
	}

	_, instancegroups, err := api.ReadConfig(stateStore)
	if err != nil {
		return err
	}

	if len(instancegroups) == 0 {
		return nil
	}

	t := &Table{}
	t.AddColumn("NAME", func(c *api.InstanceGroup) string {
		return c.Name
	})
	return t.Render(instancegroups, os.Stdout)
}
