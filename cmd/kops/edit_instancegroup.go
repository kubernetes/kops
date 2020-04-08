/*
Copyright 2019 The Kubernetes Authors.

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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/try"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	editInstancegroupLong = templates.LongDesc(i18n.T(`Edit a cluster configuration.

	This command changes the instancegroup desired configuration in the registry.

    	To set your preferred editor, you can define the EDITOR environment variable.
    	When you have done this, kops will use the editor that you have set.

	kops edit does not update the cloud resources, to apply the changes use "kops update cluster".`))

	editInstancegroupExample = templates.Examples(i18n.T(`
	# Edit an instancegroup desired configuration.
	kops edit ig --name k8s-cluster.example.com nodes --state=s3://kops-state-1234
	`))

	editInstancegroupShort = i18n.T(`Edit instancegroup.`)
)

type EditInstanceGroupOptions struct {
}

func NewCmdEditInstanceGroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &EditInstanceGroupOptions{}

	cmd := &cobra.Command{
		Use:     "instancegroup",
		Aliases: []string{"instancegroups", "ig"},
		Short:   editInstancegroupShort,
		Long:    editInstancegroupLong,
		Example: editInstancegroupExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			err := RunEditInstanceGroup(ctx, f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunEditInstanceGroup(ctx context.Context, f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *EditInstanceGroupOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("Specify name of instance group to edit")
	}
	if len(args) != 1 {
		return fmt.Errorf("Can only edit one instance group at a time")
	}

	groupName := args[0]

	cluster, err := rootCommand.Cluster(ctx)
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

	oldGroup, err := clientset.InstanceGroupsFor(cluster).Get(ctx, groupName, metav1.GetOptions{})
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
	raw, err := kopscodecs.ToVersionedYaml(oldGroup)
	if err != nil {
		return err
	}

	// launch the editor
	edited, file, err := edit.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ext, bytes.NewReader(raw))
	defer func() {
		if file != "" {
			try.RemoveFile(file)
		}
	}()
	if err != nil {
		return fmt.Errorf("error launching editor: %v", err)
	}

	if bytes.Equal(edited, raw) {
		fmt.Fprintln(os.Stderr, "Edit cancelled, no changes made.")
		return nil
	}

	newObj, _, err := kopscodecs.Decode(edited, nil)
	if err != nil {
		return fmt.Errorf("error parsing InstanceGroup: %v", err)
	}

	newGroup, ok := newObj.(*api.InstanceGroup)
	if !ok {
		return fmt.Errorf("object was not of expected type: %T", newObj)
	}

	err = validation.ValidateInstanceGroup(newGroup).ToAggregate()
	if err != nil {
		return err
	}

	fullGroup, err := cloudup.PopulateInstanceGroupSpec(cluster, newGroup, channel)
	if err != nil {
		return err
	}

	// We need the full cluster spec to perform deep validation
	// Note that we don't write it back though
	err = cloudup.PerformAssignments(cluster)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	fullCluster, err := cloudup.PopulateClusterSpec(clientset, cluster, assetBuilder)
	if err != nil {
		return err
	}

	err = validation.CrossValidateInstanceGroup(fullGroup, fullCluster, true).ToAggregate()
	if err != nil {
		return err
	}

	// Note we perform as much validation as we can, before writing a bad config
	_, err = clientset.InstanceGroupsFor(cluster).Update(ctx, fullGroup, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
