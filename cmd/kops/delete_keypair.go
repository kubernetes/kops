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
	"time"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	deleteKeypairLong = templates.LongDesc(i18n.T(`
		Delete a keypair.`))

	deleteKeypairExample = templates.Examples(i18n.T(`
	# Syntax: kops delete keypair KEYSET ID
	kops delete keypair ca 5938372002934847

	`))

	deleteKeypairShort = i18n.T(`Delete a keypair.`)
)

type DeleteKeypairOptions struct {
	ClusterName string
	Keyset      string
	KeypairID   string
}

func NewCmdDeleteKeypair(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DeleteKeypairOptions{}

	cmd := &cobra.Command{
		Use:     "keypair KEYSET ID",
		Short:   deleteKeypairShort,
		Long:    deleteKeypairLong,
		Example: deleteKeypairExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			options.ClusterName = rootCommand.ClusterName()
			if options.ClusterName == "" {
				exitWithError(fmt.Errorf("--name is required"))
				return
			}

			if len(args) != 2 {
				exitWithError(fmt.Errorf("usage: kops delete keypair KEYSET ID"))
			}
			options.Keyset = args[0]
			options.KeypairID = args[1]

			err := RunDeleteKeypair(ctx, f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunDeleteKeypair(ctx context.Context, f *util.Factory, out io.Writer, options *DeleteKeypairOptions) error {
	if options.ClusterName == "" {
		return fmt.Errorf("ClusterName is required")
	}
	if options.Keyset == "" {
		return fmt.Errorf("Keyset is required")
	}
	if options.KeypairID == "" {
		return fmt.Errorf("KeypairID is required")
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	keyset, err := keyStore.FindKeyset(options.Keyset)
	if err != nil {
		return err
	}

	if options.KeypairID == keyset.Primary.Id {
		return fmt.Errorf("cannot delete the primary keypair")
	}
	item := keyset.Items[options.KeypairID]
	if item == nil {
		return fmt.Errorf("keypair not found")
	}
	now := time.Now().UTC().Round(0)
	item.DistrustTimestamp = &now

	if err := keyStore.StoreKeyset(options.Keyset, keyset); err != nil {
		return fmt.Errorf("error deleting keypair: %w", err)
	}

	return nil
}
