package main

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"os"
	"path/filepath"
)

type EditInstanceGroupCmd struct {
}

var editInstanceGroupCmd EditInstanceGroupCmd

func init() {
	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   "Edit instancegroup",
		Long:    `Edit an instancegroup configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				glog.Exitf("Specify name of instance group to edit")
			}
			if len(args) != 1 {
				glog.Exitf("Can only edit one instance group at a time!")
			}
			err := editInstanceGroupCmd.Run(args[0])
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	editCmd.AddCommand(cmd)
}

func (c *EditInstanceGroupCmd) Run(groupName string) error {
	registry, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	if groupName == "" {
		return fmt.Errorf("name is required")
	}

	group, err := registry.Find(groupName)
	if err != nil {
		return fmt.Errorf("error reading InstanceGroup %q: %v", groupName, err)
	}
	if group == nil {
		return fmt.Errorf("InstanceGroup %q not found", groupName)
	}

	var (
		edit = editor.NewDefaultEditor(editorEnvs)
	)

	ext := "yaml"
	raw, err := api.ToYaml(group)
	if err != nil {
		return fmt.Errorf("error parsing InstanceGroup: %v", err)
	}

	// launch the editor
	edited, file, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ext, bytes.NewReader(raw))
	defer func() {
		if file != "" {
			os.Remove(file)
		}
	}()
	if err != nil {
		return fmt.Errorf("error launching editor: %v", err)
	}

	if bytes.Equal(edited, raw) {
		fmt.Fprintln(os.Stderr, "Edit cancelled, no changes made.")
		return nil
	}

	newGroup := &api.InstanceGroup{}
	err = api.ParseYaml(edited, newGroup)
	if err != nil {
		return fmt.Errorf("error parsing config: %v", err)
	}

	err = newGroup.Validate()
	if err != nil {
		return err
	}

	err = registry.Update(newGroup)
	if err != nil {
		return err
	}

	return nil
}
