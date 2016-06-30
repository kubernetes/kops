package main

import (
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit clusters",
	Long:  `edit clusters`,
}

func init() {
	rootCommand.AddCommand(editCmd)
}
