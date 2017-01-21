package main

import (
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete clusters and other resources.",
	Long:  `Delete resources, such as clusters, instance groups, and secrets.'`,
}

func init() {
	rootCommand.AddCommand(deleteCmd)
}
