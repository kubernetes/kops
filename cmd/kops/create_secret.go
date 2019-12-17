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
	"io"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"
)

var (
	createSecretLong = templates.LongDesc(i18n.T(`
	Create a secret`))

	createSecretExample = templates.Examples(i18n.T(`
	# Create a new ssh public key called admin.
	kops create secret sshpublickey admin -i ~/.ssh/id_rsa.pub \
		--name k8s-cluster.example.com --state s3://example.com

	kops create secret dockerconfig -f ~/.docker/config.json \
		--name k8s-cluster.example.com --state s3://example.com

	kops create secret encryptionconfig -f ~/.encryptionconfig.yaml \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	createSecretShort = i18n.T(`Create a secret.`)
)

func NewCmdCreateSecret(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret",
		Short:   createSecretShort,
		Long:    createSecretLong,
		Example: createSecretExample,
	}

	// create subcommands
	cmd.AddCommand(NewCmdCreateSecretPublicKey(f, out))
	cmd.AddCommand(NewCmdCreateSecretDockerConfig(f, out))
	cmd.AddCommand(NewCmdCreateSecretEncryptionConfig(f, out))
	cmd.AddCommand(NewCmdCreateKeypairSecret(f, out))
	cmd.AddCommand(NewCmdCreateSecretWeaveEncryptionConfig(f, out))

	return cmd
}
