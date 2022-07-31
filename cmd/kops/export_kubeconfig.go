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
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	exportKubeconfigLong = templates.LongDesc(i18n.T(`
	Export a kubeconfig file for a cluster from the state store. By default the configuration
	will be saved into a users $HOME/.kube/config file.
	`))

	exportKubeconfigExample = templates.Examples(i18n.T(`
	# export a kubeconfig file with the cluster admin user (make sure you keep this user safe!)
	kops export kubeconfig k8s-cluster.example.com --admin

	# export using a user already existing in the kubeconfig file
	kops export kubeconfig k8s-cluster.example.com --user my-oidc-user

	# export using the internal DNS name, bypassing the cloud load balancer
	kops export kubeconfig k8s-cluster.example.com --internal
	`))

	exportKubeconfigShort = i18n.T(`Export kubeconfig.`)
)

type ExportKubeconfigOptions struct {
	ClusterName    string
	KubeConfigPath string
	all            bool
	admin          time.Duration
	user           string
	internal       bool

	// UseKopsAuthenticationPlugin controls whether we should use the kOps auth helper instead of a static credential
	UseKopsAuthenticationPlugin bool
}

func NewCmdExportKubeconfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ExportKubeconfigOptions{}

	cmd := &cobra.Command{
		Use:     "kubeconfig [CLUSTER | --all]",
		Aliases: []string{"kubecfg"},
		Short:   exportKubeconfigShort,
		Long:    exportKubeconfigLong,
		Example: exportKubeconfigExample,
		Args: func(cmd *cobra.Command, args []string) error {
			if options.admin != 0 && options.user != "" {
				return fmt.Errorf("cannot use both --admin and --user")
			}
			if options.all {
				if len(args) != 0 {
					return fmt.Errorf("cannot use both --all flag and positional arguments")
				}
				return nil
			} else {
				return rootCommand.clusterNameArgs(&options.ClusterName)(cmd, args)
			}
		},
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunExportKubeconfig(context.TODO(), f, out, options, args)
		},
	}

	cmd.Flags().StringVar(&options.KubeConfigPath, "kubeconfig", options.KubeConfigPath, "Filename of the kubeconfig to create")
	cmd.Flags().BoolVar(&options.all, "all", options.all, "Export all clusters from the kOps state store")
	cmd.Flags().DurationVar(&options.admin, "admin", options.admin, "Also export a cluster admin user credential with the specified lifetime and add it to the cluster context")
	cmd.Flags().Lookup("admin").NoOptDefVal = kubeconfig.DefaultKubecfgAdminLifetime.String()
	cmd.Flags().StringVar(&options.user, "user", options.user, "Existing user in kubeconfig file to use")
	cmd.RegisterFlagCompletionFunc("user", completeKubecfgUser)
	cmd.Flags().BoolVar(&options.internal, "internal", options.internal, "Use the cluster's internal DNS name")
	cmd.Flags().BoolVar(&options.UseKopsAuthenticationPlugin, "auth-plugin", options.UseKopsAuthenticationPlugin, "Use the kOps authentication plugin")

	return cmd
}

func RunExportKubeconfig(ctx context.Context, f *util.Factory, out io.Writer, options *ExportKubeconfigOptions, args []string) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
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
		cluster, err := GetCluster(ctx, f, options.ClusterName)
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

		cloud, err := cloudup.BuildCloud(cluster)
		if err != nil {
			return err
		}
		conf, err := kubeconfig.BuildKubecfg(
			cluster,
			keyStore,
			secretStore,
			cloud,
			options.admin,
			options.user,
			options.internal,
			f.KopsStateStore(),
			options.UseKopsAuthenticationPlugin)
		if err != nil {
			return err
		}

		if err := conf.WriteKubecfg(buildPathOptions(options)); err != nil {
			return err
		}
	}

	return nil
}

func buildPathOptions(options *ExportKubeconfigOptions) *clientcmd.PathOptions {
	pathOptions := clientcmd.NewDefaultPathOptions()

	if len(options.KubeConfigPath) > 0 {
		pathOptions.GlobalFile = options.KubeConfigPath
		pathOptions.EnvVar = ""
		pathOptions.GlobalFileSubpath = ""
	}

	return pathOptions
}

func completeKubecfgUser(cmd *cobra.Command, args []string, complete string) ([]string, cobra.ShellCompDirective) {
	pathOptions := clientcmd.NewDefaultPathOptions()

	config, err := pathOptions.GetStartingConfig()
	if err != nil {
		return commandutils.CompletionError("reading kubeconfig", err)
	}

	var users []string
	for user := range config.AuthInfos {
		users = append(users, user)
	}

	return users, cobra.ShellCompDirectiveNoFileComp
}
