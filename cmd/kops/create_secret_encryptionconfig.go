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
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	createSecretEncryptionconfigLong = templates.LongDesc(i18n.T(`
	Create a new encryption config, and store it in the state store.
	Used to configure encryption-at-rest by the kube-apiserver process
	on each of the master nodes. The config is not updated by this command.`))

	createSecretEncryptionconfigExample = templates.Examples(i18n.T(`
	# Create a new encryption config.
	kops create secret encryptionconfig -f config.yaml \
		--name k8s-cluster.example.com --state s3://example.com
	# Create a new encryption config via stdin.
	generate-encryption-config.sh | kops create secret encryptionconfig -f - \
		--name k8s-cluster.example.com --state s3://example.com
	# Replace an existing encryption config secret.
	kops create secret encryptionconfig -f config.yaml --force \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	createSecretEncryptionconfigShort = i18n.T(`Create an encryption config.`)
)

type CreateSecretEncryptionConfigOptions struct {
	ClusterName          string
	EncryptionConfigPath string
	Force                bool
}

func NewCmdCreateSecretEncryptionConfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretEncryptionConfigOptions{}

	cmd := &cobra.Command{
		Use:     "encryptionconfig",
		Short:   createSecretEncryptionconfigShort,
		Long:    createSecretEncryptionconfigLong,
		Example: createSecretEncryptionconfigExample,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				exitWithError(fmt.Errorf("syntax: -f <EncryptionConfigPath>"))
			}

			err := rootCommand.ProcessArgs(args[0:])
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunCreateSecretEncryptionConfig(f, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.EncryptionConfigPath, "", "f", "", "Path to encryption config yaml file")
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force replace the kops secret if it already exists")

	return cmd
}

func RunCreateSecretEncryptionConfig(f *util.Factory, out io.Writer, options *CreateSecretEncryptionConfigOptions) error {
	if options.EncryptionConfigPath == "" {
		return fmt.Errorf("encryption config path is required (use -f)")
	}

	secret, err := fi.CreateSecret()
	if err != nil {
		return fmt.Errorf("error creating encryption config secret: %v", err)
	}

	cluster, err := GetCluster(f, options.ClusterName)
	if err != nil {
		return err
	}

	clientset, err := f.Clientset()
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
			return fmt.Errorf("error reading encryption config from stdin: %v", err)
		}
	} else {
		data, err = ioutil.ReadFile(options.EncryptionConfigPath)
		if err != nil {
			return fmt.Errorf("error reading encryption config %v: %v", options.EncryptionConfigPath, err)
		}
	}

	var parsedData map[string]interface{}
	err = kops.ParseRawYaml(data, &parsedData)
	if err != nil {
		return fmt.Errorf("Unable to parse yaml %v: %v", options.EncryptionConfigPath, err)
	}

	secret.Data = data

	if !options.Force {
		_, created, err := secretStore.GetOrCreateSecret("encryptionconfig", secret)
		if err != nil {
			return fmt.Errorf("error adding encryptionconfig secret: %v", err)
		}
		if !created {
			return fmt.Errorf("failed to create the encryptionconfig secret as it already exists. The `--force` flag can be passed to replace an existing secret.")
		}
	} else {
		_, err := secretStore.ReplaceSecret("encryptionconfig", secret)
		if err != nil {
			return fmt.Errorf("error updating encryptionconfig secret: %v", err)
		}
	}

	return nil
}
