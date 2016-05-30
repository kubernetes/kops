package main

import (
	"fmt"

	"crypto/x509"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"net"
	"path"
	"strings"
)

type CreateSecretsCommand struct {
	StateDir string

	Id   string
	Type string

	Usage          string
	Subject        string
	AlternateNames []string
}

var createSecretsCommand CreateSecretsCommand

func init() {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create secrets",
		Long:  `Create secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := createSecretsCommand.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	secretsCmd.AddCommand(cmd)

	cmd.Flags().StringVarP(&createSecretsCommand.StateDir, "state", "", "", "Directory in which to store state")
	cmd.Flags().StringVarP(&createSecretsCommand.Type, "type", "", "", "Type of secret to create")
	cmd.Flags().StringVarP(&createSecretsCommand.Id, "id", "", "", "Id of secret to create")
	cmd.Flags().StringVarP(&createSecretsCommand.Usage, "usage", "", "", "Usage of secret (for SSL certificate)")
	cmd.Flags().StringVarP(&createSecretsCommand.Subject, "subject", "", "", "Subject (for SSL certificate)")
	cmd.Flags().StringSliceVarP(&createSecretsCommand.AlternateNames, "san", "", nil, "Alternate name (for SSL certificate)")
}

func (cmd *CreateSecretsCommand) Run() error {
	if cmd.StateDir == "" {
		return fmt.Errorf("state dir is required")
	}

	if cmd.Id == "" {
		return fmt.Errorf("id is required")
	}

	if cmd.Type == "" {
		return fmt.Errorf("type is required")
	}

	// TODO: Prompt before replacing?
	// TODO: Keep history?

	switch cmd.Type {
	case "secret":
		{
			secretStore, err := fi.NewFilesystemSecretStore(path.Join(cmd.StateDir, "secrets"))
			if err != nil {
				return fmt.Errorf("error building secret store: %v", err)
			}
			_, err = secretStore.CreateSecret(cmd.Id)
			if err != nil {
				return fmt.Errorf("error creating secrets %v", err)
			}
			return nil
		}

	case "keypair":
		// TODO: Create a rotate command which keeps the same values?
		// Or just do it here a "replace" action - existing=fail, replace or rotate
		// TODO: Create a CreateKeypair class, move to fi (this is duplicated code)
		{
			if cmd.Subject == "" {
				return fmt.Errorf("subject is required")
			}

			subject, err := parsePkixName(cmd.Subject)
			if err != nil {
				return fmt.Errorf("Error parsing subject: %v", err)
			}
			template := &x509.Certificate{
				Subject:               *subject,
				BasicConstraintsValid: true,
				IsCA: false,
			}

			if len(template.Subject.ToRDNSequence()) == 0 {
				return fmt.Errorf("Subject name was empty")
			}

			switch cmd.Usage {
			case "client":
				template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
				template.KeyUsage = x509.KeyUsageDigitalSignature
				break

			case "server":
				template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
				template.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
				break

			default:
				return fmt.Errorf("unknown usage: %q", cmd.Usage)
			}

			for _, san := range cmd.AlternateNames {
				san = strings.TrimSpace(san)
				if san == "" {
					continue
				}
				if ip := net.ParseIP(san); ip != nil {
					template.IPAddresses = append(template.IPAddresses, ip)
				} else {
					template.DNSNames = append(template.DNSNames, san)
				}
			}

			caStore, err := fi.NewFilesystemCAStore(path.Join(cmd.StateDir, "pki"))
			if err != nil {
				return fmt.Errorf("error building CA store: %v", err)
			}

			// TODO: Allow resigning of the existing private key?

			key, err := caStore.CreatePrivateKey(cmd.Id)
			if err != nil {
				return fmt.Errorf("error creating privatekey %v", err)
			}
			_, err = caStore.IssueCert(cmd.Id, key, template)
			if err != nil {
				return fmt.Errorf("error creating certificate %v", err)
			}
			return nil
		}

	default:
		return fmt.Errorf("secret type not known: %q", cmd.Type)
	}
}
