package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/fi"
)

type DeleteSecretCmd struct {
}

var deleteSecretCmd DeleteSecretCmd

func init() {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Delete secret",
		Long:  `Delete a secret.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := deleteSecretCmd.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	deleteCmd.AddCommand(cmd)
}

func (c *DeleteSecretCmd) Run(args []string) error {
	if len(args) != 2 && len(args) != 3 {
		return fmt.Errorf("Syntax: <type> <name> [<id>]")
	}

	secretType := args[0]
	secretName := args[1]

	secretID := ""
	if len(args) == 3 {
		secretID = args[2]
	}

	secrets, err := listSecrets(secretType, []string{secretName})
	if err != nil {
		return err
	}

	if secretID != "" {
		var matches []*fi.KeystoreItem
		for _, s := range secrets {
			if s.Id == secretID {
				matches = append(matches, s)
			}
		}
		secrets = matches
	}

	if len(secrets) == 0 {
		return fmt.Errorf("secret %q not found")
	}

	if len(secrets) != 1 {
		// TODO: it would be friendly to print the matching keys
		return fmt.Errorf("found multiple matching secrets; specify the id of the key")
	}

	keyStore, err := rootCommand.KeyStore()
	if err != nil {
		return err
	}

	err = keyStore.DeleteSecret(secrets[0])
	if err != nil {
		return fmt.Errorf("error deleting secret: %v", err)
	}

	return nil
}
