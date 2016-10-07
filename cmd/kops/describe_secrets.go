package main

import (
	"fmt"

	"bytes"
	"crypto/rsa"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api/registry"
	"k8s.io/kops/upup/pkg/fi"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

type DescribeSecretsCommand struct {
	Type string
}

var describeSecretsCommand DescribeSecretsCommand

func init() {
	cmd := &cobra.Command{
		Use:     "secrets",
		Aliases: []string{"secret"},
		Short:   "Describe secrets",
		Long:    `Describe secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := describeSecretsCommand.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	describeCmd.cobraCommand.AddCommand(cmd)

	cmd.Flags().StringVarP(&describeSecretsCommand.Type, "type", "", "", "Filter by secret type")
}

func (c *DescribeSecretsCommand) Run(args []string) error {
	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	secretStore, err := registry.SecretStore(cluster)
	if err != nil {
		return err
	}

	items, err := listSecrets(keyStore, secretStore, c.Type, args)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Fprintf(os.Stderr, "No secrets found\n")

		return nil
	}

	w := new(tabwriter.Writer)
	var b bytes.Buffer

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.StripEscape)

	for _, i := range items {
		fmt.Fprintf(w, "Name:\t%s\n", i.Name)
		fmt.Fprintf(w, "Type:\t%s\n", i.Type)
		fmt.Fprintf(w, "Id:\t%s\n", i.Id)

		switch i.Type {
		case fi.SecretTypeKeypair:
			err = describeKeypair(keyStore, i, &b)
			if err != nil {
				return err
			}

		case fi.SecretTypeSSHPublicKey:
			err = describeSSHPublicKey(i, &b)
			if err != nil {
				return err
			}

		case fi.SecretTypeSecret:
			err = describeSecret(i, &b)
			if err != nil {
				return err
			}
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
		fmt.Fprintf(w, "Subject:\t%s\n", pkixNameToString(&cert.Certificate.Subject))
		fmt.Fprintf(w, "Issuer:\t%s\n", pkixNameToString(&cert.Certificate.Issuer))
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

func describeSecret(item *fi.KeystoreItem, w *bytes.Buffer) error {
	return nil
}

func describeSSHPublicKey(item *fi.KeystoreItem, w *bytes.Buffer) error {
	return nil
}
