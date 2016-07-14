package main

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"os"
	"path/filepath"
)

type CreateInstanceanceGroupCmd struct {
}

var createInstanceanceGroupCmd CreateInstanceanceGroupCmd

func init() {
	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   "Create instancegroup",
		Long:    `Create an instancegroup configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				glog.Exitf("Specify name of instance group to create")
			}
			if len(args) != 1 {
				glog.Exitf("Can only create one instance group at a time!")
			}
			err := createInstanceanceGroupCmd.Run(args[0])
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	createCmd.AddCommand(cmd)
}

func (c *CreateInstanceanceGroupCmd) Run(groupName string) error {
	instanceGroupStore, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	existing, err := instanceGroupStore.Find(groupName)
	if err != nil {
		return err
	}

	if existing != nil {
		return fmt.Errorf("instance group %q already exists", groupName)
	}

	// Populate some defaults
	ig := &api.InstanceGroup{}
	ig.Name = groupName
	ig.Spec.MinSize = fi.Int(2)
	ig.Spec.MaxSize = fi.Int(2)
	ig.Spec.MachineType = "t2.medium"
	ig.Spec.Role = api.InstanceGroupRoleNode

	var (
		edit = editor.NewDefaultEditor(editorEnvs)
	)

	raw, err := api.ToYaml(ig)
	if err != nil {
		return err
	}
	ext := "yaml"

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

	group := &api.InstanceGroup{}
	err = api.ParseYaml(edited, group)
	if err != nil {
		return fmt.Errorf("error parsing yaml: %v", err)
	}

	err = group.Validate(false)
	if err != nil {
		return err
	}

	err = instanceGroupStore.Create(group)
	if err != nil {
		return fmt.Errorf("error storing instancegroup: %v", err)
	}

	return nil
}
