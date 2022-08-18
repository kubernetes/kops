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
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createSecretEncryptionConfigLong = templates.LongDesc(i18n.T(`
	Create a new encryption config and store it in the state store.
	Used to configure encryption-at-rest by the kube-apiserver process.`))

	createSecretEncryptionConfigExample = templates.Examples(i18n.T(`
	# Create a new encryption config.
	kops create secret encryptionconfig -f config.yaml \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Create a new encryption config via stdin.
	generate-encryption-config.sh | kops create secret encryptionconfig -f - \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Replace an existing encryption config secret.
	kops create secret encryptionconfig -f config.yaml --force \
		--name k8s-cluster.example.com --state s3://my-state-store
	`))

	createSecretEncryptionConfigShort = i18n.T(`Create an encryption config.`)
)

type CreateSecretEncryptionConfigOptions struct {
	ClusterName          string
	EncryptionConfigPath string
	Force                bool
}

func NewCmdCreateSecretEncryptionConfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretEncryptionConfigOptions{}

	cmd := &cobra.Command{
		Use:               "encryptionconfig [CLUSTER] -f FILENAME",
		Short:             createSecretEncryptionConfigShort,
		Long:              createSecretEncryptionConfigLong,
		Example:           createSecretEncryptionConfigExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateSecretEncryptionConfig(context.TODO(), f, out, options)
		},
	}

	cmd.Flags().StringVarP(&options.EncryptionConfigPath, "filename", "f", "", "Path to encryption config YAML file")
	cmd.MarkFlagRequired("filename")
	cmd.RegisterFlagCompletionFunc("filename", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "json"}, cobra.ShellCompDirectiveFilterFileExt
	})
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force replace the secret if it already exists")

	return cmd
}

func RunCreateSecretEncryptionConfig(ctx context.Context, f commandutils.Factory, out io.Writer, options *CreateSecretEncryptionConfigOptions) error {
	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return err
	}
	var data []byte
	if options.EncryptionConfigPath == "-" {
		data, err = ConsumeStdin()
		if err != nil {
			return fmt.Errorf("reading encryption config from stdin: %v", err)
		}
	} else {
		data, err = os.ReadFile(options.EncryptionConfigPath)
		if err != nil {
			return fmt.Errorf("reading encryption config %v: %v", options.EncryptionConfigPath, err)
		}
	}

	var parsedData map[string]interface{}
	err = kops.ParseRawYaml(data, &parsedData)
	if err != nil {
		return fmt.Errorf("unable to parse YAML %v: %v", options.EncryptionConfigPath, err)
	}

	secret := &fi.Secret{
		Data: data,
	}

	if !options.Force {
		_, created, err := secretStore.GetOrCreateSecret("encryptionconfig", secret)
		if err != nil {
			return fmt.Errorf("adding encryptionconfig secret: %v", err)
		}
		if !created {
			return fmt.Errorf("failed to create the encryptionconfig secret as it already exists. Pass the `--force` flag to replace an existing secret")
		}
	} else {
		_, err := secretStore.ReplaceSecret("encryptionconfig", secret)
		if err != nil {
			return fmt.Errorf("updating encryptionconfig secret: %v", err)
		}
	}

	return nil
}
