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
	"k8s.io/klog/v2"

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

var keysetCommonNames = map[string]string{
	"ca":              "kubernetes",
	"service-account": "service-account",
}

// NewCmdCreateKeypair returns a create keypair command.
func NewCmdCreateKeypair(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateKeypairOptions{}

	cmd := &cobra.Command{
		Use:     "keypair KEYSET",
		Short:   createKeypairShort,
		Long:    createKeypairLong,
		Example: createKeypairExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			options.ClusterName = rootCommand.ClusterName()

			if options.ClusterName == "" {
				exitWithError(fmt.Errorf("--name is required"))
				return
			}

			if len(args) == 0 {
				exitWithError(fmt.Errorf("must specify name of keyset to add keypair to"))
			}
			if len(args) != 1 {
				exitWithError(fmt.Errorf("can only add to one keyset at a time"))
			}
			options.Keyset = args[0]

			err := RunCreateKeypair(ctx, f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.CertPath, "cert", options.CertPath, "Path to CA certificate")
	cmd.Flags().StringVar(&options.PrivateKeyPath, "key", options.PrivateKeyPath, "Path to CA private key")
	cmd.Flags().BoolVar(&options.Primary, "primary", options.Primary, "Make the keypair the one used to issue certificates")

	return cmd
}

// RunCreateKeypair adds a custom CA certificate and private key.
func RunCreateKeypair(ctx context.Context, f *util.Factory, out io.Writer, options *CreateKeypairOptions) error {
	commonName := keysetCommonNames[options.Keyset]
	if commonName == "" {
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
