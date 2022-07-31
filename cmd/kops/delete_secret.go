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

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	deleteSecretExample = templates.Examples(i18n.T(`
	# Delete the encryptionconfig secret
	kops delete secret encryptionconfig

	`))

	deleteSecretShort = i18n.T(`Delete one or more secrets.`)
)

type DeleteSecretOptions struct {
	ClusterName string
	SecretType  string
	SecretNames []string
}

func NewCmdDeleteSecret(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteSecretOptions{}

	cmd := &cobra.Command{
		Use:     "secret SECRET_NAME...",
		Short:   deleteSecretShort,
		Example: deleteSecretExample,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "sshpublickey" {
				options.SecretType = args[0]
				return nil
			}

			if len(args) > 0 && args[0] == "secret" {
				args = args[1:]
			}

			if len(args) == 0 {
				return fmt.Errorf("secret name is required")
			}
			if len(args) > 1 {
				return fmt.Errorf("too many arguments")
			}
			options.SecretNames = args
			options.ClusterName = rootCommand.ClusterName(true)

			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			return nil
		},
		ValidArgsFunction: completeSecretNames(f),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDeleteSecret(context.TODO(), f, out, options)
		},
	}

	return cmd
}

func RunDeleteSecret(ctx context.Context, f *util.Factory, out io.Writer, options *DeleteSecretOptions) error {
	if options.SecretType == "sshpublickey" {
		return RunDeleteSSHPublicKey(ctx, f, out, &DeleteSSHPublicKeyOptions{
			ClusterName: options.ClusterName,
		})
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return err
	}

	for _, name := range options.SecretNames {
		err = secretStore.DeleteSecret(name)
		if err != nil {
			return fmt.Errorf("deleting secret %q: %v", name, err)
		}
	}

	return nil
}

func completeSecretNames(f commandutils.Factory) func(cmd *cobra.Command, args []string, complete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, complete string) ([]string, cobra.ShellCompDirective) {
		commandutils.ConfigureKlogForCompletion()
		ctx := context.TODO()

		cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, f, nil)
		if cluster == nil {
			return completions, directive
		}

		secretStore, err := clientSet.SecretStore(cluster)
		if err != nil {
			return commandutils.CompletionError("constructing secret store", err)
		}

		alreadySelected := sets.NewString(args...)
		var secrets []string
		items, err := listSecrets(secretStore, nil)
		if err != nil {
			return commandutils.CompletionError("listing secrets", err)
		}
		for _, secret := range items {
			if !alreadySelected.Has(secret) {
				secrets = append(secrets, secret)
			}
		}

		return secrets, cobra.ShellCompDirectiveNoFileComp
	}
}
