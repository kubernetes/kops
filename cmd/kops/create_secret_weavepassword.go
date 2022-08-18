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
	"k8s.io/kops/pkg/commands/commandutils"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createSecretWeavePasswordLong = templates.LongDesc(i18n.T(`
	Create a new weave encryption secret and store it in the state store.
	Used by Weave networking to encrypt communication between nodes.

	If no password is provided, kOps will generate one at random.

	WARNING: cannot be enabled or changed on a running cluster without downtime.`))

	createSecretWeavePasswordExample = templates.Examples(i18n.T(`
	# Create a new random weave password.
	kops create secret weavepassword \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Install a specific weave password.
	kops create secret weavepassword -f /path/to/weavepassword \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Install a specific weave password via stdin.
	kops create secret weavepassword -f - \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Replace an existing weave password.
	kops create secret weavepassword -f /path/to/weavepassword --force \
		--name k8s-cluster.example.com --state s3://my-state-store
	`))

	createSecretWeavePasswordShort = i18n.T(`Create a Weave password.`)
)

type CreateSecretWeavePasswordOptions struct {
	ClusterName           string
	WeavePasswordFilePath string
	Force                 bool
}

func NewCmdCreateSecretWeavePassword(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretWeavePasswordOptions{}

	cmd := &cobra.Command{
		Use:               "weavepassword [CLUSTER]",
		Short:             createSecretWeavePasswordShort,
		Long:              createSecretWeavePasswordLong,
		Example:           createSecretWeavePasswordExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateSecretWeavePassword(context.TODO(), f, out, options)
		},
	}

	cmd.Flags().StringVarP(&options.WeavePasswordFilePath, "filename", "f", "", "Path to Weave password file")
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force replace the secret if it already exists")

	return cmd
}

func RunCreateSecretWeavePassword(ctx context.Context, f commandutils.Factory, out io.Writer, options *CreateSecretWeavePasswordOptions) error {
	secret, err := fi.CreateSecret()
	if err != nil {
		return fmt.Errorf("creating Weave password: %v", err)
	}

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

	if options.WeavePasswordFilePath != "" {
		var data []byte
		if options.WeavePasswordFilePath == "-" {
			data, err = ConsumeStdin()
			if err != nil {
				return fmt.Errorf("reading Weave password file from stdin: %v", err)
			}
		} else {
			data, err = os.ReadFile(options.WeavePasswordFilePath)
			if err != nil {
				return fmt.Errorf("reading Weave password file %v: %v", options.WeavePasswordFilePath, err)
			}
		}

		secret.Data = data
	}

	if !options.Force {
		_, created, err := secretStore.GetOrCreateSecret("weavepassword", secret)
		if err != nil {
			return fmt.Errorf("adding weavepassword secret: %v", err)
		}
		if !created {
			return fmt.Errorf("failed to create the weavepassword secret as it already exists. Pass the `--force` flag to replace an existing secret")
		}
	} else {
		_, err := secretStore.ReplaceSecret("weavepassword", secret)
		if err != nil {
			return fmt.Errorf("updating weavepassword secret: %v", err)
		}
	}

	return nil
}
