package main

import (
	"fmt"

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
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Usage: cluster")
		},
	},
}

func init() {
	rootCommand.AddCommand(rollingUpdateCommand.cobraCommand)
}
