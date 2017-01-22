package main

import (
	"github.com/spf13/cobra"
)

// toolboxCmd represents the toolbox command
var toolboxCmd = &cobra.Command{
	Use:   "toolbox",
	Short: "Misc infrequently used commands",
}

func init() {
	rootCommand.AddCommand(toolboxCmd)
}
