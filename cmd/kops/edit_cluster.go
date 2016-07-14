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

type EditClusterCmd struct {
}

var editClusterCmd EditClusterCmd

func init() {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Edit cluster",
		Long:  `Edit a cluster configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := editClusterCmd.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	editCmd.AddCommand(cmd)
}

func (c *EditClusterCmd) Run() error {
	clusterRegistry, cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	var (
		edit = editor.NewDefaultEditor(editorEnvs)
	)

	ext := "yaml"
	raw, err := api.ToYaml(cluster)
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

	err = newCluster.Validate(false)
	if err != nil {
		return err
	}

	err = clusterRegistry.Update(newCluster)
	if err != nil {
		return err
	}

	return nil
}
