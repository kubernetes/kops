package main

import (
	"github.com/spf13/cobra"
)

// rollingupdateCmd represents the rollingupdate command
type RollingUpdateCmd struct {
	cobraCommand *cobra.Command
}

var rollingUpdateCommand = RollingUpdateCmd{
	cobraCommand: &cobra.Command{
		Use:   "rolling-update",
		Short: "Initiate rolling updates on clusters.",
		Long:  `rolling update clusters`,
	},
}

func init() {
	rootCommand.AddCommand(rollingUpdateCommand.cobraCommand)
}
