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

	"github.com/spf13/cobra"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	createSecretWeaveEncryptionconfigLong = templates.LongDesc(i18n.T(`
	Create a new weave encryption secret, and store it in the state store.
	Used to weave networking to use encrypted communication between nodes.

	If no password is provided, kops will generate one at random.

	WARNING: cannot be enabled on a running cluster without downtime.`))

	createSecretWeaveEncryptionconfigExample = templates.Examples(i18n.T(`
	# Create a new random weave password.
	kops create secret weavepassword \
		--name k8s-cluster.example.com --state s3://example.com
	# Install a specific weave password.
	kops create secret weavepassword -f /path/to/weavepassword \
		--name k8s-cluster.example.com --state s3://example.com
	# Install a specific weave password via stdin.
	kops create secret weavepassword -f - \
		--name k8s-cluster.example.com --state s3://example.com
	# Replace an existing weavepassword secret.
	kops create secret weavepassword -f /path/to/weavepassword --force \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	createSecretWeaveEncryptionconfigShort = i18n.T(`Create a weave encryption config.`)
)

type CreateSecretWeaveEncryptionConfigOptions struct {
	ClusterName           string
	WeavePasswordFilePath string
	Force                 bool
}

func NewCmdCreateSecretWeaveEncryptionConfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretWeaveEncryptionConfigOptions{}

	cmd := &cobra.Command{
		Use:     "weavepassword",
		Short:   createSecretWeaveEncryptionconfigShort,
		Long:    createSecretWeaveEncryptionconfigLong,
		Example: createSecretWeaveEncryptionconfigExample,
		Run: func(cmd *cobra.Command, args []string) {

			err := rootCommand.ProcessArgs(args[0:])
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunCreateSecretWeaveEncryptionConfig(f, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.WeavePasswordFilePath, "", "f", "", "Path to the weave password file (optional)")
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force replace the kops secret if it already exists")

	return cmd
}

func RunCreateSecretWeaveEncryptionConfig(f *util.Factory, options *CreateSecretWeaveEncryptionConfigOptions) error {

	secret, err := fi.CreateSecret()
	if err != nil {
		return fmt.Errorf("error creating encryption secret: %v", err)
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

	if options.WeavePasswordFilePath != "" {
		var data []byte
		if options.WeavePasswordFilePath == "-" {
			data, err = ConsumeStdin()
			if err != nil {
				return fmt.Errorf("error reading weave password file from stdin: %v", err)
			}
		} else {
			data, err = ioutil.ReadFile(options.WeavePasswordFilePath)
			if err != nil {
				return fmt.Errorf("error reading weave password file %v: %v", options.WeavePasswordFilePath, err)
			}

		}

		secret.Data = data
	}

	if !options.Force {
		_, created, err := secretStore.GetOrCreateSecret("weavepassword", secret)
		if err != nil {
			return fmt.Errorf("error adding weavepassword secret: %v", err)
		}
		if !created {
			return fmt.Errorf("failed to create the weavepassword secret as it already exists. The `--force` flag can be passed to replace an existing secret")
		}
	} else {
		_, err := secretStore.ReplaceSecret("weavepassword", secret)
		if err != nil {
			return fmt.Errorf("error updating weavepassword secret: %v", err)
		}
	}

	return nil
}
