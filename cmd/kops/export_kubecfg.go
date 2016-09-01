package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
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

	clusterRegistry, cluster, err := rootCommand.Cluster()
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
		KeyStore:         clusterRegistry.KeyStore(clusterName),
		SecretStore:      clusterRegistry.SecretStore(cluster.Name),
		MasterPublicName: master,
	}
	defer x.Close()

	return x.WriteKubecfg()
}
