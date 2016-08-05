package main

import (
	"fmt"

	"github.com/spf13/cobra"
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
			fmt.Printf("the 'secrets expose' command has been replaced by 'get secrets -oplaintext'\n")
		},
	}

	secretsCmd.AddCommand(cmd)

	cmd.Flags().StringVarP(&exposeSecretsCommand.Type, "type", "", "", "Type of secret to create")
	cmd.Flags().StringVarP(&exposeSecretsCommand.ID, "id", "", "", "Id of secret to create")
}
