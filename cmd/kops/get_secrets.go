package main

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/upup/pkg/fi"
)

type GetSecretsCommand struct {
}

var getSecretsCommand GetSecretsCommand

func init() {
	cmd := &cobra.Command{
		Use:     "secrets",
		Aliases: []string{"secret"},
		Short:   "get secrets",
		Long:    `List or get secrets.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getSecretsCommand.Run()
			if err != nil {
				glog.Exitf("%v", err)
			}
		},
	}

	getCmd.AddCommand(cmd)
}

func (c *GetSecretsCommand) Run() error {
	var items []*fi.KeystoreItem
	{
		caStore, err := rootCommand.KeyStore()
		if err != nil {
			return err
		}
		l, err := caStore.List()
		if err != nil {
			return fmt.Errorf("error listing CA store items %v", err)
		}

		for _, i := range l {
			items = append(items, i)
		}
	}

	{
		secretStore, err := rootCommand.SecretStore()
		if err != nil {
			return err
		}
		l, err := secretStore.ListSecrets()
		if err != nil {
			return fmt.Errorf("error listing secrets %v", err)
		}

		for _, id := range l {
			info := &fi.KeystoreItem{
				Name: id,
				Type: fi.SecretTypeSecret,
			}
			items = append(items, info)
		}
	}

	if len(items) == 0 {
		return nil
	}

	t := &Table{}
	t.AddColumn("NAME", func(i *fi.KeystoreItem) string {
		return i.Name
	})
	t.AddColumn("ID", func(i *fi.KeystoreItem) string {
		return i.Id
	})
	t.AddColumn("TYPE", func(i *fi.KeystoreItem) string {
		return i.Type
	})
	return t.Render(items, os.Stdout, "TYPE", "NAME", "ID")
}
