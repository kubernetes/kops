package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"strconv"
	"strings"
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
	registry, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	instancegroups, err := registry.ReadAll()
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
	t.AddColumn("ROLE", func(c *api.InstanceGroup) string {
		return string(c.Spec.Role)
	})
	t.AddColumn("MACHINETYPE", func(c *api.InstanceGroup) string {
		return c.Spec.MachineType
	})
	t.AddColumn("ZONES", func(c *api.InstanceGroup) string {
		return strings.Join(c.Spec.Zones, ",")
	})
	t.AddColumn("MIN", func(c *api.InstanceGroup) string {
		return intPointerToString(c.Spec.MinSize)
	})
	t.AddColumn("MAX", func(c *api.InstanceGroup) string {
		return intPointerToString(c.Spec.MinSize)
	})
	return t.Render(instancegroups, os.Stdout, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "ZONES")
}

func intPointerToString(v *int) string {
	if v == nil {
		return "-"
	}
	return strconv.Itoa(*v)
}
