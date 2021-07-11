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
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getKeypairExample = templates.Examples(i18n.T(`
	# List the cluster CA keypairs.
	kops get keypairs ca

	# List the service-account keypairs, including distrusted ones.
	kops get keypairs service-account --distrusted`))

	getKeypairShort = i18n.T(`Get one or many keypairs.`)
)

type GetKeypairsOptions struct {
	*GetOptions
	Distrusted bool
}

func NewCmdGetKeypairs(f *util.Factory, out io.Writer, getOptions *GetOptions) *cobra.Command {
	options := &GetKeypairsOptions{
		GetOptions: getOptions,
	}
	cmd := &cobra.Command{
		Use:     "keypairs [KEYSET]...",
		Aliases: []string{"keypair"},
		Short:   getKeypairShort,
		Example: getKeypairExample,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeGetKeypairs(options, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return RunGetKeypairs(ctx, out, options, args)
		},
	}

	cmd.Flags().BoolVar(&options.Distrusted, "distrusted", options.Distrusted, "Include distrusted keypairs")

	return cmd
}

type keypairItem struct {
	Name              string
	Id                string
	DistrustTimestamp *time.Time
	IsPrimary         bool
	Certificate       *pki.Certificate
	HasPrivateKey     bool
}

func listKeypairs(keyStore fi.CAStore, names []string, includeDistrusted bool) ([]*keypairItem, error) {
	var items []*keypairItem

	l, err := keyStore.ListKeysets()
	if err != nil {
		return nil, fmt.Errorf("error listing Keysets: %v", err)
	}

	for name, keyset := range l {
		if len(names) != 0 {
			found := false
			for _, n := range names {
				if n == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		for _, item := range keyset.Items {
			if includeDistrusted || item.DistrustTimestamp == nil {
				items = append(items, &keypairItem{
					Name:              name,
					Id:                item.Id,
					DistrustTimestamp: item.DistrustTimestamp,
					IsPrimary:         item.Id == keyset.Primary.Id,
					Certificate:       item.Certificate,
					HasPrivateKey:     item.PrivateKey != nil,
				})
			}
		}
	}

	return items, nil
}

func RunGetKeypairs(ctx context.Context, out io.Writer, options *GetKeypairsOptions, args []string) error {
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

	items, err := listKeypairs(keyStore, args, options.Distrusted)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("no keypairs found")
	}
	switch options.output {

	case OutputTable:
		t := &tables.Table{}
		t.AddColumn("NAME", func(i *keypairItem) string {
			return i.Name
		})
		t.AddColumn("ID", func(i *keypairItem) string {
			return i.Id
		})
		t.AddColumn("DISTRUSTED", func(i *keypairItem) string {
			if i.DistrustTimestamp != nil {
				return i.DistrustTimestamp.Local().Format("2006-01-02")
			}
			return ""
		})
		t.AddColumn("ISSUED", func(i *keypairItem) string {
			return i.Certificate.Certificate.NotBefore.Local().Format("2006-01-02")
		})
		t.AddColumn("EXPIRES", func(i *keypairItem) string {
			return i.Certificate.Certificate.NotAfter.Local().Format("2006-01-02")
		})
		t.AddColumn("PRIMARY", func(i *keypairItem) string {
			if i.IsPrimary {
				return "*"
			}
			return ""
		})
		t.AddColumn("HASPRIVATE", func(i *keypairItem) string {
			if i.HasPrivateKey {
				return "*"
			}
			return ""
		})
		columnNames := []string{"NAME", "ID", "ISSUED", "EXPIRES"}
		if options.Distrusted {
			columnNames = append(columnNames, "DISTRUSTED")
		}
		columnNames = append(columnNames, "PRIMARY", "HASPRIVATE")
		return t.Render(items, out, columnNames...)

	case OutputYaml:
		return fmt.Errorf("yaml output format is not (currently) supported for keypairs")
	case OutputJSON:
		return fmt.Errorf("json output format is not (currently) supported for keypairs")

	default:
		return fmt.Errorf("Unknown output format: %q", options.output)
	}
}

func completeGetKeypairs(options *GetKeypairsOptions, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	commandutils.ConfigureKlogForCompletion()
	ctx := context.TODO()

	cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, &rootCommand, nil)
	if cluster == nil {
		return completions, directive
	}

	alreadySelected := sets.NewString(args...).Insert("all")
	_, _, completions, directive = completeKeyset(cluster, clientSet, nil, func(name string, keyset *fi.Keyset) bool {
		return !alreadySelected.Has(name)
	})

	return completions, directive
}
