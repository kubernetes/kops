package main

import (
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
)

func NewCmdUpdate(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update clusters",
		Long:  `Update clusters`,
	}

	//  subcommands
	cmd.AddCommand(NewCmdUpdateCluster(f, out))
	cmd.AddCommand(NewCmdUpdateFederation(f, out))

	return cmd
}
