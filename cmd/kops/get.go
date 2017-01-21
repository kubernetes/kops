package main

import (
	"github.com/spf13/cobra"
)

// GetCmd represents the get command
type GetCmd struct {
	output string

	cobraCommand *cobra.Command
}

var getCmd = GetCmd{
	cobraCommand: &cobra.Command{
		Use:        "get",
		SuggestFor: []string{"list"},
		Short:      "List all instances of a resource.",
		Long:       `list or get objects`,
	},
}

const (
	OutputYaml  = "yaml"
	OutputTable = "table"
)

func init() {
	cmd := getCmd.cobraCommand

	rootCommand.AddCommand(cmd)

	cmd.PersistentFlags().StringVarP(&getCmd.output, "output", "o", OutputTable, "output format.  One of: table, yaml")
}
