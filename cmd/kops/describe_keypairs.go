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
	"bytes"
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	describeKeypairLong = templates.LongDesc(i18n.T(`
	Get additional information about keypairs.
	`))

	describeKeypairExample = templates.Examples(i18n.T(`
	# Describe a keypair
	kops describe keypairs ca
	`))
	describeKeypairShort = i18n.T(`Describe a cluster keypair`)
)

type DescribeKeypairsCommand struct {
}

var describeKeypairsCommand DescribeKeypairsCommand

func init() {
	cmd := &cobra.Command{
		Use:     "keypairs",
		Aliases: []string{"keypair"},
		Short:   describeKeypairShort,
		Long:    describeKeypairLong,
		Example: describeKeypairExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()
			err := describeKeypairsCommand.Run(ctx, args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	describeCmd.cobraCommand.AddCommand(cmd)
}

func (c *DescribeKeypairsCommand) Run(ctx context.Context, args []string) error {
	cluster, err := rootCommand.Cluster(ctx)
	if err != nil {
		return err
	}

	clientset, err := rootCommand.Clientset()
	if err != nil {
		return err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return err
	}

	items, err := listKeypairs(keyStore, args)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Fprintf(os.Stderr, "No keypairs found\n")

		return nil
	}

	w := new(tabwriter.Writer)
	var b bytes.Buffer

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.StripEscape)

	for _, i := range items {
		fmt.Fprintf(w, "Name:\t%s\n", i.Name)
		fmt.Fprintf(w, "Type:\t%s\n", i.Type)
		fmt.Fprintf(w, "Id:\t%s\n", i.ID)

		err = describeKeypair(keyStore, i, &b)
		if err != nil {
			return err
		}

		b.WriteString("\n")
		_, err = w.Write(b.Bytes())
		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}

		b.Reset()
	}

	return w.Flush()
}

func describeKeypair(keyStore fi.CAStore, item *fi.KeystoreItem, w *bytes.Buffer) error {
	name := item.Name

	cert, err := keyStore.FindCert(name)
	if err != nil {
		return fmt.Errorf("error retrieving cert %q: %v", name, err)
	}

	key, err := keyStore.FindPrivateKey(name)
	if err != nil {
		return fmt.Errorf("error retrieving private key %q: %v", name, err)
	}

	var alternateNames []string
	if cert != nil {
		alternateNames = append(alternateNames, cert.Certificate.DNSNames...)
		alternateNames = append(alternateNames, cert.Certificate.EmailAddresses...)
		for _, ip := range cert.Certificate.IPAddresses {
			alternateNames = append(alternateNames, ip.String())
		}
		sort.Strings(alternateNames)
	}

	if cert != nil {
		fmt.Fprintf(w, "Subject:\t%s\n", pki.PkixNameToString(&cert.Certificate.Subject))
		fmt.Fprintf(w, "Issuer:\t%s\n", pki.PkixNameToString(&cert.Certificate.Issuer))
		fmt.Fprintf(w, "AlternateNames:\t%s\n", strings.Join(alternateNames, ", "))
		fmt.Fprintf(w, "CA:\t%v\n", cert.IsCA)
		fmt.Fprintf(w, "NotAfter:\t%s\n", cert.Certificate.NotAfter)
		fmt.Fprintf(w, "NotBefore:\t%s\n", cert.Certificate.NotBefore)

		// PublicKeyAlgorithm doesn't have a String() function.  Also, is this important information?
		//fmt.Fprintf(w, "PublicKeyAlgorithm:\t%v\n", c.Certificate.PublicKeyAlgorithm)
		//fmt.Fprintf(w, "SignatureAlgorithm:\t%v\n", c.Certificate.SignatureAlgorithm)
	}

	if key != nil {
		if rsaPrivateKey, ok := key.Key.(*rsa.PrivateKey); ok {
			fmt.Fprintf(w, "PrivateKeyType:\t%v\n", "rsa")
			fmt.Fprintf(w, "KeyLength:\t%v\n", rsaPrivateKey.N.BitLen())
		} else {
			fmt.Fprintf(w, "PrivateKeyType:\tunknown (%T)\n", key.Key)
		}
	}

	return nil
}
