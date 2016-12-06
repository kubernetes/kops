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
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
)

type EditClusterOptions struct {
}

func NewCmdEditCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &EditClusterOptions{}

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Edit cluster",
		Long:  `Edit a cluster configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunEditCluster(f, cmd, args, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunEditCluster(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *EditClusterOptions) error {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	oldCluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	err = oldCluster.FillDefaults()
	if err != nil {
		return err
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	list, err := clientset.InstanceGroups(oldCluster.ObjectMeta.Name).List(k8sapi.ListOptions{})
	if err != nil {
		return err
	}
	var instancegroups []*api.InstanceGroup
	for i := range list.Items {
		instancegroups = append(instancegroups, &list.Items[i])
	}

	var (
		edit = editor.NewDefaultEditor(editorEnvs)
	)

	ext := "yaml"
	raw, err := api.ToYaml(oldCluster)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
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

	newCluster := &api.Cluster{}
	err = api.ParseYaml(edited, newCluster)
	if err != nil {
		return fmt.Errorf("error parsing config: %v", err)
	}

	err = newCluster.PerformAssignments()
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	fullCluster, err := cloudup.PopulateClusterSpec(newCluster)
	if err != nil {
		return err
	}

	err = api.DeepValidate(fullCluster, instancegroups, true)
	if err != nil {
		return err
	}

	configBase, err := registry.ConfigBase(newCluster)
	if err != nil {
		return err
	}

	// Note we perform as much validation as we can, before writing a bad config
	_, err = clientset.Clusters().Update(newCluster)
	if err != nil {
		return err
	}

	err = registry.WriteConfigDeprecated(configBase.Join(registry.PathClusterCompleted), fullCluster)
	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	return nil
}
