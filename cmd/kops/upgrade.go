package main

import (
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade clusters",
	Long:  `upgrade clusters`,
}

func init() {
	rootCommand.AddCommand(upgradeCmd)
}
