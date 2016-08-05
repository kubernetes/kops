package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"io/ioutil"
)

type CreateSecretPublickeyCommand struct {
	cobraCommand *cobra.Command

	Pubkey string
}

var createSecretPublickeyCommand = CreateSecretPublickeyCommand{
	cobraCommand: &cobra.Command{
		Use:   "publickey",
		Short: "Create SSH publickey",
		Long:  `Create SSH publickey.`,
	},
}

func init() {
	cmd := createSecretPublickeyCommand.cobraCommand

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := createSecretPublickeyCommand.Run(args)
		if err != nil {
			exitWithError(err)
		}
	}

	cmd.Flags().StringVarP(&createSecretPublickeyCommand.Pubkey, "pubkey", "i", "", "Path to SSH public key")

	createSecretCmd.cobraCommand.AddCommand(cmd)
}

func (cmd *CreateSecretPublickeyCommand) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("syntax: NAME -i <PublickeyPath>")
	}
	if len(args) != 1 {
		return fmt.Errorf("syntax: NAME -i <PublickeyPath>")
	}
	name := args[0]

	if cmd.Pubkey == "" {
		return fmt.Errorf("pubkey path is required (use -i)")
	}

	caStore, err := rootCommand.KeyStore()
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(cmd.Pubkey)
	if err != nil {
		return fmt.Errorf("error reading SSH public key %v: %v", cmd.Pubkey, err)
	}

	err = caStore.AddSSHPublicKey(name, data)
	if err != nil {
		return fmt.Errorf("error adding SSH public key: %v", err)
	}

	return nil
}
