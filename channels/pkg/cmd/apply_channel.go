/*
Copyright 2019 The Kubernetes Authors.

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

package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/spf13/cobra"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/util/pkg/tables"
)

type ApplyChannelOptions struct {
	Yes   bool
	Files []string
}

func NewCmdApplyChannel(f Factory, out io.Writer) *cobra.Command {
	var options ApplyChannelOptions

	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Apply channel",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunApplyChannel(f, out, &options, args)
		},
	}

	cmd.Flags().BoolVar(&options.Yes, "yes", false, "Apply update")
	cmd.Flags().StringSliceVarP(&options.Files, "filename", "f", []string{}, "Apply from a local file")

	return cmd
}

func RunApplyChannel(f Factory, out io.Writer, options *ApplyChannelOptions, args []string) error {
	k8sClient, err := f.KubernetesClient()
	if err != nil {
		return err
	}

	kubernetesVersionInfo, err := k8sClient.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("error querying kubernetes version: %v", err)
	}

	//kubernetesVersion, err := semver.Parse(kubernetesVersionInfo.Major + "." + kubernetesVersionInfo.Minor + ".0")
	//if err != nil {
	//	return fmt.Errorf("cannot parse kubernetes version %q", kubernetesVersionInfo.Major+"."+kubernetesVersionInfo.Minor + ".0")
	//}

	kubernetesVersion, err := semver.ParseTolerant(kubernetesVersionInfo.GitVersion)
	if err != nil {
		return fmt.Errorf("cannot parse kubernetes version %q", kubernetesVersionInfo.GitVersion)
	}

	// Remove Pre and Patch, as they make semver comparisons impractical
	kubernetesVersion.Pre = nil

	menu := channels.NewAddonMenu()

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

		current, err := o.GetCurrent(kubernetesVersion)
		if err != nil {
			return fmt.Errorf("error processing latest versions in %q: %v", location, err)
		}
		menu.MergeAddons(current)
	}

	for _, f := range options.Files {
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

		current, err := o.GetCurrent(kubernetesVersion)
		if err != nil {
			return fmt.Errorf("error processing latest versions in %q: %v", f, err)
		}
		menu.MergeAddons(current)
	}

	var updates []*channels.AddonUpdate
	var needUpdates []*channels.Addon
	for _, addon := range menu.Addons {
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

	if !options.Yes {
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
				fmt.Printf("Updated %q to %s\n", update.Name, *update.NewVersion.Version)
			} else {
				fmt.Printf("Updated %q\n", update.Name)
			}
		}
	}

	fmt.Printf("\n")

	return nil
}
