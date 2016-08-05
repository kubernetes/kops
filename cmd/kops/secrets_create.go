package main

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
)


func init() {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create secrets",
		Long:  `Create secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			exitWithError(fmt.Errorf("The 'secrets create' command has been replaced by 'create secrets'"))
		},
	}

	secretsCmd.AddCommand(cmd)
}
