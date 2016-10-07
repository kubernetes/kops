package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/api/registry"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/tables"
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

	cmd.Flags().StringVarP(&getSecretsCommand.Type, "type", "", "", "Filter by secret type")
}

func listSecrets(keyStore fi.CAStore, secretStore fi.SecretStore, secretType string, names []string) ([]*fi.KeystoreItem, error) {
	var items []*fi.KeystoreItem

	findType := strings.ToLower(secretType)
	switch findType {
	case "":
	// OK
	case "sshpublickey", "keypair", "secret":
	// OK
	default:
		return nil, fmt.Errorf("unknown secret type %q", secretType)
	}

	{
		l, err := keyStore.List()
		if err != nil {
			return nil, fmt.Errorf("error listing CA store items %v", err)
		}

		for _, i := range l {
			if findType != "" && findType != strings.ToLower(i.Type) {
				continue
			}
			items = append(items, i)
		}
	}

	if findType == "" || findType == strings.ToLower(fi.SecretTypeSecret) {
		l, err := secretStore.ListSecrets()
		if err != nil {
			return nil, fmt.Errorf("error listing secrets %v", err)
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

	if len(names) != 0 {
		var matches []*fi.KeystoreItem
		for _, arg := range names {
			var found []*fi.KeystoreItem
			for _, i := range items {
				// There may be multiple secrets with the same name (of different type)
				if i.Name == arg {
					found = append(found, i)
				}
			}

			if len(found) == 0 {
				return nil, fmt.Errorf("Secret not found: %q", arg)
			}

			matches = append(matches, found...)
		}
		items = matches
	}

	return items, nil
}

func (c *GetSecretsCommand) Run(args []string) error {
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

	output := getCmd.output
	if output == OutputTable {
		t := &tables.Table{}
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
