package main

import (
	"github.com/spf13/cobra"
)

// updateCmd represents the create command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Creates or updates cloud resources to match cluster spec.",
	Long:  `Update clusters`,
}

func init() {
	rootCommand.AddCommand(updateCmd)
}
