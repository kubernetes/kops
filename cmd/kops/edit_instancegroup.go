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
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/edit"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/pretty"
	"k8s.io/kops/pkg/try"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	editInstancegroupLong = pretty.LongDesc(i18n.T(`Edit a cluster configuration.

	This command changes the instance group desired configuration in the registry.

	To set your preferred editor, you can define the EDITOR environment variable.
	When you have done this, kOps will use the editor that you have set.

	kops edit does not update the cloud resources; to apply the changes use ` + pretty.Bash("kops update cluster") + `.`))

	editInstancegroupExample = templates.Examples(i18n.T(`
	# Edit an instancegroup desired configuration.
	kops edit instancegroup --name k8s-cluster.example.com nodes --state=s3://my-state-store
	`))

	editInstancegroupShort = i18n.T(`Edit instancegroup.`)
)

type EditInstanceGroupOptions struct {
	ClusterName string
	GroupName   string

	// Sets allows setting values directly in the spec.
	Sets []string
	// Unsets allows unsetting values directly in the spec.
	Unsets []string
}

func NewCmdEditInstanceGroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &EditInstanceGroupOptions{}

	cmd := &cobra.Command{
		Use:     "instancegroup INSTANCE_GROUP",
		Aliases: []string{"instancegroups", "ig"},
		Short:   editInstancegroupShort,
		Long:    editInstancegroupLong,
		Example: editInstancegroupExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)

			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify the name of the instance group to edit")
			}

			options.GroupName = args[0]

			if len(args) != 1 {
				return fmt.Errorf("can only edit one instance group at a time")
			}

			return nil
		},
		ValidArgsFunction: completeInstanceGroup(f, nil, nil),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunEditInstanceGroup(context.TODO(), f, out, options)
		},
	}

	if featureflag.SpecOverrideFlag.Enabled() {
		cmd.Flags().StringSliceVar(&options.Sets, "set", options.Sets, "Directly set values in the spec")
		cmd.RegisterFlagCompletionFunc("set", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		})
		cmd.Flags().StringSliceVar(&options.Unsets, "unset", options.Unsets, "Directly unset values in the spec")
		cmd.RegisterFlagCompletionFunc("unset", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		})
	}

	return cmd
}

func RunEditInstanceGroup(ctx context.Context, f *util.Factory, out io.Writer, options *EditInstanceGroupOptions) error {
	groupName := options.GroupName

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	channel, err := cloudup.ChannelForCluster(cluster)
	if err != nil {
		klog.Warningf("%v", err)
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	oldGroup, err := clientset.InstanceGroupsFor(cluster).Get(ctx, groupName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error reading InstanceGroup %q: %v", groupName, err)
	}
	if oldGroup == nil {
		return fmt.Errorf("InstanceGroup %q not found", groupName)
	}

	if len(options.Unsets)+len(options.Sets) > 0 {
		newGroup := oldGroup.DeepCopy()
		if err := commands.UnsetInstancegroupFields(options.Unsets, newGroup); err != nil {
			return err
		}
		if err := commands.SetInstancegroupFields(options.Sets, newGroup); err != nil {
			return err
		}

		failure, err := updateInstanceGroup(ctx, clientset, channel, cluster, newGroup)
		if err != nil {
			return err
		}
		if failure != "" {
			return fmt.Errorf("%s", failure)
		}
		return nil
	}

	editor := editor.NewDefaultEditor(commandutils.EditorEnvs)

	ext := "yaml"
	raw, err := kopscodecs.ToVersionedYaml(oldGroup)
	if err != nil {
		return err
	}

	var (
		results = editResults{}
		edited  = []byte{}
		file    string
	)

	containsError := false

	for {
		buf := &bytes.Buffer{}
		results.header.writeTo(buf)
		results.header.flush()

		if !containsError {
			buf.Write(raw)
		} else {
			buf.Write(stripComments(edited))
		}

		// launch the editor
		editedDiff := edited
		edited, file, err = editor.LaunchTempFile(fmt.Sprintf("%s-edit-", filepath.Base(os.Args[0])), ext, buf)
		if err != nil {
			return preservedFile(fmt.Errorf("error launching editor: %v", err), results.file, out)
		}

		if containsError {
			if bytes.Equal(stripComments(editedDiff), stripComments(edited)) {
				return preservedFile(fmt.Errorf("%s", "Edit cancelled: no valid changes were saved."), file, out)
			}
		}

		if len(results.file) > 0 {
			try.RemoveFile(results.file)
		}

		if bytes.Equal(stripComments(raw), stripComments(edited)) {
			try.RemoveFile(file)
			fmt.Fprintln(out, "Edit cancelled: no changes made.")
			return nil
		}

		lines, err := hasLines(bytes.NewBuffer(edited))
		if err != nil {
			return preservedFile(err, file, out)
		}
		if !lines {
			try.RemoveFile(file)
			fmt.Fprintln(out, "Edit cancelled: saved file was empty.")
			return nil
		}

		newObj, _, err := kopscodecs.Decode(edited, nil)
		if err != nil {
			return preservedFile(fmt.Errorf("error parsing InstanceGroup: %v", err), file, out)
		}

		newGroup, ok := newObj.(*api.InstanceGroup)
		if !ok {
			results = editResults{
				file: file,
			}
			results.header.addError(fmt.Sprintf("object was not of expected type: %T", newObj))
			containsError = true
			continue
		}

		extraFields, err := edit.HasExtraFields(string(edited), newObj)
		if err != nil {
			results = editResults{
				file: file,
			}
			results.header.addError(fmt.Sprintf("error checking for extra fields: %v", err))
			containsError = true
			continue
		}
		if extraFields != "" {
			results = editResults{
				file: file,
			}
			lines := strings.Split(extraFields, "\n")
			for _, line := range lines {
				results.header.addExtraFields(line)
			}
			containsError = true
			continue
		}

		failure, err := updateInstanceGroup(ctx, clientset, channel, cluster, newGroup)
		if err != nil {
			return preservedFile(err, file, out)
		}
		if failure != "" {
			results = editResults{
				file: file,
			}
			results.header.addError(failure)
			containsError = true
			continue
		}

		return nil
	}
}

func updateInstanceGroup(ctx context.Context, clientset simple.Clientset, channel *api.Channel, cluster *api.Cluster, newGroup *api.InstanceGroup) (string, error) {
	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return "", err
	}

	fullGroup, err := cloudup.PopulateInstanceGroupSpec(cluster, newGroup, cloud, channel)
	if err != nil {
		return fmt.Sprintf("error populating instance group spec: %s", err), nil
	}

	// We need the full cluster spec to perform deep validation
	// Note that we don't write it back though
	err = cloudup.PerformAssignments(cluster, cloud)
	if err != nil {
		return "", fmt.Errorf("error populating configuration: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(cluster, false)
	fullCluster, err := cloudup.PopulateClusterSpec(clientset, cluster, cloud, assetBuilder)
	if err != nil {
		return fmt.Sprintf("error populating cluster spec: %s", err), nil
	}

	err = validation.CrossValidateInstanceGroup(fullGroup, fullCluster, cloud, true).ToAggregate()
	if err != nil {
		return fmt.Sprintf("validation failed: %s", err), nil
	}

	// Note we perform as much validation as we can, before writing a bad config
	_, err = clientset.InstanceGroupsFor(cluster).Update(ctx, newGroup, metav1.UpdateOptions{})
	return "", err
}
