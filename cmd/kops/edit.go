package main

import (
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
)

func NewCmdEdit(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "edit items",
	}

	// create subcommands
	cmd.AddCommand(NewCmdEditCluster(f, out))
	cmd.AddCommand(NewCmdEditInstanceGroup(f, out))
	cmd.AddCommand(NewCmdEditFederation(f, out))

	return cmd
}
