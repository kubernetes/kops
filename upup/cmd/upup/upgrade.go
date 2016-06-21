package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade clusters",
	Long:  `upgrade clusters`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Usage: cluster")
	},
}

func init() {
	rootCommand.AddCommand(upgradeCmd)
}
