package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"path/filepath"
)

type CreateInstanceGroupCmd struct {
}

var createInstanceGroupCmd CreateInstanceGroupCmd

func init() {
	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   "Create instancegroup",
		Long:    `Create an instancegroup configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				exitWithError(fmt.Errorf("Specify name of instance group to create"))
			}
			if len(args) != 1 {
				exitWithError(fmt.Errorf("Can only create one instance group at a time!"))
			}
			err := createInstanceGroupCmd.Run(args[0])
			if err != nil {
				exitWithError(err)
			}
		},
	}

	createCmd.AddCommand(cmd)
}

func (c *CreateInstanceGroupCmd) Run(groupName string) error {
	cluster, err := rootCommand.Cluster()

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	channel, err := cloudup.ChannelForCluster(cluster)
	if err != nil {
		return err
	}

	existing, err := clientset.InstanceGroups(cluster.Name).Get(groupName)
	if err != nil {
		return err
	}

	if existing != nil {
		return fmt.Errorf("instance group %q already exists", groupName)
	}

	// Populate some defaults
	ig := &api.InstanceGroup{}
	ig.Name = groupName
	ig.Spec.Role = api.InstanceGroupRoleNode

	ig, err = cloudup.PopulateInstanceGroupSpec(cluster, ig, channel)
	if err != nil {
		return err
	}

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

	_, err = clientset.InstanceGroups(cluster.Name).Create(group)
	if err != nil {
		return fmt.Errorf("error storing InstanceGroup: %v", err)
	}

	return nil
}
