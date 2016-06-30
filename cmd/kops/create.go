package main

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create clusters",
	Long:  `Create clusters`,
}

func init() {
	rootCommand.AddCommand(createCmd)
}
