package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"os"
)

type ExposeSecretsCommand struct {
	ID   string
	Type string
}

var exposeSecretsCommand ExposeSecretsCommand

func init() {
	cmd := &cobra.Command{
		Use:   "expose",
		Short: "Expose secrets",
		Long:  `Expose secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := exposeSecretsCommand.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	secretsCmd.AddCommand(cmd)

	cmd.Flags().StringVarP(&exposeSecretsCommand.Type, "type", "", "", "Type of secret to create")
	cmd.Flags().StringVarP(&exposeSecretsCommand.ID, "id", "", "", "Id of secret to create")
}

func (cmd *ExposeSecretsCommand) Run() error {
	id := cmd.ID
	if id == "" {
		return fmt.Errorf("id is required")
	}

	if cmd.Type == "" {
		return fmt.Errorf("type is required")
	}

	var value string
	switch cmd.Type {
	case "secret":
		{
			secretStore, err := rootCommand.Secrets()
			if err != nil {
				return err
			}
			secret, err := secretStore.FindSecret(id)
			if err != nil {
				return fmt.Errorf("error finding secret %q: %v", id, err)
			}
			if secret == nil {
				return fmt.Errorf("secret not found: %q", id)
			}
			value = string(secret.Data)
		}

	case "certificate", "privatekey":
		{
			caStore, err := rootCommand.CA()
			if err != nil {
				return fmt.Errorf("error building CA store: %v", err)
			}

			if cmd.Type == "privatekey" {
				k, err := caStore.FindPrivateKey(id)
				if err != nil {
					return fmt.Errorf("error finding privatekey: %v", err)
				}
				if k == nil {
					return fmt.Errorf("privatekey not found: %q", id)
				}
				value, err = k.AsString()
				if err != nil {
					return fmt.Errorf("error encoding privatekey: %v", err)
				}
			} else {
				c, err := caStore.FindCert(id)
				if err != nil {
					return fmt.Errorf("error finding certificate: %v", err)
				}
				if c == nil {
					return fmt.Errorf("certificate not found: %q", id)
				}
				value, err = c.AsString()
				if err != nil {
					return fmt.Errorf("error encoding certiifcate: %v", err)
				}
			}
		}

	default:
		return fmt.Errorf("secret type not known: %q", cmd.Type)
	}

	_, err := fmt.Fprint(os.Stdout, value)
	if err != nil {
		return fmt.Errorf("error writing to output: %v", err)
	}

	return nil

}
