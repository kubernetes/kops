package main

import (
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api/registry"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/kutil"
)

type ExportKubecfgCommand struct {
	tmpdir   string
	keyStore fi.CAStore
}

var exportKubecfgCommand ExportKubecfgCommand

func init() {
	cmd := &cobra.Command{
		Use:   "kubecfg CLUSTERNAME",
		Short: "Generate a kubecfg file for a cluster",
		Long:  `Creates a kubecfg file for a cluster, based on the state`,
		Run: func(cmd *cobra.Command, args []string) {
			err := exportKubecfgCommand.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	exportCmd.AddCommand(cmd)
}

func (c *ExportKubecfgCommand) Run(args []string) error {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	clusterName := cluster.Name

	master := cluster.Spec.MasterPublicName
	if master == "" {
		master = "api." + clusterName
	}

	x := &kutil.CreateKubecfg{
		ClusterName:      clusterName,
		KeyStore:         keyStore,
		SecretStore:      secretStore,
		MasterPublicName: master,
	}
	defer x.Close()

	return x.WriteKubecfg()
}
