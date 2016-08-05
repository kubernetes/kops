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
			fmt.Printf("the 'secrets get' command has been replaced by 'get secrets -oplaintext'\n")
		},
	}

	secretsCmd.AddCommand(cmd)
}
