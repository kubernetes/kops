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
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	export_kubecfg_long = templates.LongDesc(i18n.T(`
	Export a kubecfg file for a cluster from the state store. The configuration
	will be saved into a users $HOME/.kube/config file.
	To export the kubectl configuration to a specific file set the KUBECONFIG
	environment variable.`))

	export_kubecfg_example = templates.Examples(i18n.T(`
	# export a kubecfg file
	kops export kubecfg kubernetes-cluster.example.com
		`))

	export_kubecfg_short = i18n.T(`Export kubecfg.`)
)

type ExportKubecfgOptions struct {
	tmpdir   string
	keyStore fi.CAStore
}

func NewCmdExportKubecfg(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ExportKubecfgOptions{}

	cmd := &cobra.Command{
		Use:     "kubecfg CLUSTERNAME",
		Short:   export_kubecfg_short,
		Long:    export_kubecfg_long,
		Example: export_kubecfg_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunExportKubecfg(f, out, options, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunExportKubecfg(f *util.Factory, out io.Writer, options *ExportKubecfgOptions, args []string) error {
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

	conf, err := kubeconfig.BuildKubecfg(cluster, keyStore, secretStore, &cloudDiscoveryStatusStore{})
	if err != nil {
		return err
	}

	return conf.WriteKubecfg()
}
