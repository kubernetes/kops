package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops/registry"
	"os"
)

type CreateSecretPublickeyOptions struct {
	Pubkey string
}

func NewCmdCreateSecretPublicKey(f *util.Factory, out io.Writer) *cobra.Command {
	options := &CreateSecretPublickeyOptions{}

	cmd := &cobra.Command{
		Use:   "sshpublickey",
		Short: "Create SSH publickey",
		Long:  `Create SSH publickey.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCreateSecretPublicKey(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVarP(&options.Pubkey, "pubkey", "i", "", "Path to SSH public key")

	return cmd
}

func RunCreateSecretPublicKey(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *CreateSecretPublickeyOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("syntax: NAME -i <PublicKeyPath>")
	}
	if len(args) != 1 {
		return fmt.Errorf("syntax: NAME -i <PublicKeyPath>")
	}
	name := args[0]

	if options.Pubkey == "" {
		return fmt.Errorf("pubkey path is required (use -i)")
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	keyStore, err := registry.KeyStore(cluster)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(options.Pubkey)
	if err != nil {
		return fmt.Errorf("error reading SSH public key %v: %v", options.Pubkey, err)
	}

	err = keyStore.AddSSHPublicKey(name, data)
	if err != nil {
		return fmt.Errorf("error adding SSH public key: %v", err)
	}

	return nil
}
