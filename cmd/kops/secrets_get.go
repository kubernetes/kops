package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get secrets",
		Long:  `Get secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			exitWithError(fmt.Errorf("The 'secrets get' command has been replaced by 'get secret'"))
		},
	}

	secretsCmd.AddCommand(cmd)
}
