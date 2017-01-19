/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/util/pkg/tables"
	"net/url"
	"os"
	"strings"
)

type ApplyChannelCmd struct {
	Yes   bool
	Files []string
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
	cmd.Flags().StringSliceVar(&applyChannel.Files, "f", []string{}, "Apply from a local file")

	applyCmd.AddCommand(cmd)
}

func (c *ApplyChannelCmd) Run(args []string) error {
	k8sClient, err := rootCommand.KubernetesClient()
	if err != nil {
		return err
	}

	var addons []*channels.Addon
	for _, name := range args {
		location, err := url.Parse(name)
		if err != nil {
			return fmt.Errorf("unable to parse argument %q as url", name)
		}
		if !location.IsAbs() {
			// We recognize the following "well-known" format:
			// <name> with no slashes ->
			if strings.Contains(name, "/") {
				return fmt.Errorf("Channel format not recognized (did you mean to use `-f` to specify a local file?): %q", name)
			}
			expanded := "https://raw.githubusercontent.com/kubernetes/kops/master/addons/" + name + "/addon.yaml"
			location, err = url.Parse(expanded)
			if err != nil {
				return fmt.Errorf("unable to parse expanded argument %q as url", expanded)
			}
		}
		o, err := channels.LoadAddons(name, location)
		if err != nil {
			return fmt.Errorf("error loading channel %q: %v", location, err)
		}

		current, err := o.GetCurrent()
		if err != nil {
			return fmt.Errorf("error processing latest versions in %q: %v", location, err)
		}
		addons = append(addons, current...)
	}

	for _, f := range c.Files {
		location, err := url.Parse(f)
		if err != nil {
			return fmt.Errorf("unable to parse argument %q as url", f)
		}
		if !location.IsAbs() {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("error getting current directory: %v", err)
			}
			baseURL, err := url.Parse(cwd + string(os.PathSeparator))
			if err != nil {
				return fmt.Errorf("error building url for current directory %q: %v", cwd, err)
			}
			location = baseURL.ResolveReference(location)
		}
		o, err := channels.LoadAddons(f, location)
		if err != nil {
			return fmt.Errorf("error loading file %q: %v", f, err)
		}

		current, err := o.GetCurrent()
		if err != nil {
			return fmt.Errorf("error processing latest versions in %q: %v", f, err)
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
		// Could have been a concurrent request
		if update != nil {
			if update.NewVersion.Version != nil {
				fmt.Printf("Updated %q to %d\n", update.Name, *update.NewVersion)
			} else {
				fmt.Printf("Updated %q\n", update.Name)
			}
		}
	}

	fmt.Printf("\n")

	return nil
}
