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
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createSecretCiliumEncryptionconfigLong = templates.LongDesc(i18n.T(`
	Create a new cilium encryption secret, and store it in the state store.
	Used by Cilium to generate encrypted communication between pods/nodes.`))

	createSecretCiliumEncryptionconfigExample = templates.Examples(i18n.T(`
	# Create a new cilium encryption key.
	kops create secret ciliumpassword -f /path/to/ciliumpassword \
		--name k8s-cluster.example.com --state s3://example.com
	# Create a new cilium encryption key via stdin.
	cat <<EOF | kops create secret ciliumpassword --name k8s-cluster.example.com --state s3://example.com -f -
      keys: $(echo "3 rfc4106(gcm(aes)) $(echo $(dd if=/dev/urandom count=20 bs=1 2> /dev/null| xxd -p -c 64)) 128")
    EOF	
	# Replace an existing ciliumpassword secret
	kops create secret ciliumpassword -f /path/to/ciliumpassword --force \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	createSecretCiliumEncryptionconfigShort = i18n.T(`Create a cilium encryption key.`)
)

type CreateSecretCiliumEncryptionConfigOptions struct {
	ClusterName            string
	CiliumPasswordFilePath string
	Force                  bool
}

func NewCmdCreateSecretCiliumEncryptionConfig(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretCiliumEncryptionConfigOptions{}

	cmd := &cobra.Command{
		Use:     "ciliumpassword",
		Short:   createSecretCiliumEncryptionconfigShort,
		Long:    createSecretCiliumEncryptionconfigLong,
		Example: createSecretCiliumEncryptionconfigExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			if len(args) != 0 {
				exitWithError(fmt.Errorf("syntax: -f <CiliumPasswordFilePath>"))
			}

			err := rootCommand.ProcessArgs(args[0:])
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunCreateSecretCiliumEncryptionConfig(ctx, f, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.CiliumPasswordFilePath, "", "f", "", "Path to the cilium encryption config file")
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force replace the kops secret if it already exists")

	return cmd
}

func RunCreateSecretCiliumEncryptionConfig(ctx context.Context, f *util.Factory, out io.Writer, options *CreateSecretCiliumEncryptionConfigOptions) error {
	if options.CiliumPasswordFilePath == "" {
		return fmt.Errorf("cilium encryption config path is required (use -f)")
	}

	secret, err := fi.CreateSecret()
	if err != nil {
		return fmt.Errorf("error creating encryption secret: %v", err)
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
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
	if options.CiliumPasswordFilePath == "-" {
		data, err = ConsumeStdin()
		if err != nil {
			return fmt.Errorf("error reading cilium encryption config from stdin: %v", err)
		}
	} else {
		data, err = ioutil.ReadFile(options.CiliumPasswordFilePath)
		if err != nil {
			return fmt.Errorf("error reading encryption config %v: %v", options.CiliumPasswordFilePath, err)
		}
	}

	var parsedData map[string]interface{}
	err = kops.ParseRawYaml(data, &parsedData)
	if err != nil {
		return fmt.Errorf("unable to parse yaml %v: %v", options.CiliumPasswordFilePath, err)
	}

	secret.Data = data

	if !options.Force {
		_, created, err := secretStore.GetOrCreateSecret("ciliumpassword", secret)
		if err != nil {
			return fmt.Errorf("error adding ciliumpassword secret: %v", err)
		}
		if !created {
			return fmt.Errorf("failed to create the encryptionconfig secret as it already exists. The `--force` flag can be passed to replace an existing secret")
		}
	} else {
		_, err := secretStore.ReplaceSecret("ciliumpassword", secret)
		if err != nil {
			return fmt.Errorf("error updating ciliumpassword secret: %v", err)
		}
	}

	return nil
}
