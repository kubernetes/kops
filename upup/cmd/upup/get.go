package main

import (
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "list or get obejcts",
	Long:  `list or get obejcts`,
}

func init() {
	rootCommand.AddCommand(getCmd)
}
