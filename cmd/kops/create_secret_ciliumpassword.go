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
	createSecretCiliumPasswordLong = templates.LongDesc(i18n.T(`
	Create a new Cilium IPsec configuration and store it in the state store.
	This is used by Cilium to encrypt communication between pods/nodes.`))

	createSecretCiliumPasswordExample = templates.Examples(i18n.T(`
	# Create a new Cilium IPsec configuration.
	kops create secret ciliumpassword -f /path/to/configuration.yaml \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Create a new Cilium IPsec key via stdin.
	cat <<EOF | kops create secret ciliumpassword --name k8s-cluster.example.com --state s3://my-state-store -f -
      keys: $(echo "3 rfc4106(gcm(aes)) $(echo $(dd if=/dev/urandom count=20 bs=1 2> /dev/null| xxd -p -c 64)) 128")
    EOF	

	# Replace an existing Cilium IPsec configuration secret
	kops create secret ciliumpassword -f /path/to/configuration.yaml --force \
		--name k8s-cluster.example.com --state s3://my-state-store
	`))

	createSecretCiliumPasswordShort = i18n.T(`Create a Cilium IPsec configuration.`)
)

type CreateSecretCiliumPasswordOptions struct {
	ClusterName            string
	CiliumPasswordFilePath string
	Force                  bool
}

func NewCmdCreateSecretCiliumPassword(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretCiliumPasswordOptions{}

	cmd := &cobra.Command{
		Use:               "ciliumpassword [CLUSTER] -f FILENAME",
		Short:             createSecretCiliumPasswordShort,
		Long:              createSecretCiliumPasswordLong,
		Example:           createSecretCiliumPasswordExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateSecretCiliumEncryptionConfig(context.TODO(), f, out, options)
		},
	}

	cmd.Flags().StringVarP(&options.CiliumPasswordFilePath, "filename", "f", "", "Path to the Cilium IPsec configuration file")
	cmd.MarkFlagRequired("filename")
	cmd.RegisterFlagCompletionFunc("filename", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "json"}, cobra.ShellCompDirectiveFilterFileExt
	})
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force replace the secret if it already exists")

	return cmd
}

func RunCreateSecretCiliumEncryptionConfig(ctx context.Context, f commandutils.Factory, out io.Writer, options *CreateSecretCiliumPasswordOptions) error {
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
	if options.CiliumPasswordFilePath == "-" {
		data, err = ConsumeStdin()
		if err != nil {
			return fmt.Errorf("reading Cilium IPSec config from stdin: %v", err)
		}
	} else {
		data, err = os.ReadFile(options.CiliumPasswordFilePath)
		if err != nil {
			return fmt.Errorf("reading Cilium IPSec config %v: %v", options.CiliumPasswordFilePath, err)
		}
	}

	var parsedData map[string]interface{}
	err = kops.ParseRawYaml(data, &parsedData)
	if err != nil {
		return fmt.Errorf("unable to parse YAML %v: %v", options.CiliumPasswordFilePath, err)
	}

	secret := &fi.Secret{
		Data: data,
	}

	if !options.Force {
		_, created, err := secretStore.GetOrCreateSecret("ciliumpassword", secret)
		if err != nil {
			return fmt.Errorf("error adding Cilium IPSec secret: %v", err)
		}
		if !created {
			return fmt.Errorf("failed to create the Cilium IPSec secret as it already exists. Pass the `--force` flag to replace an existing secret")
		}
	} else {
		_, err := secretStore.ReplaceSecret("ciliumpassword", secret)
		if err != nil {
			return fmt.Errorf("updating Cilium IPSec secret: %v", err)
		}
	}

	return nil
}
