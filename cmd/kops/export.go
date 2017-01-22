package main

import (
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports a kubecfg for target cluster.",
	Long:  `export clusters/kubecfg`,
}

func init() {
	rootCommand.AddCommand(exportCmd)
}
