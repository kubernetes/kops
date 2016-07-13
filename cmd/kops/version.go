package main

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	// value overwritten during build. This can be used to resolve issues.
	BuildVersion = "0.1"
)

type VersionCmd struct {
	cobraCommand *cobra.Command
}

var versionCmd = VersionCmd{
	cobraCommand: &cobra.Command{
		Use:   "version",
		Short: "Print the client version information",
	},
}

func init() {
	cmd := versionCmd.cobraCommand
	rootCommand.cobraCommand.AddCommand(cmd)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := versionCmd.Run()
		if err != nil {
			glog.Exitf("%v", err)
		}
	}
}

func (c *VersionCmd) Run() error {
	fmt.Printf("Version %s\n", BuildVersion)

	return nil
}
