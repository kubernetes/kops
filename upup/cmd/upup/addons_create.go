package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
)

type AddonsCreateCmd struct {
	cobraCommand *cobra.Command
}

var addonsCreateCmd = AddonsCreateCmd{
	cobraCommand: &cobra.Command{
		Use:   "create",
		Short: "Create an addons",
		Long:  `Create an addon in a cluster.`,
	},
}

func init() {
	cmd := addonsCreateCmd.cobraCommand
	addonsCmd.cobraCommand.AddCommand(cmd)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := addonsCreateCmd.Run(args)
		if err != nil {
			glog.Exitf("%v", err)
		}
	}
}

func (c *AddonsCreateCmd) Run(args []string) error {
	k, err := addonsCmd.buildClusterAddons()
	if err != nil {
		return err
	}

	addonFiles := make(map[string][]vfs.Path)

	for _, path := range args {
		vfsPath := vfs.NewFSPath(path)

		files, err := vfsPath.ReadDir()
		if err != nil {
			return fmt.Errorf("error listing path %s: %v", vfsPath, err)
		}

		key := vfsPath.Base()
		addonFiles[key] = files
	}

	for key, files := range addonFiles {
		glog.Infof("Creating addon %q", key)
		err := k.CreateAddon(key, files)
		if err != nil {
			return fmt.Errorf("error creating addon %q: %v", key, err)
		}
	}

	return nil
}
