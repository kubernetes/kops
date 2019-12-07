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
	"fmt"
	"io"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	exportKubecfgLong = templates.LongDesc(i18n.T(`
	Export a kubecfg file for a cluster from the state store. The configuration
	will be saved into a users $HOME/.kube/config file.
	To export the kubectl configuration to a specific file set the KUBECONFIG
	environment variable.`))

	exportKubecfgExample = templates.Examples(i18n.T(`
	# export a kubecfg file
	kops export kubecfg kubernetes-cluster.example.com
		`))

	exportKubecfgShort = i18n.T(`Export kubecfg.`)
)

type ExportKubecfgOptions struct {
	KubeConfigPath string
	all            bool
}

func NewCmdExportKubecfg(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ExportKubecfgOptions{}

	cmd := &cobra.Command{
		Use:     "kubecfg CLUSTERNAME",
		Short:   exportKubecfgShort,
		Long:    exportKubecfgLong,
		Example: exportKubecfgExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunExportKubecfg(f, out, options, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.KubeConfigPath, "kubeconfig", options.KubeConfigPath, "The location of the kubeconfig file to create.")
	cmd.Flags().BoolVar(&options.all, "all", options.all, "export all clusters from the kops state store")

	return cmd
}

func RunExportKubecfg(f *util.Factory, out io.Writer, options *ExportKubecfgOptions, args []string) error {
	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	var clusterList []*api.Cluster
	if options.all {
		if len(args) != 0 {
			return fmt.Errorf("Cannot use both --all flag and positional arguments")
		}
		list, err := clientset.ListClusters(metav1.ListOptions{})
		if err != nil {
			return err
		}
		for i := range list.Items {
			clusterList = append(clusterList, &list.Items[i])
		}
	} else {
		err := rootCommand.ProcessArgs(args)
		if err != nil {
			return err
		}
		cluster, err := rootCommand.Cluster()
		if err != nil {
			return err
		}
		clusterList = append(clusterList, cluster)
	}

	for _, cluster := range clusterList {
		keyStore, err := clientset.KeyStore(cluster)
		if err != nil {
			return err
		}

		secretStore, err := clientset.SecretStore(cluster)
		if err != nil {
			return err
		}

		conf, err := kubeconfig.BuildKubecfg(cluster, keyStore, secretStore, &commands.CloudDiscoveryStatusStore{}, buildPathOptions(options))
		if err != nil {
			return err
		}

		if err := conf.WriteKubecfg(); err != nil {
			return err
		}
	}

	return nil
}

func buildPathOptions(options *ExportKubecfgOptions) *clientcmd.PathOptions {
	pathOptions := clientcmd.NewDefaultPathOptions()

	if len(options.KubeConfigPath) > 0 {
		pathOptions.GlobalFile = options.KubeConfigPath
		pathOptions.EnvVar = ""
		pathOptions.GlobalFileSubpath = ""
	}

	return pathOptions
}
