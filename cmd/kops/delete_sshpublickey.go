/*
Copyright 2021 The Kubernetes Authors.

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
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	deleteSSHPublicKeyExample = templates.Examples(i18n.T(`
	# Delete the SSH public key for a cluster
	kops delete sshpublickey k8s-cluster.example.com

	`))

	deleteSSHPublicKeyShort = i18n.T(`Delete an SSH public key.`)
)

type DeleteSSHPublicKeyOptions struct {
	ClusterName string
}

func NewCmdDeleteSSHPublicKey(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteSSHPublicKeyOptions{}

	cmd := &cobra.Command{
		Use:               "sshpublickey [CLUSTER]",
		Short:             deleteSSHPublicKeyShort,
		Example:           deleteSSHPublicKeyExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()

			return RunDeleteSSHPublicKey(ctx, f, out, options)
		},
	}

	return cmd
}

func RunDeleteSSHPublicKey(ctx context.Context, f *util.Factory, out io.Writer, options *DeleteSSHPublicKeyOptions) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
	if err != nil {
		return err
	}

	if err := sshCredentialStore.DeleteSSHCredential(); err != nil {
		return fmt.Errorf("error deleting SSH public key: %v", err)
	}

	return nil
}
