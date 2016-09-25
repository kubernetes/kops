package main

import (
	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply resources from a channel",
}

func init() {
	rootCommand.AddCommand(applyCmd)
}
