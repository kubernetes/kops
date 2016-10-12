package main

import (
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
)

func NewCmdCreateSecret(f *util.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Create secrets",
		Long:  `Create secrets.`,
	}

	// create subcommands
	cmd.AddCommand(NewCmdCreateSecretPublicKey(f, out))

	return cmd
}
