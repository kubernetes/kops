package main

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a resource by filename or stdin.",
	Long:  `Create resources`,
}

func init() {
	rootCommand.AddCommand(createCmd)
}
