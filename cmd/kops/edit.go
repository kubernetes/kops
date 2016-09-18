package main

import (
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit items",
}

func init() {
	rootCommand.AddCommand(editCmd)
}
