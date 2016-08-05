package main

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create resources",
	Long:  `Create resources`,
}

func init() {
	rootCommand.AddCommand(createCmd)
}
