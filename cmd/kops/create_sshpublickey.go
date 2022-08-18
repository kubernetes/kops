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
	"github.com/spf13/pflag"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createSSHPublicKeyLong = templates.LongDesc(i18n.T(`
	Create a new SSH public key, and store the key in the state store.  The
	key is not updated by this command.`))

	createSSHPublicKeyExample = templates.Examples(i18n.T(`
	# Create a new SSH public key from the file ""~/.ssh/id_rsa.pub".
	kops create sshpublickey k8s-cluster.example.com -i ~/.ssh/id_rsa.pub
	`))

	createSSHPublicKeyShort = i18n.T(`Create an SSH public key.`)
)

type CreateSSHPublicKeyOptions struct {
	ClusterName   string
	PublicKeyPath string
}

func NewCmdCreateSSHPublicKey(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSSHPublicKeyOptions{}

	cmd := &cobra.Command{
		Use:               "sshpublickey [CLUSTER]",
		Short:             createSSHPublicKeyShort,
		Long:              createSSHPublicKeyLong,
		Example:           createSSHPublicKeyExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateSSHPublicKey(context.TODO(), f, out, options)
		},
	}

	cmd.Flags().StringVarP(&options.PublicKeyPath, "ssh-public-key", "i", "", "Path to SSH public key")
	cmd.MarkFlagRequired("ssh-public-key")
	cmd.RegisterFlagCompletionFunc("ssh-public-key", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"pub"}, cobra.ShellCompDirectiveFilterFileExt
	})

	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "pubkey":
			name = "ssh-public-key"
		}
		return pflag.NormalizedName(name)
	})

	return cmd
}

func RunCreateSSHPublicKey(ctx context.Context, f *util.Factory, out io.Writer, options *CreateSSHPublicKeyOptions) error {
	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(options.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("error reading SSH public key %v: %v", options.PublicKeyPath, err)
	}

	err = sshCredentialStore.AddSSHPublicKey(data)
	if err != nil {
		return fmt.Errorf("error adding SSH public key: %v", err)
	}

	return nil
}
