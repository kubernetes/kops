package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"os"
)

func init() {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe secrets",
		Long:  `Describe secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stdout, "The 'secrets describe' command has been replaced by 'describe secrets'")
		},
	}

	secretsCmd.AddCommand(cmd)
}
