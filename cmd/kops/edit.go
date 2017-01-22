package main

import (
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit clusters and other resrouces.",
}

func init() {
	rootCommand.AddCommand(editCmd)
}
