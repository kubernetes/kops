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
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getKeypairLong = templates.LongDesc(i18n.T(`
	Display one or many keypairs.`))

	getKeypairExample = templates.Examples(i18n.T(`
	# List the cluster CA keypairs.
	kops get keypairs ca

	# List the service-account keypairs.
	kops get keypairs service-account`))

	getKeypairShort = i18n.T(`Get one or many keypairs.`)
)

type GetKeypairsOptions struct {
	*GetOptions
}

func NewCmdGetKeypairs(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := GetKeypairsOptions{
		GetOptions: getOptions,
	}
	cmd := &cobra.Command{
		Use:     "keypairs",
		Aliases: []string{"keypair"},
		Short:   getKeypairShort,
		Long:    getKeypairLong,
		Example: getKeypairExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()
			err := RunGetKeypairs(ctx, &options, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func listKeypairs(keyStore fi.CAStore, names []string) ([]*fi.KeystoreItem, error) {
	var items []*fi.KeystoreItem

	{
		l, err := keyStore.ListKeysets()
		if err != nil {
			return nil, fmt.Errorf("error listing Keysets: %v", err)
		}

		for _, keyset := range l {
			for _, key := range keyset.Spec.Keys {
				item := &fi.KeystoreItem{
					Name: keyset.Name,
					Type: keyset.Spec.Type,
					ID:   key.Id,
				}
				items = append(items, item)
			}
		}
	}

	if len(names) != 0 {
		var matches []*fi.KeystoreItem
		for _, arg := range names {
			var found []*fi.KeystoreItem
			for _, i := range items {
				if i.Name == arg {
					found = append(found, i)
				}
			}

			matches = append(matches, found...)
		}
		items = matches
	}

	return items, nil
}

func RunGetKeypairs(ctx context.Context, options *GetKeypairsOptions, args []string) error {
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

	items, err := listKeypairs(keyStore, args)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("no keypairs found")
	}
	switch options.output {

	case OutputTable:

		t := &tables.Table{}
		t.AddColumn("NAME", func(i *fi.KeystoreItem) string {
			return i.Name
		})
		t.AddColumn("ID", func(i *fi.KeystoreItem) string {
			return i.ID
		})
		t.AddColumn("TYPE", func(i *fi.KeystoreItem) string {
			return string(i.Type)
		})
		return t.Render(items, os.Stdout, "TYPE", "NAME", "ID")

	case OutputYaml:
		return fmt.Errorf("yaml output format is not (currently) supported for keypairs")
	case OutputJSON:
		return fmt.Errorf("json output format is not (currently) supported for keypairs")

	default:
		return fmt.Errorf("Unknown output format: %q", options.output)
	}
}
