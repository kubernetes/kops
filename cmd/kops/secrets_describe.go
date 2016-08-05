package main

import (
	"fmt"

	"github.com/spf13/cobra"
)


func init() {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe secrets",
		Long:  `Describe secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			exitWithError(fmt.Errorf("The 'secrets describe' command has been replaced by 'describe secrets'"))
		},
	}

	secretsCmd.AddCommand(cmd)
}
