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

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
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
		Use:     "keypair",
		Short:   deleteKeypairShort,
		Long:    deleteKeypairLong,
		Example: deleteKeypairExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			if len(args) != 2 && len(args) != 3 {
				exitWithError(fmt.Errorf("Syntax: <keyset> <id>"))
			}

			options.Keyset = args[0]
			options.KeypairID = args[1]

			options.ClusterName = rootCommand.ClusterName()

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

	keypairs, err := listKeypairs(keyStore, []string{options.Keyset})
	if err != nil {
		return err
	}

	{
		var matches []*fi.KeystoreItem
		for _, s := range keypairs {
			if s.ID == options.KeypairID {
				matches = append(matches, s)
			}
		}
		keypairs = matches
	}

	if len(keypairs) == 0 {
		return fmt.Errorf("keypair not found")
	}

	if len(keypairs) != 1 {
		// TODO: it would be friendly to print the matching keys
		return fmt.Errorf("found multiple matching keypairs; specify the id of the key")
	}

	keyset := &kops.Keyset{}
	keyset.Name = keypairs[0].Name
	keyset.Spec.Type = keypairs[0].Type
	err = keyStore.DeleteKeysetItem(keyset, keypairs[0].ID)
	if err != nil {
		return fmt.Errorf("error deleting keypair: %v", err)
	}

	return nil
}
