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
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"
)

var (
	getKeypairExample = templates.Examples(i18n.T(`
	# List the cluster CA keypairs.
	kops get keypairs kubernetes-ca

	# List the service-account keypairs, including distrusted ones.
	kops get keypairs service-account --distrusted`))

	getKeypairShort = i18n.T(`Get one or many keypairs.`)
)

type GetKeypairsOptions struct {
	*GetOptions
	KeysetNames []string
	Distrusted  bool
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
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)
			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			options.KeysetNames = args
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeGetKeypairs(f, options, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGetKeypairs(context.TODO(), f, out, options)
		},
	}

	cmd.Flags().BoolVar(&options.Distrusted, "distrusted", options.Distrusted, "Include distrusted keypairs")

	return cmd
}

type keypairItem struct {
	Name              string     `json:"name"`
	ID                string     `json:"id"`
	DistrustTimestamp *time.Time `json:"distrustTimestamp,omitempty"`
	IsPrimary         bool       `json:"isPrimary,omitempty"`
	Subject           string     `json:"subject"`
	Issuer            string     `json:"issuer"`
	AlternateNames    []string   `json:"alternateNames,omitempty"`
	IsCA              bool       `json:"isCA,omitempty"`
	NotBefore         time.Time  `json:"notBefore"`
	NotAfter          time.Time  `json:"notAfter"`
	KeyLength         *int       `json:"keyLength,omitempty"`
	HasPrivateKey     bool       `json:"hasPrivateKey,omitempty"`
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
				var alternateNames []string
				alternateNames = append(alternateNames, item.Certificate.Certificate.DNSNames...)
				alternateNames = append(alternateNames, item.Certificate.Certificate.EmailAddresses...)
				for _, ip := range item.Certificate.Certificate.IPAddresses {
					alternateNames = append(alternateNames, ip.String())
				}
				sort.Strings(alternateNames)

				keypair := keypairItem{
					Name:              name,
					ID:                item.Id,
					DistrustTimestamp: item.DistrustTimestamp,
					IsPrimary:         item.Id == keyset.Primary.Id,
					Subject:           item.Certificate.Subject.String(),
					Issuer:            item.Certificate.Certificate.Issuer.String(),
					AlternateNames:    alternateNames,
					IsCA:              item.Certificate.IsCA,
					NotBefore:         item.Certificate.Certificate.NotBefore.UTC(),
					NotAfter:          item.Certificate.Certificate.NotAfter.UTC(),
					HasPrivateKey:     item.PrivateKey != nil,
				}
				if rsaKey, ok := item.Certificate.PublicKey.(*rsa.PublicKey); ok {
					keypair.KeyLength = fi.Int(rsaKey.N.BitLen())
				}
				items = append(items, &keypair)
			}
		}
	}

	return items, nil
}

func RunGetKeypairs(ctx context.Context, f commandutils.Factory, out io.Writer, options *GetKeypairsOptions) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	items, err := listKeypairs(keyStore, options.KeysetNames, options.Distrusted)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("no keypairs found")
	}
	switch options.Output {

	case OutputTable:
		t := &tables.Table{}
		t.AddColumn("NAME", func(i *keypairItem) string {
			return i.Name
		})
		t.AddColumn("ID", func(i *keypairItem) string {
			return i.ID
		})
		t.AddColumn("DISTRUSTED", func(i *keypairItem) string {
			if i.DistrustTimestamp != nil {
				return i.DistrustTimestamp.Local().Format("2006-01-02")
			}
			return ""
		})
		t.AddColumn("ISSUED", func(i *keypairItem) string {
			return i.NotBefore.Local().Format("2006-01-02")
		})
		t.AddColumn("EXPIRES", func(i *keypairItem) string {
			return i.NotAfter.Local().Format("2006-01-02")
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
		y, err := yaml.Marshal(items)
		if err != nil {
			return fmt.Errorf("unable to marshal YAML: %v", err)
		}
		if _, err := out.Write(y); err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
	case OutputJSON:
		j, err := json.Marshal(items)
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %v", err)
		}
		if _, err := out.Write(j); err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}

	default:
		return fmt.Errorf("Unknown output format: %q", options.Output)
	}

	return nil
}

func completeGetKeypairs(f commandutils.Factory, options *GetKeypairsOptions, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	commandutils.ConfigureKlogForCompletion()
	ctx := context.TODO()

	cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, f, nil)
	if cluster == nil {
		return completions, directive
	}

	alreadySelected := sets.NewString(args...).Insert("all")
	_, _, completions, directive = completeKeyset(cluster, clientSet, nil, func(name string, keyset *fi.Keyset) bool {
		return !alreadySelected.Has(name)
	})

	return completions, directive
}
