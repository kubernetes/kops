package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/api/registry"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
)

type EditClusterCmd struct {
}

var editClusterCmd EditClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Edit cluster",
		Long:  `Edit a cluster configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := editClusterCmd.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	editCmd.AddCommand(cmd)
}

func (c *EditClusterCmd) Run(args []string) error {
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

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	list, err := clientset.InstanceGroups(oldCluster.Name).List(k8sapi.ListOptions{})
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

	err = registry.WriteConfig(configBase.Join(registry.PathClusterCompleted), fullCluster)
	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	return nil
}
