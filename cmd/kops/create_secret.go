package main

import (
	"github.com/spf13/cobra"
)

type CreateSecretCommand struct {
	cobraCommand *cobra.Command
}

var createSecretCmd = CreateSecretCommand{
	cobraCommand: &cobra.Command{
		Use:   "secret",
		Short: "Create secrets",
		Long:  `Create secrets.`,
	},
}

func init() {
	cmd := createSecretCmd.cobraCommand

	createCmd.AddCommand(cmd)
}
