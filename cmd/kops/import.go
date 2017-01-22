package main

import (
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import existing resources into the state store.",
	Long:  `Import existing resources, such as clusters into the state store.`,
}

func init() {
	rootCommand.AddCommand(importCmd)
}
