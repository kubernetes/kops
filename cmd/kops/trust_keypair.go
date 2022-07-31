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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/pretty"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	trustKeypairLong = pretty.LongDesc(i18n.T(`
	Trust one or more keypairs in a keyset.

	Trusting adds the certificates of the specified keypairs to trust
	stores. It is the reverse of the ` + pretty.Bash("kops distrust keypair") + ` command.
	`))

	trustKeypairExample = templates.Examples(i18n.T(`
	kops trust keypair ca 6977545226837259959403993899

	`))

	trustKeypairShort = i18n.T(`Trust a keypair.`)
)

type TrustKeypairOptions struct {
	ClusterName string
	Keyset      string
	KeypairIDs  []string
}

func NewCmdTrustKeypair(f *util.Factory, out io.Writer) *cobra.Command {
	options := &TrustKeypairOptions{}

	cmd := &cobra.Command{
		Use:     "keypair KEYSET ID...",
		Short:   trustKeypairShort,
		Long:    trustKeypairLong,
		Example: trustKeypairExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)
			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify name of keyset to trust keypair in")
			}
			options.Keyset = args[0]

			if len(args) == 1 {
				return fmt.Errorf("must specify names of keypairs to trust keypair in")
			}
			options.KeypairIDs = args[1:]

			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeTrustKeyset(f, options, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()

			return RunTrustKeypair(ctx, f, out, options)
		},
	}

	return cmd
}

func RunTrustKeypair(ctx context.Context, f *util.Factory, out io.Writer, options *TrustKeypairOptions) error {
	clientset, err := f.KopsClient()
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

	for _, id := range options.KeypairIDs {
		item := keyset.Items[id]
		if item == nil {
			return fmt.Errorf("keypair not found")
		}

		if item.DistrustTimestamp == nil {
			continue
		}

		item.DistrustTimestamp = nil

		if err := keyStore.StoreKeyset(options.Keyset, keyset); err != nil {
			return fmt.Errorf("error storing keypair: %w", err)
		}

		fmt.Fprintf(out, "Trusted %s %s\n", options.Keyset, id)
	}

	return nil
}

func completeTrustKeyset(f commandutils.Factory, options *TrustKeypairOptions, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	commandutils.ConfigureKlogForCompletion()
	ctx := context.TODO()

	cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, f, nil)
	if cluster == nil {
		return completions, directive
	}

	keyset, _, completions, directive := completeKeyset(cluster, clientSet, args, func(name string, keyset *fi.Keyset) bool {
		if name == "all" {
			return false
		}
		return rotatableKeysetFilter(name, keyset)
	})
	if keyset == nil {
		return completions, directive
	}

	alreadySelected := sets.NewString(args[1:]...)
	return completeKeypairID(keyset, func(keyset *fi.Keyset, item *fi.KeysetItem) bool {
		return item.DistrustTimestamp != nil && !alreadySelected.Has(item.Id)
	})
}
