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
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/pkg/apis/kops/util"
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
			ctx := context.TODO()
			return RunApplyChannel(ctx, f, out, &options, args)
		},
	}

	cmd.Flags().BoolVar(&options.Yes, "yes", false, "Apply update")
	cmd.Flags().StringSliceVarP(&options.Files, "filename", "f", []string{}, "Apply from a local file")

	return cmd
}

func RunApplyChannel(ctx context.Context, f Factory, out io.Writer, options *ApplyChannelOptions, args []string) error {
	k8sClient, err := f.KubernetesClient()
	if err != nil {
		return err
	}

	cmClient, err := f.CertManagerClient()
	if err != nil {
		return err
	}

	dynamicClient, err := f.DynamicClient()
	if err != nil {
		return err
	}

	restMapper, err := f.RESTMapper()
	if err != nil {
		return err
	}

	kubernetesVersionInfo, err := k8sClient.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("error querying kubernetes version: %v", err)
	}

	kubernetesVersion, err := semver.ParseTolerant(kubernetesVersionInfo.GitVersion)
	if err != nil {
		return fmt.Errorf("cannot parse kubernetes version %q", kubernetesVersionInfo.GitVersion)
	}

	// Remove Pre and Patch, as they make semver comparisons impractical
	kubernetesVersion.Pre = nil

	// menu is the expected list of addons in the cluster and their configurations.
	menu, err := buildMenu(kubernetesVersion, args, false)
	if err != nil {
		return fmt.Errorf("cannot build the addon menu from args: %w", err)
	}

	filesMenu, err := buildMenu(kubernetesVersion, options.Files, true)
	if err != nil {
		return fmt.Errorf("cannot build the addon menu from files: %w", err)
	}
	menu.MergeAddons(filesMenu)

	return applyMenu(ctx, menu, k8sClient, cmClient, dynamicClient, restMapper, options.Yes)
}

func applyMenu(ctx context.Context, menu *channels.AddonMenu, k8sClient kubernetes.Interface, cmClient versioned.Interface, dynamicClient dynamic.Interface, restMapper *restmapper.DeferredDiscoveryRESTMapper, apply bool) error {
	// channelVersions is the list of installed addons in the cluster.
	// It is keyed by <namespace>:<addon name>.
	channelVersions, err := getChannelVersions(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("cannot fetch channel versions from namespaces: %w", err)
	}

	updates, needUpdates, err := getUpdates(ctx, menu, k8sClient, cmClient, channelVersions)
	if err != nil {
		return fmt.Errorf("failed to get updates: %w", err)
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
			return r.ExistingVersion.ManifestHash
		})
		t.AddColumn("UPDATE", func(r *channels.AddonUpdate) string {
			if r.NewVersion == nil {
				return "-"
			}
			return r.NewVersion.ManifestHash
		})
		t.AddColumn("PKI", func(r *channels.AddonUpdate) string {
			if r.InstallPKI {
				return "yes"
			}
			return "no"
		})

		columns := []string{"NAME", "CURRENT", "UPDATE", "PKI"}
		err := t.Render(updates, os.Stdout, columns...)
		if err != nil {
			return err
		}
	}

	if !apply {
		fmt.Printf("\nMust specify --yes to update\n")
		return nil
	}

	pruner := &channels.Pruner{
		Client:     dynamicClient,
		RESTMapper: restMapper,
	}

	for _, needUpdate := range needUpdates {
		update, err := needUpdate.EnsureUpdated(ctx, k8sClient, cmClient, pruner, channelVersions[needUpdate.GetNamespace()+":"+needUpdate.Name])
		if err != nil {
			fmt.Printf("error updating %q: %v", needUpdate.Name, err)
		} else if update != nil {
			fmt.Printf("Updated %q\n", update.Name)
		}
	}

	fmt.Printf("\n")

	return nil
}

func getUpdates(ctx context.Context, menu *channels.AddonMenu, k8sClient kubernetes.Interface, cmClient versioned.Interface, channelVersions map[string]*channels.ChannelVersion) ([]*channels.AddonUpdate, []*channels.Addon, error) {
	var updates []*channels.AddonUpdate
	var needUpdates []*channels.Addon
	for _, addon := range menu.Addons {
		update, err := addon.GetRequiredUpdates(ctx, k8sClient, cmClient, channelVersions[addon.GetNamespace()+":"+addon.Name])
		if err != nil {
			return nil, nil, fmt.Errorf("error checking for required update: %v", err)
		}
		if update != nil {
			updates = append(updates, update)
			needUpdates = append(needUpdates, addon)
		}
	}
	return updates, needUpdates, nil
}

func getChannelVersions(ctx context.Context, k8sClient kubernetes.Interface) (map[string]*channels.ChannelVersion, error) {
	namespaces, err := k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing namespaces: %v", err)
	}

	channelVersions := make(map[string]*channels.ChannelVersion)
	for i := range namespaces.Items {
		ns := &namespaces.Items[i]
		addons := channels.FindChannelVersions(ns)
		for name, version := range addons {
			channelVersions[ns.Name+":"+name] = version
		}
	}
	return channelVersions, nil
}

func buildMenu(kubernetesVersion semver.Version, args []string, localFiles bool) (*channels.AddonMenu, error) {
	menu := channels.NewAddonMenu()

	for _, name := range args {
		location, err := url.Parse(name)
		if err != nil {
			return nil, fmt.Errorf("unable to parse argument %q as url", name)
		}
		if !location.IsAbs() {
			if !localFiles {
				// We recognize the following "well-known" format:
				// <name> with no slashes ->
				if strings.Contains(name, "/") {
					return nil, fmt.Errorf("channel format not recognized (did you mean to use `-f` to specify a local file?): %q", name)
				}
				expanded := "https://raw.githubusercontent.com/kubernetes/kops/master/addons/" + name + "/addon.yaml"
				location, err = url.Parse(expanded)
				if err != nil {
					return nil, fmt.Errorf("unable to parse expanded argument %q as url", expanded)
				}
				// Disallow the use of legacy addons from the "well-known" location starting Kubernetes 1.23:
				// https://raw.githubusercontent.com/kubernetes/kops/master/addons/<name>/addon.yaml
				if util.IsKubernetesGTE("1.23", kubernetesVersion) {
					return nil, fmt.Errorf("legacy addons are deprecated and unmaintained, use managed addons instead of %s", expanded)
				} else {
					klog.Warningf("Legacy addons are deprecated and unmaintained, use managed addons instead of %s", expanded)
				}
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return nil, fmt.Errorf("error getting current directory: %v", err)
				}
				baseURL, err := url.Parse(cwd + string(os.PathSeparator))
				if err != nil {
					return nil, fmt.Errorf("error building url for current directory %q: %v", cwd, err)
				}
				location = baseURL.ResolveReference(location)
			}
		}
		o, err := channels.LoadAddons(name, location)
		if err != nil {
			return nil, fmt.Errorf("error loading channel %q: %v", location, err)
		}

		current, err := o.GetCurrent(kubernetesVersion)
		if err != nil {
			return nil, fmt.Errorf("error processing latest versions in %q: %v", location, err)
		}
		menu.MergeAddons(current)
	}
	return menu, nil
}
