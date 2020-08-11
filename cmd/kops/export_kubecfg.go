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
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	exportKubecfgLong = templates.LongDesc(i18n.T(`
	Export a kubecfg file for a cluster from the state store. By default the configuration
	will be saved into a users $HOME/.kube/config file. Kops will respect the KUBECONFIG environment variable
	if the --kubeconfig flag is not set.
	`))

	exportKubecfgExample = templates.Examples(i18n.T(`
	# export a kubeconfig file with the cluster admin user (make sure you keep this user safe!)
	kops export kubecfg kubernetes-cluster.example.com --admin

	# export using a user already existing in the kubeconfig file
	kops export kubecfg kubernetes-cluster.example.com --user my-oidc-user

	# export using the internal DNS name, bypassing the cloud load balancer
	kops export kubecfg kubernetes-cluster.example.com --internal
		`))

	exportKubecfgShort = i18n.T(`Export kubecfg.`)
)

type ExportKubecfgOptions struct {
	KubeConfigPath string
	all            bool
	admin          time.Duration
	user           string
	internal       bool
}

func NewCmdExportKubecfg(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ExportKubecfgOptions{}

	cmd := &cobra.Command{
		Use:     "kubecfg CLUSTERNAME",
		Short:   exportKubecfgShort,
		Long:    exportKubecfgLong,
		Example: exportKubecfgExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()
			err := RunExportKubecfg(ctx, f, out, options, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.KubeConfigPath, "kubeconfig", options.KubeConfigPath, "the location of the kubeconfig file to create.")
	cmd.Flags().BoolVar(&options.all, "all", options.all, "export all clusters from the kops state store")
	cmd.Flags().DurationVar(&options.admin, "admin", options.admin, "export a cluster admin user credential with the given lifetime and add it to the cluster context")
	cmd.Flags().Lookup("admin").NoOptDefVal = kubeconfig.DefaultKubecfgAdminLifetime.String()
	cmd.Flags().StringVar(&options.user, "user", options.user, "add an existing user to the cluster context")
	cmd.Flags().BoolVar(&options.internal, "internal", options.internal, "use the cluster's internal DNS name")

	return cmd
}

func RunExportKubecfg(ctx context.Context, f *util.Factory, out io.Writer, options *ExportKubecfgOptions, args []string) error {
	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}
	if options.all {
		if len(args) != 0 {
			return fmt.Errorf("cannot use both --all flag and positional arguments")
		}
	}
	if options.admin != 0 && options.user != "" {
		return fmt.Errorf("cannot use both --admin and --user")
	}

	var clusterList []*kopsapi.Cluster
	if options.all {
		list, err := clientset.ListClusters(ctx, metav1.ListOptions{})
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
		cluster, err := rootCommand.Cluster(ctx)
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

		conf, err := kubeconfig.BuildKubecfg(cluster, keyStore, secretStore, &commands.CloudDiscoveryStatusStore{}, buildPathOptions(options), options.admin, options.user, options.internal)
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
