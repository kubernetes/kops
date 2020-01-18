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

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	deleteSecretLong = templates.LongDesc(i18n.T(`
		Delete a secret.`))

	deleteSecretExample = templates.Examples(i18n.T(`

		`))

	deleteSecretShort = i18n.T(`Delete a secret`)
)

type DeleteSecretOptions struct {
	ClusterName string
	SecretType  string
	SecretName  string
	SecretID    string
}

func NewCmdDeleteSecret(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteSecretOptions{}

	cmd := &cobra.Command{
		Use:     "secret",
		Short:   deleteSecretShort,
		Long:    deleteSecretLong,
		Example: deleteSecretExample,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 && len(args) != 3 {
				exitWithError(fmt.Errorf("Syntax: <type> <name> [<id>]"))
			}

			options.SecretType = args[0]
			options.SecretName = args[1]
			if len(args) == 3 {
				options.SecretID = args[2]
			}

			options.ClusterName = rootCommand.ClusterName()

			err := RunDeleteSecret(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunDeleteSecret(f *util.Factory, out io.Writer, options *DeleteSecretOptions) error {
	if options.ClusterName == "" {
		return fmt.Errorf("ClusterName is required")
	}
	if options.SecretType == "" {
		return fmt.Errorf("SecretType is required")
	}
	if options.SecretName == "" {
		return fmt.Errorf("SecretName is required")
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	cluster, err := GetCluster(f, options.ClusterName)
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return err
	}

	sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
	if err != nil {
		return err
	}

	secrets, err := listSecrets(keyStore, secretStore, sshCredentialStore, options.SecretType, []string{options.SecretName})
	if err != nil {
		return err
	}

	if options.SecretID != "" {
		var matches []*fi.KeystoreItem
		for _, s := range secrets {
			if s.Id == options.SecretID {
				matches = append(matches, s)
			}
		}
		secrets = matches
	}

	if len(secrets) == 0 {
		return fmt.Errorf("secret not found")
	}

	if len(secrets) != 1 {
		// TODO: it would be friendly to print the matching keys
		return fmt.Errorf("found multiple matching secrets; specify the id of the key")
	}

	switch secrets[0].Type {
	case kops.SecretTypeSecret:
		err = secretStore.DeleteSecret(secrets[0].Name)
	case SecretTypeSSHPublicKey:
		sshCredential := &kops.SSHCredential{}
		sshCredential.Name = secrets[0].Name
		if secrets[0].Data != nil {
			sshCredential.Spec.PublicKey = string(secrets[0].Data)
		}
		err = sshCredentialStore.DeleteSSHCredential(sshCredential)
	default:
		keyset := &kops.Keyset{}
		keyset.Name = secrets[0].Name
		keyset.Spec.Type = secrets[0].Type
		err = keyStore.DeleteKeysetItem(keyset, secrets[0].Id)
	}
	if err != nil {
		return fmt.Errorf("error deleting secret: %v", err)
	}

	return nil
}
