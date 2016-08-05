package main

import (
	"github.com/spf13/cobra"
)

// DescribeCmd represents the describe command
type DescribeCmd struct {
	cobraCommand *cobra.Command
}

var describeCmd = DescribeCmd{
	cobraCommand: &cobra.Command{
		Use:   "describe",
		Short: "describe objects",
	},
}

func init() {
	cmd := describeCmd.cobraCommand

	rootCommand.AddCommand(cmd)
}
