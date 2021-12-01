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
	"k8s.io/kubectl/pkg/util/i18n"
)

var createSecretShort = i18n.T(`Create a secret.`)

func NewCmdCreateSecret(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: createSecretShort,
	}

	// create subcommands
	cmd.AddCommand(NewCmdCreateSecretCiliumPassword(f, out))
	cmd.AddCommand(NewCmdCreateSecretDockerConfig(f, out))
	cmd.AddCommand(NewCmdCreateSecretEncryptionConfig(f, out))
	cmd.AddCommand(NewCmdCreateSecretWeavePassword(f, out))

	sshPublicKey := NewCmdCreateSSHPublicKey(f, out)
	sshPublicKey.Hidden = true
	innerArgs := sshPublicKey.Args
	sshPublicKey.Args = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && args[0] == "admin" {
			// Backwards compatibility
			args = args[1:]
		}

		return innerArgs(cmd, args)
	}
	cmd.AddCommand(sshPublicKey)

	return cmd
}
