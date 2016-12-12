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
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"os"
)

type CreateSecretPublickeyOptions struct {
	ClusterName   string
	Name          string
	PublicKeyPath string
}

func NewCmdCreateSecretPublicKey(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretPublickeyOptions{}

	cmd := &cobra.Command{
		Use:   "sshpublickey",
		Short: "Create SSH publickey",
		Long:  `Create SSH publickey.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				exitWithError(fmt.Errorf("syntax: NAME -i <PublicKeyPath>"))
			}
			if len(args) != 1 {
				exitWithError(fmt.Errorf("syntax: NAME -i <PublicKeyPath>"))
			}
			options.Name = args[0]

			err := rootCommand.ProcessArgs(args)
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
