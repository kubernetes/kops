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
		Short: "rolling update clusters",
		Long:  `rolling update clusters`,
	},
}

func init() {
	rootCommand.AddCommand(rollingUpdateCommand.cobraCommand)
}
