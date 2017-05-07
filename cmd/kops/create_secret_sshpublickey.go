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
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	create_secret_sshpublickey_long = templates.LongDesc(i18n.T(`
	Create a new ssh public key, and store the key in the state store.  The
	key is not updated by this command.`))

	create_secret_sshpublickey_example = templates.Examples(i18n.T(`
	# Create an new ssh public key called admin.
	kops create secret sshpublickey admin -i ~/.ssh/id_rsa.pub \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	create_secret_sshpublickey_short = i18n.T(`Create a ssh public key.`)
)

type CreateSecretPublickeyOptions struct {
	ClusterName   string
	Name          string
	PublicKeyPath string
}

func NewCmdCreateSecretPublicKey(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretPublickeyOptions{}

	cmd := &cobra.Command{
		Use:     "sshpublickey",
		Short:   create_secret_sshpublickey_short,
		Long:    create_secret_sshpublickey_long,
		Example: create_secret_sshpublickey_example,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				exitWithError(fmt.Errorf("syntax: NAME -i <PublicKeyPath>"))
			}
			if len(args) != 1 {
				exitWithError(fmt.Errorf("syntax: NAME -i <PublicKeyPath>"))
			}
			options.Name = args[0]

			err := rootCommand.ProcessArgs(args[1:])
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunCreateSecretPublicKey(f, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.PublicKeyPath, "pubkey", "i", "", "Path to SSH public key")

	return cmd
}

func RunCreateSecretPublicKey(f *util.Factory, out io.Writer, options *CreateSecretPublickeyOptions) error {
	if options.PublicKeyPath == "" {
		return fmt.Errorf("public key path is required (use -i)")
	}

	if options.Name == "" {
		return fmt.Errorf("Name is required")
	}

	cluster, err := GetCluster(f, options.ClusterName)
	if err != nil {
		return err
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(options.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("error reading SSH public key %v: %v", options.PublicKeyPath, err)
	}

	err = keyStore.AddSSHPublicKey(options.Name, data)
	if err != nil {
		return fmt.Errorf("error adding SSH public key: %v", err)
	}

	return nil
}
