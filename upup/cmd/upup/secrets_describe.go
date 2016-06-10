package main

import (
	"fmt"

	"bytes"
	"crypto/rsa"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

type DescribeSecretsCommand struct {
}

var describeSecretsCommand DescribeSecretsCommand

func init() {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe secrets",
		Long:  `Describe secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := describeSecretsCommand.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	secretsCmd.AddCommand(cmd)
}

func (c *DescribeSecretsCommand) Run() error {

	w := new(tabwriter.Writer)
	var b bytes.Buffer

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', tabwriter.StripEscape)

	{
		caStore, err := rootCommand.CA()
		if err != nil {
			return err
		}
		ids, err := caStore.List()
		if err != nil {
			return fmt.Errorf("error listing CA store items %v", err)
		}

		for _, id := range ids {
			cert, err := caStore.FindCert(id)
			if err != nil {
				return fmt.Errorf("error retrieving cert %q: %v", id, err)
			}

			key, err := caStore.FindPrivateKey(id)
			if err != nil {
				return fmt.Errorf("error retrieving private key %q: %v", id, err)
			}

			if key == nil && cert == nil {
				continue
			}

			err = describeKeypair(id, cert, key, &b)
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

	}

	{
		secretStore, err := rootCommand.Secrets()
		if err != nil {
			return err
		}
		ids, err := secretStore.ListSecrets()
		if err != nil {
			return fmt.Errorf("error listing secrets %v", err)
		}

		for _, id := range ids {
			secret, err := secretStore.FindSecret(id)
			if err != nil {
				return fmt.Errorf("error retrieving secret %q: %v", id, err)
			}

			if secret == nil {
				continue
			}

			err = describeSecret(id, secret, &b)
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
	}

	return w.Flush()
}

func describeKeypair(id string, c *fi.Certificate, k *fi.PrivateKey, w *bytes.Buffer) error {
	var alternateNames []string
	if c != nil {
		alternateNames = append(alternateNames, c.Certificate.DNSNames...)
		alternateNames = append(alternateNames, c.Certificate.EmailAddresses...)
		for _, ip := range c.Certificate.IPAddresses {
			alternateNames = append(alternateNames, ip.String())
		}
		sort.Strings(alternateNames)
	}

	fmt.Fprintf(w, "Id:\t%s\n", id)
	if c != nil && k != nil {
		fmt.Fprintf(w, "Type:\t%s\n", "keypair")
	} else if c != nil && k == nil {
		fmt.Fprintf(w, "Type:\t%s\n", "certificate")
	} else if k != nil && c == nil {
		// Unexpected!
		fmt.Fprintf(w, "Type:\t%s\n", "privatekey")
	} else {
		return fmt.Errorf("expected either certificate or key to be set")
	}

	if c != nil {
		fmt.Fprintf(w, "Subject:\t%s\n", pkixNameToString(&c.Certificate.Subject))
		fmt.Fprintf(w, "Issuer:\t%s\n", pkixNameToString(&c.Certificate.Issuer))
		fmt.Fprintf(w, "AlternateNames:\t%s\n", strings.Join(alternateNames, ", "))
		fmt.Fprintf(w, "CA:\t%v\n", c.IsCA)
		fmt.Fprintf(w, "NotAfter:\t%s\n", c.Certificate.NotAfter)
		fmt.Fprintf(w, "NotBefore:\t%s\n", c.Certificate.NotBefore)

		// PublicKeyAlgorithm doesn't have a String() function.  Also, is this important information?
		//fmt.Fprintf(w, "PublicKeyAlgorithm:\t%v\n", c.Certificate.PublicKeyAlgorithm)
		//fmt.Fprintf(w, "SignatureAlgorithm:\t%v\n", c.Certificate.SignatureAlgorithm)
	}

	if k != nil {
		if rsaPrivateKey, ok := k.Key.(*rsa.PrivateKey); ok {
			fmt.Fprintf(w, "PrivateKeyType:\t%v\n", "rsa")
			fmt.Fprintf(w, "KeyLength:\t%v\n", rsaPrivateKey.N.BitLen())
		} else {
			fmt.Fprintf(w, "PrivateKeyType:\tunknown (%T)\n", k.Key)
		}
	}

	return nil
}

func describeSecret(id string, s *fi.Secret, w *bytes.Buffer) error {
	fmt.Fprintf(w, "Id:\t%s\n", id)
	fmt.Fprintf(w, "Type:\t%s\n", "secret")

	return nil
}
