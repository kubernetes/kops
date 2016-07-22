package main

import (
	"fmt"

	"bytes"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/editor"
	"os"
	"path/filepath"
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
				glog.Exitf("%v", err)
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

	clusterRegistry, oldCluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	instanceGroupRegistry, err := rootCommand.InstanceGroupRegistry()
	if err != nil {
		return err
	}

	instancegroups, err := instanceGroupRegistry.ReadAll()
	if err != nil {
		return err
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

	fullCluster, err := cloudup.PopulateClusterSpec(newCluster, clusterRegistry)
	if err != nil {
		return err
	}

	err = api.DeepValidate(fullCluster, instancegroups, true)
	if err != nil {
		return err
	}

	// Note we perform as much validation as we can, before writing a bad config
	err = clusterRegistry.Update(newCluster)
	if err != nil {
		return err
	}

	err = clusterRegistry.WriteCompletedConfig(fullCluster)
	if err != nil {
		return fmt.Errorf("error writing completed cluster spec: %v", err)
	}

	return nil
}
