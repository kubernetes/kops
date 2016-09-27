package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/util/pkg/tables"
	"net/url"
	"os"
)

type ApplyChannelCmd struct {
	Yes bool
}

var applyChannel ApplyChannelCmd

func init() {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Apply channel",
		Run: func(cmd *cobra.Command, args []string) {
			err := applyChannel.Run(args)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().BoolVar(&applyChannel.Yes, "yes", false, "Apply update")

	applyCmd.AddCommand(cmd)
}

func (c *ApplyChannelCmd) Run(args []string) error {
	k8sClient, err := rootCommand.KubernetesClient()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}
	baseURL, err := url.Parse(cwd + string(os.PathSeparator))
	if err != nil {
		return fmt.Errorf("error building url for current directory %q: %v", cwd, err)
	}

	var addons []*channels.Addon
	for _, arg := range args {
		channel, err := url.Parse(arg)
		if err != nil {
			return fmt.Errorf("unable to parse argument %q as url", arg)
		}
		if !channel.IsAbs() {
			channel = baseURL.ResolveReference(channel)
		}
		o, err := channels.LoadAddons(channel)
		if err != nil {
			return fmt.Errorf("error loading file %q: %v", arg, err)
		}

		current, err := o.GetCurrent()
		if err != nil {
			return fmt.Errorf("error processing latest versions in %q: %v", arg, err)
		}
		addons = append(addons, current...)
	}

	var updates []*channels.AddonUpdate
	var needUpdates []*channels.Addon
	for _, addon := range addons {
		// TODO: Cache lookups to prevent repeated lookups?
		update, err := addon.GetRequiredUpdates(k8sClient)
		if err != nil {
			return fmt.Errorf("error checking for required update: %v", err)
		}
		if update != nil {
			updates = append(updates, update)
			needUpdates = append(needUpdates, addon)
		}
	}

	if len(updates) == 0 {
		fmt.Printf("No update required\n")
		return nil
	}

	{
		t := &tables.Table{}
		t.AddColumn("NAME", func(r *channels.AddonUpdate) string {
			return r.Name
		})
		t.AddColumn("CURRENT", func(r *channels.AddonUpdate) string {
			if r.ExistingVersion == nil {
				return "-"
			}
			if r.ExistingVersion.Version != nil {
				return *r.ExistingVersion.Version
			}
			return "?"
		})
		t.AddColumn("UPDATE", func(r *channels.AddonUpdate) string {
			if r.NewVersion == nil {
				return "-"
			}
			if r.NewVersion.Version != nil {
				return *r.NewVersion.Version
			}
			return "?"
		})

		columns := []string{"NAME", "CURRENT", "UPDATE"}
		err := t.Render(updates, os.Stdout, columns...)
		if err != nil {
			return err
		}
	}

	if !c.Yes {
		fmt.Printf("\nMust specify --yes to update\n")
		return nil
	}

	for _, needUpdate := range needUpdates {
		update, err := needUpdate.EnsureUpdated(k8sClient)
		if err != nil {
			return fmt.Errorf("error updating %q: %v", needUpdate.Name, err)
		}
		if update.NewVersion.Version != nil {
			fmt.Printf("Updated %q to %d\n", update.Name, *update.NewVersion)
		} else {
			fmt.Printf("Updated %q\n", update.Name)
		}
	}

	fmt.Printf("\n")

	return nil
}
