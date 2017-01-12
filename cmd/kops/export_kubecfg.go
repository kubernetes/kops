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
	"github.com/spf13/cobra"
	"k8s.io/kops/pkg/apis/kops/registry"
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

	clusterName := cluster.ObjectMeta.Name

	master := cluster.Spec.MasterPublicName
	if master == "" {
		master = "api." + clusterName
	}

	x := &kutil.CreateKubecfg{
		ContextName:  clusterName,
		KeyStore:     keyStore,
		SecretStore:  secretStore,
		KubeMasterIP: master,
	}

	return x.WriteKubecfg()
}
