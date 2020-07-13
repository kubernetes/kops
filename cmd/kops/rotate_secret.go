/*
Copyright 2020 The Kubernetes Authors.

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

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	rotateSecretLong = templates.LongDesc(i18n.T(`
	    Rotate a secret.
	`))

	rotateSecretExample = templates.Examples(i18n.T(`
	# Rotate the service-account key
	kops rotate secret service-account
	`))
	rotateSecretShort = i18n.T(`Rotate a secret.`)
)

type RotateSecretCmd struct {
}

var rotateSecretCommand RotateSecretCmd

func init() {
	cmd := &cobra.Command{
		Use:     "secret",
		Aliases: []string{"secrets"},
		Short:   rotateSecretShort,
		Long:    rotateSecretLong,
		Example: rotateSecretExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()
			err := rotateSecretCommand.Run(ctx, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	rotateCmd.cobraCommand.AddCommand(cmd)
}

func (c *RotateSecretCmd) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("specify name of secret to rotate")
	}
	if len(args) != 1 {
		return fmt.Errorf("can only rotate one secret at a time")
	}
	if args[0] != "service-account" {
		return fmt.Errorf("can only rotate the service-account secret")
	}

	cluster, err := rootCommand.Cluster(ctx)
	if err != nil {
		return err
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	nextCert, nextKey, _, err := keyStore.FindKeypair("service-account-next")
	if err != nil {
		return fmt.Errorf("reading service-account-next key: %v", err)
	}
	if nextKey == nil {
		return fmt.Errorf("no service-account-next key to rotate in")
	}
	privateKeyset, err := keyStore.FindPrivateKeyset("service-account-next")
	if err != nil {
		return fmt.Errorf("reading service-account-next keyset: %v", err)
	}

	currentCert, currentKey, _, err := keyStore.FindKeypair("service-account")
	if err != nil {
		return fmt.Errorf("reading service-account key: %v", err)
	}

	if currentKey != nil {
		err = keyStore.StoreKeypair("service-account-previous", currentCert, currentKey)
		if err != nil {
			return fmt.Errorf("storing service-account-previous key: %v", err)
		}
	}

	err = keyStore.StoreKeypair("service-account", nextCert, nextKey)
	if err != nil {
		return fmt.Errorf("storing service-account key: %v", err)
	}

	for _, key := range privateKeyset.Spec.Keys {
		err = keyStore.DeleteKeysetItem(privateKeyset, key.Id)
		if err != nil {
			return fmt.Errorf("deleting service-account-previous key: %v", err)
		}
	}
	return nil
}
