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
	"math/big"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	promoteKeypairLong = templates.LongDesc(i18n.T(`
	Promote a keypair to be the primary, used for signing.
    `))

	promoteKeypairExample = templates.Examples(i18n.T(`
	# Promote the newest ca keypair to be the primary.
	kops promote keypair ca \
		--name k8s-cluster.example.com --state s3://my-state-store

    # Promote a specific service-account keypair to be the primary.
	kops promote keypair service-account 5938372002934847 \
		--name k8s-cluster.example.com --state s3://my-state-store 
	`))

	promoteKeypairShort = i18n.T(`Promote a keypair to be the primary, used for signing.`)
)

type PromoteKeypairOptions struct {
	ClusterName string
	Keyset      string
	KeypairID   string
}

// NewCmdPromoteKeypair returns a promote keypair command.
func NewCmdPromoteKeypair(f *util.Factory, out io.Writer) *cobra.Command {
	options := &PromoteKeypairOptions{}

	cmd := &cobra.Command{
		Use:     "keypair KEYSET [ID]",
		Short:   promoteKeypairShort,
		Long:    promoteKeypairLong,
		Example: promoteKeypairExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)

			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify name of keyset promote keypair in")
			}

			options.Keyset = args[0]

			if len(args) > 2 {
				return fmt.Errorf("can only promote to one keyset at a time")
			}
			if len(args) > 1 {
				options.KeypairID = args[1]
			}

			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completePromoteKeyset(options, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()

			return RunPromoteKeypair(ctx, f, out, options)
		},
	}

	return cmd
}

// RunPromoteKeypair promotes a keypair.
func RunPromoteKeypair(ctx context.Context, f *util.Factory, out io.Writer, options *PromoteKeypairOptions) error {
	if !rotatableKeysets.Has(options.Keyset) {
		return fmt.Errorf("promoting keypairs for %q is not supported", options.Keyset)
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return fmt.Errorf("getting cluster: %q: %v", options.ClusterName, err)
	}

	clientSet, err := f.Clientset()
	if err != nil {
		return fmt.Errorf("getting clientset: %v", err)
	}

	keyStore, err := clientSet.KeyStore(cluster)
	if err != nil {
		return fmt.Errorf("getting keystore: %v", err)
	}

	keyset, err := keyStore.FindKeyset(options.Keyset)
	if err != nil {
		return fmt.Errorf("reading keyset: %v", err)
	} else if keyset == nil {
		return fmt.Errorf("keyset not found")
	}

	keypairID := options.KeypairID
	if keypairID == "" {
		highestTrustedId := big.NewInt(0)
		for id, item := range keyset.Items {
			if item.PrivateKey != nil && item.DistrustTimestamp == nil {
				itemId, ok := big.NewInt(0).SetString(id, 10)
				if ok && highestTrustedId.Cmp(itemId) < 0 {
					highestTrustedId = itemId
				}
			}
		}

		keypairID = highestTrustedId.String()
		if keypairID == keyset.Primary.Id {
			return fmt.Errorf("no keypair newer than current primary %s", keypairID)
		}
	} else if item := keyset.Items[keypairID]; item != nil {
		if item.DistrustTimestamp != nil {
			return fmt.Errorf("keypair is distrusted")
		}
		if item.PrivateKey == nil {
			return fmt.Errorf("keypair has no private key")
		}
	} else {
		return fmt.Errorf("keypair not found")
	}

	keyset.Primary = keyset.Items[keypairID]
	err = keyStore.StoreKeyset(options.Keyset, keyset)
	if err != nil {
		return fmt.Errorf("writing keyset: %v", err)
	}

	fmt.Fprintf(out, "promoted keypair %s", keypairID)
	return nil
}

func completePromoteKeyset(options *PromoteKeypairOptions, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	commandutils.ConfigureKlogForCompletion()
	ctx := context.TODO()

	cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, &rootCommand, nil)
	if cluster == nil {
		return completions, directive
	}

	keyset, _, completions, directive := completeKeyset(cluster, clientSet, args, rotatableKeysetFilter)
	if keyset == nil {
		return completions, directive
	}

	if len(args) == 1 {
		return completeKeypairID(keyset, func(keyset *fi.Keyset, item *fi.KeysetItem) bool {
			return item.DistrustTimestamp == nil && item.Id != keyset.Primary.Id
		})
	}

	if len(args) > 2 {
		return commandutils.CompletionError("too many arguments", nil)
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeKeypairID(keyset *fi.Keyset, filter func(keyset *fi.Keyset, item *fi.KeysetItem) bool) (completions []string, directive cobra.ShellCompDirective) {
	for _, item := range keyset.Items {
		if filter(keyset, item) {
			completions = append(completions, fmt.Sprintf("%s\tissued %s", item.Id, item.Certificate.Certificate.NotBefore.Format("2006-01-02 15:04:05")))
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
