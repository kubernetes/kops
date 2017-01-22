package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type VersionCmd struct {
	cobraCommand *cobra.Command
}

var versionCmd = VersionCmd{
	cobraCommand: &cobra.Command{
		Use:   "version",
		Short: "Print the client version information.",
	},
}

func init() {
	cmd := versionCmd.cobraCommand
	rootCommand.cobraCommand.AddCommand(cmd)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := versionCmd.Run()
		if err != nil {
			exitWithError(err)
		}
	}
}

func (c *VersionCmd) Run() error {
	fmt.Printf("Version %s\n", BuildVersion)

	return nil
}
