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
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
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
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			options.ClusterName = rootCommand.ClusterName()

			if options.ClusterName == "" {
				exitWithError(fmt.Errorf("--name is required"))
				return
			}

			if len(args) == 0 {
				exitWithError(fmt.Errorf("must specify name of keyset promote keypair in"))
			}
			if len(args) > 2 {
				exitWithError(fmt.Errorf("can only promote to one keyset at a time"))
			}
			options.Keyset = args[0]
			if len(args) > 1 {
				options.KeypairID = args[1]
			}

			err := RunPromoteKeypair(ctx, f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

// RunPromoteKeypair promotes a keypair.
func RunPromoteKeypair(ctx context.Context, f *util.Factory, out io.Writer, options *PromoteKeypairOptions) error {
	if keysetCommonNames[options.Keyset] == "" {
		return fmt.Errorf("promoting keypairs for %q is not supported", options.Keyset)
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return fmt.Errorf("error getting cluster: %q: %v", options.ClusterName, err)
	}

	clientSet, err := f.Clientset()
	if err != nil {
		return fmt.Errorf("error getting clientset: %v", err)
	}

	keyStore, err := clientSet.KeyStore(cluster)
	if err != nil {
		return fmt.Errorf("error getting keystore: %v", err)
	}

	keyset, err := keyStore.FindKeyset(options.Keyset)
	if err != nil {
		return fmt.Errorf("reading keyset: %v", err)
	} else if keyset == nil {
		return fmt.Errorf("keyset not found")
	}

	keypairID := options.KeypairID
	if keypairID == "" {
		highestId := big.NewInt(0)
		for id, item := range keyset.Items {
			if item.PrivateKey != nil {
				itemId, ok := big.NewInt(0).SetString(id, 10)
				if ok && highestId.Cmp(itemId) < 0 {
					highestId = itemId
				}
			}
		}

		keypairID = highestId.String()
		if keypairID == keyset.Primary.Id {
			return fmt.Errorf("no keypair newer than current primary %s", keypairID)
		}
	} else if item := keyset.Items[keypairID]; item != nil {
		if item.PrivateKey == nil {
			return fmt.Errorf("keypair has no private key")
		}
	} else {
		return fmt.Errorf("keypair not found")
	}

	keyset.Primary = keyset.Items[keypairID]
	err = keyStore.StoreKeyset(options.Keyset, keyset)
	if err != nil {
		return fmt.Errorf("error writing keyset: %v", err)
	}

	klog.Infof("promoted keypair %s", keypairID)
	return nil
}
