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
	"math/big"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/fi"
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
	name := args[0]
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

	keyset, err := keyStore.FindKeyset(name)
	if err != nil {
		return fmt.Errorf("reading %s keyset: %v", name, err)
	}

	nextVersion := big.NewInt(0)
	var nextItem *fi.KeysetItem
	for id, item := range keyset.Items {
		version, ok := big.NewInt(0).SetString(id, 10)
		if ok && version.Cmp(nextVersion) > 0 {
			nextVersion = version
			nextItem = item
		}
	}
	primaryVersion, ok := big.NewInt(0).SetString(keyset.Primary.Id, 10)
	if !ok || nextItem == nil || nextVersion.Cmp(primaryVersion) <= 0 {
		return fmt.Errorf("no next %s key to rotate in", name)
	}

	items := map[string]*fi.KeysetItem{
		nextItem.Id:       nextItem,
		keyset.Primary.Id: keyset.Primary,
	}
	keyset = &fi.Keyset{
		Items:   items,
		Primary: items[nextItem.Id],
	}

	err = keyStore.StoreKeypair(name, keyset)
	if err != nil {
		return fmt.Errorf("storing %s key: %v", name, err)
	}
	return nil
}
