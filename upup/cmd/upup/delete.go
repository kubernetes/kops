package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete clusters",
	Long:  `Delete clusters`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Usage: cluster")
	},
}

func init() {
	rootCommand.AddCommand(deleteCmd)
}
