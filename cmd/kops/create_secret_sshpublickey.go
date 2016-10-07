package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/kops/upup/pkg/api/registry"
)

type CreateSecretPublickeyCommand struct {
	cobraCommand *cobra.Command

	Pubkey string
}

var createSecretPublickeyCommand = CreateSecretPublickeyCommand{
	cobraCommand: &cobra.Command{
		Use:   "sshpublickey",
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
		return fmt.Errorf("syntax: NAME -i <PublicKeyPath>")
	}
	if len(args) != 1 {
		return fmt.Errorf("syntax: NAME -i <PublicKeyPath>")
	}
	name := args[0]

	if cmd.Pubkey == "" {
		return fmt.Errorf("pubkey path is required (use -i)")
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(cmd.Pubkey)
	if err != nil {
		return fmt.Errorf("error reading SSH public key %v: %v", cmd.Pubkey, err)
	}

	err = keyStore.AddSSHPublicKey(name, data)
	if err != nil {
		return fmt.Errorf("error adding SSH public key: %v", err)
	}

	return nil
}
