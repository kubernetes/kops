package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/util/pkg/tables"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"os"
)

type GetAddonsCmd struct {
}

var getAddonsCmd GetAddonsCmd

func init() {
	cmd := &cobra.Command{
		Use:     "addons",
		Aliases: []string{"addon"},
		Short:   "get addons",
		Long:    `List or get addons.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := getAddonsCmd.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	getCmd.cobraCommand.AddCommand(cmd)
}

type addonInfo struct {
	Name      string
	Version   *channels.ChannelVersion
	Namespace *v1.Namespace
}

func (c *GetAddonsCmd) Run(args []string) error {
	k8sClient, err := rootCommand.KubernetesClient()
	if err != nil {
		return err
	}

	namespaces, err := k8sClient.Namespaces().List(k8sapi.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing namespaces: %v", err)
	}

	var info []*addonInfo

	for i := range namespaces.Items {
		ns := &namespaces.Items[i]
		addons := channels.FindAddons(ns)
		for name, version := range addons {
			i := &addonInfo{
				Name:      name,
				Version:   version,
				Namespace: ns,
			}
			info = append(info, i)
		}
	}

	if len(info) == 0 {
		fmt.Printf("\nNo managed addons found\n")
		return nil
	}

	{
		t := &tables.Table{}
		t.AddColumn("NAME", func(r *addonInfo) string {
			return r.Name
		})
		t.AddColumn("NAMESPACE", func(r *addonInfo) string {
			return r.Namespace.Name
		})
		t.AddColumn("VERSION", func(r *addonInfo) string {
			if r.Version == nil {
				return "-"
			}
			if r.Version.Version != nil {
				return *r.Version.Version
			}
			return "?"
		})

		columns := []string{"NAMESPACE", "NAME", "VERSION"}
		err := t.Render(info, os.Stdout, columns...)
		if err != nil {
			return err
		}
	}

	fmt.Printf("\n")

	return nil
}
