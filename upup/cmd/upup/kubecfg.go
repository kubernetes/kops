package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// kubecfgCmd represents the kubecfg command
var kubecfgCmd = &cobra.Command{
	Use:   "kubecfg",
	Short: "Manage kubecfg files",
	Long:  `Manage kubecfg files`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Usage: generate")
	},
}

func init() {
	rootCommand.AddCommand(kubecfgCmd)
}
