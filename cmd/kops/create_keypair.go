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
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
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
	"service-account" keyset, service-account tokens). As a consequence, the
	first entry in a keyset must be made primary.
	`))

	createKeypairExample = templates.Examples(i18n.T(`
	Add a CA certificate and private key to a keyset.
	kops create keypair ca \
		--cert ~/ca.pem --key ~/ca-key.pem \
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

var rotatableKeysets = sets.NewString(
	"ca",
	"etcd-clients-ca-cilium",
	"service-account",
)

func rotatableKeysetFilter(name string, _ *fi.Keyset) bool {
	return rotatableKeysets.Has(name)
}

// NewCmdCreateKeypair returns a create keypair command.
func NewCmdCreateKeypair(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateKeypairOptions{}

	cmd := &cobra.Command{
		Use:     "keypair keyset",
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

			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeCreateKeyset(options, args, toComplete)
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
	if !rotatableKeysets.Has(options.Keyset) {
		return fmt.Errorf("adding keypair to %q is not supported", options.Keyset)
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

	var privateKey *pki.PrivateKey
	if options.PrivateKeyPath != "" {
		options.PrivateKeyPath = utils.ExpandPath(options.PrivateKeyPath)
		privateKeyBytes, err := ioutil.ReadFile(options.PrivateKeyPath)
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

		commonName := options.Keyset
		if commonName == "ca" {
			commonName = "kubernetes"
		}
		req := pki.IssueCertRequest{
			Type:       "ca",
			Subject:    pkix.Name{CommonName: "cn=" + commonName},
			Serial:     pki.BuildPKISerial(time.Now().UnixNano()),
			PrivateKey: privateKey,
		}
		cert, _, _, err = pki.IssueCert(&req, nil)
		if err != nil {
			return fmt.Errorf("error issuing certificate: %v", err)
		}
	} else {
		options.CertPath = utils.ExpandPath(options.CertPath)
		certBytes, err := ioutil.ReadFile(options.CertPath)
		if err != nil {
			return fmt.Errorf("error reading user provided cert %q: %v", options.CertPath, err)
		}

		cert, err = pki.ParsePEMCertificate(certBytes)
		if err != nil {
			return fmt.Errorf("error loading certificate %q: %v", options.CertPath, err)
		}
	}

	keyset, err := keyStore.FindKeyset(options.Keyset)
	if os.IsNotExist(err) || (err == nil && keyset == nil) {
		if options.Primary {
			keyset, err = fi.NewKeyset(cert, privateKey)
		} else {
			return fmt.Errorf("the first keypair added to a keyset must be primary")
		}
	} else if err != nil {
		return fmt.Errorf("reading existing keyset: %v", err)
	} else {
		err = keyset.AddItem(cert, privateKey, options.Primary)
	}
	if err != nil {
		return err
	}

	err = keyStore.StoreKeyset(options.Keyset, keyset)
	if err != nil {
		return fmt.Errorf("error storing user provided keys %q %q: %v", options.CertPath, options.PrivateKeyPath, err)
	}

	if options.CertPath != "" {
		klog.Infof("using user provided cert: %v\n", options.CertPath)
	}
	if options.PrivateKeyPath != "" {
		klog.Infof("using user provided private key: %v\n", options.PrivateKeyPath)
	}
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

		return nil, nil, keysets, cobra.ShellCompDirectiveNoFileComp
	}

	keyset, err = keyStore.FindKeyset(args[0])
	if err != nil {
		completions, directive := commandutils.CompletionError("finding keyset", err)
		return nil, nil, completions, directive
	}

	return keyset, keyStore, nil, 0
}

func completeCreateKeyset(options *CreateKeypairOptions, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	commandutils.ConfigureKlogForCompletion()
	ctx := context.TODO()

	cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, &rootCommand)
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
