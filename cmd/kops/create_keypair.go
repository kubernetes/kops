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
	"crypto/x509/pkix"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands/commandutils"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createKeypairLong = templates.LongDesc(i18n.T(`
	Add a CA certificate and private key to a keyset.

	If neither a certificate nor a private key is provided, a new self-signed
	certificate and private key will be generated.

	If no certificate is provided but a private key is, a self-signed
	certificate will be generated from the provided private key.

	If a certificate is provided but no private key is, the certificate
	will be added to the keyset without a private key. Such a certificate
	cannot be made primary.

	One of the certificate/private key pairs in each keyset must be primary.
	The primary keypair is the one used to issue certificates (or, for the
	"service-account" keyset, service-account tokens). As a consequence, a
	keypair added to an empty keyset must be made primary.

	If the keyset is specified as "all", a newly generated secondary
	certificate and private key will be added to each rotatable keyset.
	`))

	createKeypairExample = templates.Examples(i18n.T(`
	# Add a CA certificate and private key to a keyset.
	kops create keypair kubernetes-ca \
		--cert ~/ca.pem --key ~/ca-key.pem \
		--name k8s-cluster.example.com --state s3://my-state-store

	# Add a newly generated certificate and private key to each rotatable keyset.
	kops create keypair all \
		--name k8s-cluster.example.com --state s3://my-state-store
	`))

	createKeypairShort = i18n.T(`Add a CA certificate and private key to a keyset.`)
)

type CreateKeypairOptions struct {
	ClusterName    string
	Keyset         string
	PrivateKeyPath string
	CertPath       string
	Primary        bool
}

func rotatableKeysetFilter(name string, _ *fi.Keyset) bool {
	return name == "all" || name == "service-account" || strings.Contains(name, "-ca")
}

// NewCmdCreateKeypair returns a create keypair command.
func NewCmdCreateKeypair(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateKeypairOptions{}

	cmd := &cobra.Command{
		Use:     "keypair {KEYSET | all}",
		Short:   createKeypairShort,
		Long:    createKeypairLong,
		Example: createKeypairExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)

			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify name of keyset to add keypair to")
			}

			options.Keyset = args[0]

			if len(args) != 1 {
				return fmt.Errorf("can only add to one keyset at a time")
			}

			if options.Keyset == "all" {
				if options.CertPath != "" {
					return fmt.Errorf("cannot specify --cert with \"all\"")
				}
				if options.PrivateKeyPath != "" {
					return fmt.Errorf("cannot specify --key with \"all\"")
				}
				if options.Primary {
					return fmt.Errorf("cannot specify --primary with \"all\"")
				}
			}

			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeCreateKeypair(f, options, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunCreateKeypair(context.TODO(), f, out, options)
		},
	}

	cmd.Flags().StringVar(&options.CertPath, "cert", options.CertPath, "Path to CA certificate")
	cmd.Flags().StringVar(&options.PrivateKeyPath, "key", options.PrivateKeyPath, "Path to CA private key")
	cmd.Flags().BoolVar(&options.Primary, "primary", options.Primary, "Make the keypair the one used to issue certificates")

	return cmd
}

// RunCreateKeypair adds a custom CA certificate and private key.
func RunCreateKeypair(ctx context.Context, f *util.Factory, out io.Writer, options *CreateKeypairOptions) error {
	if !rotatableKeysetFilter(options.Keyset, nil) {
		return fmt.Errorf("adding keypair to %q is not supported", options.Keyset)
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return fmt.Errorf("error getting cluster: %q: %v", options.ClusterName, err)
	}

	clientSet, err := f.KopsClient()
	if err != nil {
		return fmt.Errorf("error getting clientset: %v", err)
	}

	keyStore, err := clientSet.KeyStore(cluster)
	if err != nil {
		return fmt.Errorf("error getting keystore: %v", err)
	}

	if options.Keyset != "all" {
		return createKeypair(out, options, options.Keyset, keyStore)
	}

	keysets, err := keyStore.ListKeysets()
	if err != nil {
		return fmt.Errorf("listing keysets: %v", err)
	}

	for name := range keysets {
		if rotatableKeysetFilter(name, nil) {
			if err := createKeypair(out, options, name, keyStore); err != nil {
				return fmt.Errorf("creating keypair for %s: %v", name, err)
			}
		}
	}

	return nil
}

func createKeypair(out io.Writer, options *CreateKeypairOptions, name string, keyStore fi.CAStore) error {
	var err error
	var privateKey *pki.PrivateKey
	if options.PrivateKeyPath != "" {
		options.PrivateKeyPath = utils.ExpandPath(options.PrivateKeyPath)
		privateKeyBytes, err := os.ReadFile(options.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("error reading user provided private key %q: %v", options.PrivateKeyPath, err)
		}

		privateKey, err = pki.ParsePEMPrivateKey(privateKeyBytes)
		if err != nil {
			return fmt.Errorf("error loading private key %q: %v", privateKeyBytes, err)
		}
	}

	var cert *pki.Certificate
	if options.CertPath == "" {
		if privateKey == nil {
			privateKey, err = pki.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("error generating private key: %v", err)
			}
		}

		serial := pki.BuildPKISerial(time.Now().UnixNano())
		req := pki.IssueCertRequest{
			Type:       "ca",
			Subject:    pkix.Name{CommonName: name, SerialNumber: serial.String()},
			Serial:     serial,
			PrivateKey: privateKey,
		}
		cert, _, _, err = pki.IssueCert(&req, nil)
		if err != nil {
			return fmt.Errorf("error issuing certificate: %v", err)
		}
	} else {
		options.CertPath = utils.ExpandPath(options.CertPath)
		certBytes, err := os.ReadFile(options.CertPath)
		if err != nil {
			return fmt.Errorf("error reading user provided cert %q: %v", options.CertPath, err)
		}

		cert, err = pki.ParsePEMCertificate(certBytes)
		if err != nil {
			return fmt.Errorf("error loading certificate %q: %v", options.CertPath, err)
		}
	}

	keyset, err := keyStore.FindKeyset(name)
	var item *fi.KeysetItem
	if os.IsNotExist(err) || (err == nil && keyset == nil) {
		if options.Primary {
			if keyset, err = fi.NewKeyset(cert, privateKey); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("the first keypair added to a keyset must be primary")
		}
		item = keyset.Primary
	} else if err != nil {
		return fmt.Errorf("reading existing keyset: %v", err)
	} else {
		item, err = keyset.AddItem(cert, privateKey, options.Primary)
	}
	if err != nil {
		return err
	}

	err = keyStore.StoreKeyset(name, keyset)
	if err != nil {
		return fmt.Errorf("error storing user provided keys %q %q: %v", options.CertPath, options.PrivateKeyPath, err)
	}

	if options.CertPath != "" {
		fmt.Fprintf(out, "using user provided cert: %v\n", options.CertPath)
	}
	if options.PrivateKeyPath != "" {
		fmt.Fprintf(out, "using user provided private key: %v\n", options.PrivateKeyPath)
	}
	fmt.Fprintf(out, "Created %s %s\n", name, item.Id)
	return nil
}

func completeKeyset(cluster *kopsapi.Cluster, clientSet simple.Clientset, args []string, filter func(name string, keyset *fi.Keyset) bool) (keyset *fi.Keyset, keyStore fi.CAStore, completions []string, directive cobra.ShellCompDirective) {
	keyStore, err := clientSet.KeyStore(cluster)
	if err != nil {
		completions, directive := commandutils.CompletionError("getting keystore", err)
		return nil, nil, completions, directive
	}

	if len(args) == 0 {
		list, err := keyStore.ListKeysets()
		if err != nil {
			completions, directive := commandutils.CompletionError("listing keystore", err)
			return nil, nil, completions, directive
		}

		var keysets []string
		for name, keyset := range list {
			if filter(name, keyset) {
				keysets = append(keysets, name)
			}
		}

		if filter("all", keyset) {
			keysets = append(keysets, "all")
		}

		return nil, nil, keysets, cobra.ShellCompDirectiveNoFileComp
	}

	keyset, err = keyStore.FindKeyset(args[0])
	if err != nil {
		completions, directive := commandutils.CompletionError("finding keyset", err)
		return nil, keyStore, completions, directive
	}

	return keyset, keyStore, nil, cobra.ShellCompDirectiveNoFileComp
}

func completeCreateKeypair(f commandutils.Factory, options *CreateKeypairOptions, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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

	if len(args) > 1 {
		return commandutils.CompletionError("too many arguments", nil)
	}

	var flags []string
	if options.CertPath == "" {
		flags = append(flags, "--cert")
	}
	if options.PrivateKeyPath == "" {
		flags = append(flags, "--key")
	}
	if !options.Primary && (options.CertPath == "" || options.PrivateKeyPath != "") {
		flags = append(flags, "--primary")
	}
	return flags, cobra.ShellCompDirectiveNoFileComp
}
