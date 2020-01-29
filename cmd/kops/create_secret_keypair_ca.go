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
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createSecretCacertLong = templates.LongDesc(i18n.T(`
	Add a ca certificate and private key.
    `))

	createSecretCacertExample = templates.Examples(i18n.T(`
	Add a ca certificate and private key.
	kops create secret keypair ca \
		--cert ~/ca.pem --key ~/ca-key.pem \
		--name k8s-cluster.example.com --state s3://example.com
	`))

	createSecretCacertShort = i18n.T(`Add a ca cert and key`)
)

type CreateSecretCaCertOptions struct {
	ClusterName      string
	CaPrivateKeyPath string
	CaCertPath       string
}

// NewCmdCreateSecretCaCert returns create ca certificate command
func NewCmdCreateSecretCaCert(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretCaCertOptions{}

	cmd := &cobra.Command{
		Use:     "ca",
		Short:   createSecretCacertShort,
		Long:    createSecretCacertLong,
		Example: createSecretCacertExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunCreateSecretCaCert(f, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.CaCertPath, "cert", options.CaCertPath, "Path to ca cert")
	cmd.Flags().StringVar(&options.CaPrivateKeyPath, "key", options.CaPrivateKeyPath, "Path to ca cert private key")

	return cmd
}

// RunCreateSecretCaCert adds a custom ca certificate and private key
func RunCreateSecretCaCert(f *util.Factory, out io.Writer, options *CreateSecretCaCertOptions) error {
	if options.CaCertPath == "" {
		return fmt.Errorf("error cert provided")
	}

	if options.CaPrivateKeyPath == "" {
		return fmt.Errorf("error no private key provided")
	}

	cluster, err := GetCluster(f, options.ClusterName)
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

	options.CaCertPath = utils.ExpandPath(options.CaCertPath)
	options.CaPrivateKeyPath = utils.ExpandPath(options.CaPrivateKeyPath)

	certBytes, err := ioutil.ReadFile(options.CaCertPath)
	if err != nil {
		return fmt.Errorf("error reading user provided cert %q: %v", options.CaCertPath, err)
	}
	privateKeyBytes, err := ioutil.ReadFile(options.CaPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error reading user provided private key %q: %v", options.CaPrivateKeyPath, err)
	}

	privateKey, err := pki.ParsePEMPrivateKey(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("error loading private key %q: %v", privateKeyBytes, err)
	}
	cert, err := pki.ParsePEMCertificate(certBytes)
	if err != nil {
		return fmt.Errorf("error loading certificate %q: %v", options.CaCertPath, err)
	}

	err = keyStore.StoreKeypair(fi.CertificateId_CA, cert, privateKey)
	if err != nil {
		return fmt.Errorf("error storing user provided keys %q %q: %v", options.CaCertPath, options.CaPrivateKeyPath, err)
	}

	klog.Infof("using user provided cert: %v\n", options.CaCertPath)
	klog.Infof("using user provided private key: %v\n", options.CaPrivateKeyPath)

	return nil
}
