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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	distrustKeypairLong = templates.LongDesc(i18n.T(`
	Distrust one or more keypairs in a keyset.

	Distrusting removes the certificates of the specified keypairs from trust
	stores.

	Only secondary keypairs may be distrusted.

	If no keypair IDs are specified, all keypairs in the keyset that
	are older than the primary keypair will be distrusted.

	If the keyset is specified as "all", each rotatable keyset will have
	all keypairs older than their respective primary keypairs distrusted.
	`))

	distrustKeypairExample = templates.Examples(i18n.T(`
	# Distrust all cluster CA keypairs older than the primary.
	kops distrust keypair kubernetes-ca

	# Distrust a particular keypair.
	kops distrust keypair kubernetes-ca 6977545226837259959403993899

	# Distrust all rotatable keypairs older than their respective primaries.
	kops distrust keypair all
	`))

	distrustKeypairShort = i18n.T(`Distrust a keypair.`)
)

type DistrustKeypairOptions struct {
	ClusterName string
	Keyset      string
	KeypairIDs  []string
}

func NewCmdDistrustKeypair(f *util.Factory, out io.Writer) *cobra.Command {
	options := &DistrustKeypairOptions{}

	cmd := &cobra.Command{
		Use:     "keypair {KEYSET [ID]... | all}",
		Short:   distrustKeypairShort,
		Long:    distrustKeypairLong,
		Example: distrustKeypairExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)
			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify name of keyset to distrust keypair in")
			}
			options.Keyset = args[0]

			if len(args) > 1 {
				if options.Keyset == "all" {
					return fmt.Errorf("cannot specify ID with \"all\"")
				}

				options.KeypairIDs = args[1:]
			}

			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeDistrustKeyset(f, options, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDistrustKeypair(context.TODO(), f, out, options)
		},
	}

	return cmd
}

func RunDistrustKeypair(ctx context.Context, f *util.Factory, out io.Writer, options *DistrustKeypairOptions) error {
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

	if options.Keyset != "all" {
		return distrustKeypair(out, options.Keyset, options.KeypairIDs[:], keyStore)
	}

	keysets, err := keyStore.ListKeysets()
	if err != nil {
		return fmt.Errorf("listing keysets: %v", err)
	}

	for name := range keysets {
		if rotatableKeysetFilter(name, nil) {
			if err := distrustKeypair(out, name, nil, keyStore); err != nil {
				return fmt.Errorf("distrusting keypair for %s: %v", name, err)
			}
		}
	}

	return nil
}

func distrustKeypair(out io.Writer, name string, keypairIDs []string, keyStore fi.CAStore) error {
	keyset, err := keyStore.FindKeyset(name)
	if err != nil {
		return err
	} else if keyset == nil {
		return fmt.Errorf("keyset not found")
	}

	if len(keypairIDs) == 0 {
		primarySerial := keyset.Primary.Certificate.Certificate.SerialNumber
		for id, item := range keyset.Items {
			if item.DistrustTimestamp == nil && item.Certificate.Certificate.SerialNumber.Cmp(primarySerial) < 0 {
				keypairIDs = append(keypairIDs, id)
			}
		}

		if len(keypairIDs) == 0 {
			klog.Infof("No %s keypairs older than the primary.", name)
			return nil
		}
	}

	for _, id := range keypairIDs {
		if id == keyset.Primary.Id {
			return fmt.Errorf("cannot distrust the primary keypair")
		}
		item := keyset.Items[id]
		if item == nil {
			return fmt.Errorf("keypair not found")
		}

		if item.DistrustTimestamp != nil {
			continue
		}

		now := time.Now().UTC().Round(0)
		item.DistrustTimestamp = &now

		if err := keyStore.StoreKeyset(name, keyset); err != nil {
			return fmt.Errorf("error storing keyset: %w", err)
		}

		fmt.Fprintf(out, "Distrusted %s %s\n", name, id)
	}

	return nil
}

func completeDistrustKeyset(f commandutils.Factory, options *DistrustKeypairOptions, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	commandutils.ConfigureKlogForCompletion()
	ctx := context.TODO()

	cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, f, nil)
	if cluster == nil {
		return completions, directive
	}

	keyset, _, completions, directive := completeKeyset(cluster, clientSet, args, rotatableKeysetFilter)
	if keyset == nil {
		return completions, directive
	}

	alreadySelected := sets.NewString(args[1:]...)
	return completeKeypairID(keyset, func(keyset *fi.Keyset, item *fi.KeysetItem) bool {
		return item.DistrustTimestamp == nil && item.Id != keyset.Primary.Id && !alreadySelected.Has(item.Id)
	})
}
