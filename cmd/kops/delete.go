package main

import (
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete clusters",
	Long:  `Delete clusters`,
}

func init() {
	rootCommand.AddCommand(deleteCmd)
}
