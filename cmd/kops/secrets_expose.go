package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "expose",
		Short: "Expose secrets",
		Long:  `Expose secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			exitWithError(fmt.Errorf("The 'secrets export' command has been replaced by 'get secrets -oplaintext'"))
		},
	}

	secretsCmd.AddCommand(cmd)
}
