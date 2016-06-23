package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export clusters",
	Long:  `export clusters`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Usage: cluster")
	},
}

func init() {
	rootCommand.AddCommand(exportCmd)
}
