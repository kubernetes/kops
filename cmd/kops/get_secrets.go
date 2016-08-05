package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/fi"
	"strings"
)

type GetSecretsCommand struct {
	Type string
}

var getSecretsCommand GetSecretsCommand

func init() {
	cmd := &cobra.Command{
		Use:     "secrets",
		Aliases: []string{"secret"},
		Short:   "get secrets",
		Long:    `List or get secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getSecretsCommand.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	getCmd.cobraCommand.AddCommand(cmd)

	cmd.Flags().StringVarP(&getSecretsCommand.Type, "type", "", "", "Type of secret to create")
}

func (c *GetSecretsCommand) Run(args []string) error {
	var items []*fi.KeystoreItem

	findType := strings.ToLower(c.Type)
	switch findType {
	case "":
	// OK
	case "sshpublickey", "keypair", "secret":
	// OK
	default:
		return fmt.Errorf("unknown secret type %q", c.Type)
	}

	{
		caStore, err := rootCommand.KeyStore()
		if err != nil {
			return err
		}
		l, err := caStore.List()
		if err != nil {
			return fmt.Errorf("error listing CA store items %v", err)
		}

		for _, i := range l {
			if findType != "" && findType != strings.ToLower(i.Type) {
				continue
			}
			items = append(items, i)
		}
	}

	secretStore, err := rootCommand.SecretStore()
	if err != nil {
		return err
	}

	if findType == "" || findType == strings.ToLower(fi.SecretTypeSecret) {
		l, err := secretStore.ListSecrets()
		if err != nil {
			return fmt.Errorf("error listing secrets %v", err)
		}

		for _, id := range l {
			i := &fi.KeystoreItem{
				Name: id,
				Type: fi.SecretTypeSecret,
			}
			if findType != "" && findType != strings.ToLower(i.Type) {
				continue
			}

			items = append(items, i)
		}
	}

	if len(args) != 0 {
		var matches []*fi.KeystoreItem
		for _, arg := range args {
			var found []*fi.KeystoreItem
			for _, i := range items {
				// There may be multiple secrets with the same name (of different type)
				if i.Name == arg {
					found = append(found, i)
				}
			}

			if len(found) == 0 {
				return fmt.Errorf("Secret not found: %q", arg)
			}

			matches = append(matches, found...)
		}
		items = matches
	}

	if len(items) == 0 {
		fmt.Fprintf(os.Stdout, "No secrets found\n")

		return nil
	}

	output := getCmd.output
	if output == OutputTable {
		t := &Table{}
		t.AddColumn("NAME", func(i *fi.KeystoreItem) string {
			return i.Name
		})
		t.AddColumn("ID", func(i *fi.KeystoreItem) string {
			return i.Id
		})
		t.AddColumn("TYPE", func(i *fi.KeystoreItem) string {
			return i.Type
		})
		return t.Render(items, os.Stdout, "TYPE", "NAME", "ID")
	} else if output == OutputYaml {
		return fmt.Errorf("yaml output format is not (currently) supported for secrets")
	} else if output == "plaintext" {
		for _, i := range items {
			var data string
			switch i.Type {
			case fi.SecretTypeSecret:
				secret, err := secretStore.FindSecret(i.Name)
				if err != nil {
					return fmt.Errorf("error getting secret %q: %v", i.Name, err)
				}
				if secret == nil {
					return fmt.Errorf("cannot find secret %q", i.Name)
				}
				data = string(secret.Data)

			default:
				return fmt.Errorf("secret type %v cannot (currently) be exported as plaintext", i.Type)
			}

			_, err := fmt.Fprintf(os.Stdout, "%s\n", data)
			if err != nil {
				return fmt.Errorf("error writing output: %v", err)
			}
		}
		return nil
	} else {
		return fmt.Errorf("Unknown output format: %q", output)
	}

}
