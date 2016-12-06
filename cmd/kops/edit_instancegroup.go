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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
)

type EditInstanceGroupOptions struct {
}

func NewCmdEditInstanceGroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &EditInstanceGroupOptions{}

	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   "Edit instancegroup",
		Long:    `Edit an instancegroup configuration.`,
		Run: func(cmd *cobra.Command, args []string) {

			err := RunEditInstanceGroup(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunEditInstanceGroup(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *EditInstanceGroupOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("Specify name of instance group to edit")
	}
	if len(args) != 1 {
		return fmt.Errorf("Can only edit one instance group at a time")
	}

	groupName := args[0]

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	channel, err := cloudup.ChannelForCluster(cluster)
	if err != nil {
		return err
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	if groupName == "" {
		return fmt.Errorf("name is required")
	}

	oldGroup, err := clientset.InstanceGroups(cluster.ObjectMeta.Name).Get(groupName)
	if err != nil {
		return fmt.Errorf("error reading InstanceGroup %q: %v", groupName, err)
	}
	if oldGroup == nil {
		return fmt.Errorf("InstanceGroup %q not found", groupName)
	}

	var (
		edit = editor.NewDefaultEditor(editorEnvs)
	)

	ext := "yaml"
	raw, err := api.ToYaml(oldGroup)
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

	fullGroup, err := cloudup.PopulateInstanceGroupSpec(cluster, newGroup, channel)
	if err != nil {
		return err
	}

	// We need the full cluster spec to perform deep validation
	// Note that we don't write it back though
	err = cluster.PerformAssignments()
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	fullCluster, err := cloudup.PopulateClusterSpec(cluster)
	if err != nil {
		return err
	}

	err = fullGroup.CrossValidate(fullCluster, true)
	if err != nil {
		return err
	}

	// Note we perform as much validation as we can, before writing a bad config
	_, err = clientset.InstanceGroups(cluster.ObjectMeta.Name).Update(fullGroup)
	if err != nil {
		return err
	}

	return nil
}
