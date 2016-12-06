/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"os"
	"path/filepath"
)

type CreateInstanceGroupOptions struct {
}

func NewCmdCreateInstanceGroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateInstanceGroupOptions{}

	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   "Create instancegroup",
		Long:    `Create an instancegroup configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCreateInstanceGroup(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunCreateInstanceGroup(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, c *CreateInstanceGroupOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("Specify name of instance group to create")
	}
	if len(args) != 1 {
		return fmt.Errorf("Can only create one instance group at a time!")
	}
	groupName := args[0]

	cluster, err := rootCommand.Cluster()

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	channel, err := cloudup.ChannelForCluster(cluster)
	if err != nil {
		return err
	}

	existing, err := clientset.InstanceGroups(cluster.ObjectMeta.Name).Get(groupName)
	if err != nil {
		return err
	}

	if existing != nil {
		return fmt.Errorf("instance group %q already exists", groupName)
	}

	// Populate some defaults
	ig := &api.InstanceGroup{}
	ig.ObjectMeta.Name = groupName
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

	err = group.Validate()
	if err != nil {
		return err
	}

	_, err = clientset.InstanceGroups(cluster.ObjectMeta.Name).Create(group)
	if err != nil {
		return fmt.Errorf("error storing InstanceGroup: %v", err)
	}

	return nil
}
