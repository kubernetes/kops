/*
Copyright 2016 The Kubernetes Authors.

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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	create_secret_cacert_long = templates.LongDesc(i18n.T(`
	Add a intermediate ca certificate kay private key. and store the key
	in state store.`))

	create_secret_cacert_example = templates.Examples(i18n.T(`
	# Add a intermediate ca certificate kay private key. and store the key
	kops create secret ca-cert \
		--ca-cert ~/ca.pem \
		--ca-cert-key ~/ca-key.pem \
		--name k8s-cluster.example.com \
		--state s3://example.com
	`))

	create_secret_cacert_short = i18n.T(`Add intermediate ca cert and private key.`)
)

type CreateSecretCaCertOptions struct {
	ClusterName    string
	PrivateKeyPath string
	CaCertPath     string
}

func NewCmdCreateSecretCaCert(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretCaCertOptions{}

	cmd := &cobra.Command{
		Use:     "ca-cert",
		Short:   create_secret_cacert_short,
		Long:    create_secret_cacert_long,
		Example: create_secret_cacert_example,
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

	cmd.Flags().StringVar(&options.CaCertPath, "ca-cert", options.CaCertPath, "Path to ca cert")
	cmd.Flags().StringVar(&options.PrivateKeyPath, "ca-cert-key", options.PrivateKeyPath, "Path to ca cert private key")

	return cmd
}

func RunCreateSecretCaCert(f *util.Factory, out io.Writer, options *CreateSecretCaCertOptions) error {
	if options.CaCertPath == "" {
		return fmt.Errorf("No ca cert provided")
	}

	if options.PrivateKeyPath == "" {
		return fmt.Errorf("No ca cert private key provided")
	}

	cluster, err := GetCluster(f, options.ClusterName)
	if err != nil {
		return err
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	options.CaCertPath = utils.ExpandPath(options.CaCertPath)
	options.PrivateKeyPath = utils.ExpandPath(options.PrivateKeyPath)

	caCertByteArr, err := ioutil.ReadFile(options.CaCertPath)
	if err != nil {
		return fmt.Errorf("error reading user provided CaCert %q: %v", options.CaCertPath, err)
	}
	caCertPrivateKeyByteArr, err := ioutil.ReadFile(options.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error reading user provided CaCert Private Key %q: %v", options.PrivateKeyPath, err)
	}

	caPrivateKey, err := fi.LoadPEMPrivateKey(caCertPrivateKeyByteArr)
	if err != nil {
		return fmt.Errorf("error loading ca private key %q: %v", caCertPrivateKeyByteArr, err)
	}
	caCert, err := fi.LoadPEMCertificate(caCertByteArr)
	if err != nil {
		return fmt.Errorf("error loading ca certificate %q: %v", options.CaCertPath, err)
	}

	err = keyStore.StoreKeypair(fi.CertificateId_CA, caCert, caPrivateKey)
	if err != nil {
		return fmt.Errorf("error storing user provided keys %q %q: %v", options.CaCertPath, options.PrivateKeyPath, err)
	}

	glog.Infof("Using User Provided CaCert: %v\n", options.CaCertPath)
	glog.Infof("Using User Provided Private Key: %v\n", options.PrivateKeyPath)

	return nil
}
