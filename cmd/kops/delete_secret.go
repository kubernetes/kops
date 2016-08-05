package main

import (
	"fmt"

	"github.com/spf13/cobra"
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
	if len(args) != 2 {
		return fmt.Errorf("Syntax: <type> <name>")
	}

	secretType := args[0]
	secretName := args[1]

	secrets, err := listSecrets(secretType, []string{secretName})
	if err != nil {
		return err
	}

	if len(secrets) == 0 {
		return fmt.Errorf("secret %q not found")
	}

	if len(secrets) != 1 {
		return fmt.Errorf("found multiple matching secrets")
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
