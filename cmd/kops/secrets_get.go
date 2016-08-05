package main

import (
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get secrets",
		Long:  `Get secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getSecretsCommand.Run(args)
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	secretsCmd.AddCommand(cmd)
}
