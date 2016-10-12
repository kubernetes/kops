package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/federation"
	"os"
)

type UpdateFederationOptions struct {
}

func NewCmdUpdateFederation(f *util.Factory, out io.Writer) *cobra.Command {
	options := &UpdateFederationOptions{}

	cmd := &cobra.Command{
		Use:   "federation",
		Short: "Update federation",
		Long:  `Updates a k8s federation.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunUpdateFederation(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunUpdateFederation(factory *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *UpdateFederationOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("Specify name of Federation to update")
	}
	if len(args) != 1 {
		return fmt.Errorf("Can only update one Federation at a time")
	}

	name := args[0]

	clientset, err := factory.Clientset()
	if err != nil {
		return err
	}

	f, err := clientset.Federations().Get(name)
	if err != nil {
		return fmt.Errorf("error reading federation %q: %v", name, err)
	}

	applyCmd := &federation.ApplyFederationOperation{
		Federation: f,
		KopsClient: clientset,
	}
	err = applyCmd.Run()
	if err != nil {
		return err
	}

	kubecfg, err := applyCmd.FindKubecfg()
	if err != nil {
		return err
	}

	if kubecfg == nil {
		return fmt.Errorf("cannot find configuration for federation")
	}

	err = kubecfg.WriteKubecfg()
	if err != nil {
		return err
	}

	return nil
}
