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
	"os/signal"
	"syscall"
	"time"

	"github.com/blang/semver/v4"
	certmanager "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"

	"k8s.io/kops/channels/pkg/channels"
	"k8s.io/kops/channels/pkg/nodelabeler"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kops/util/pkg/vfs"
)

type ApplyChannelOptions struct {
	Yes      bool
	Interval time.Duration
	NodeName string
}

func NewCmdApplyChannel(f *ChannelsFactory, out io.Writer) *cobra.Command {
	var options ApplyChannelOptions

	cmd := &cobra.Command{
		Use:   "channel CHANNEL...",
		Short: "Applies updates from the given channel(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.Interval > 0 {
				ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
				defer cancel()
				return runApplyChannelLoop(ctx, out, &options, args)
			}
			return runApplyChannelIteration(context.TODO(), NewChannelsFactory(), out, &options, args)
		},
	}

	cmd.Flags().BoolVar(&options.Yes, "yes", false, "Apply update")
	cmd.Flags().DurationVar(&options.Interval, "interval", 0, "If non-zero, re-apply the channel on this interval until interrupted (e.g. 60s)")
	cmd.Flags().StringVar(&options.NodeName, "node-name", "", "If set, patch the named node with the mandatory control-plane labels each iteration; typically supplied via the downward API.")

	return cmd
}

// runApplyChannelIteration patches node labels (when --node-name is set) then
// applies the channel. Labels go first so addons targeting the control-plane
// label can schedule on the local node as soon as their manifests land.
func runApplyChannelIteration(ctx context.Context, f *ChannelsFactory, out io.Writer, options *ApplyChannelOptions, args []string) error {
	var merr error
	if options.NodeName != "" {
		labelerClient, err := buildKubernetesClient(f)
		if err != nil {
			merr = multierr.Append(merr, fmt.Errorf("building kubernetes client for node labeler: %w", err))
		} else if err := nodelabeler.BootstrapControlPlaneNodeLabels(ctx, labelerClient, options.NodeName); err != nil {
			merr = multierr.Append(merr, fmt.Errorf("bootstrapping node labels: %w", err))
		}
	}
	if err := RunApplyChannel(ctx, f, out, options, args); err != nil {
		merr = multierr.Append(merr, err)
	}
	return merr
}

// runApplyChannelLoop runs iterations on a fixed interval until ctx is cancelled.
// A fresh ChannelsFactory per iteration drops cached REST configs, HTTP clients,
// and the discovery cache — picks up cert rotation and CRD additions without restart.
func runApplyChannelLoop(ctx context.Context, out io.Writer, options *ApplyChannelOptions, args []string) error {
	ticker := time.NewTicker(options.Interval)
	defer ticker.Stop()
	for {
		if err := runApplyChannelIteration(ctx, NewChannelsFactory(), out, options, args); err != nil {
			klog.Warningf("error in apply iteration (will retry in %s): %v", options.Interval, err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func buildKubernetesClient(f *ChannelsFactory) (kubernetes.Interface, error) {
	restConfig, err := f.RESTConfig()
	if err != nil {
		return nil, err
	}
	httpClient, err := f.HTTPClient()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfigAndClient(restConfig, httpClient)
}

func RunApplyChannel(ctx context.Context, f *ChannelsFactory, out io.Writer, options *ApplyChannelOptions, args []string) error {
	restConfig, err := f.RESTConfig()
	if err != nil {
		return err
	}
	httpClient, err := f.HTTPClient()
	if err != nil {
		return err
	}

	k8sClient, err := kubernetes.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return fmt.Errorf("building kube client: %w", err)
	}

	cmClient, err := certmanager.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return fmt.Errorf("building cert manager client: %w", err)
	}

	dynamicClient, err := f.DynamicClient()
	if err != nil {
		return fmt.Errorf("building dynamic client: %w", err)
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

	if len(args) == 0 {
		return fmt.Errorf("at least one channel URL is required")
	}

	var merr error
	for _, channelLocation := range args {
		menu, err := buildMenu(f.VFSContext(), kubernetesVersion, channelLocation)
		if err != nil {
			merr = multierr.Append(merr, fmt.Errorf("building menu for %q: %w", channelLocation, err))
			continue
		}
		if err := applyMenu(ctx, menu, f.VFSContext(), k8sClient, cmClient, dynamicClient, restMapper, options.Yes); err != nil {
			merr = multierr.Append(merr, fmt.Errorf("applying %q: %w", channelLocation, err))
		}
	}
	return merr
}

func applyMenu(ctx context.Context, menu *channels.AddonMenu, vfsContext *vfs.VFSContext, k8sClient kubernetes.Interface, cmClient certmanager.Interface, dynamicClient dynamic.Interface, restMapper *restmapper.DeferredDiscoveryRESTMapper, apply bool) error {
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

	applier := &channels.ClientApplier{
		Client:     dynamicClient,
		RESTMapper: restMapper,
	}

	var merr error

	for _, needUpdate := range needUpdates {
		update, err := needUpdate.EnsureUpdated(ctx, vfsContext, k8sClient, cmClient, pruner, applier, channelVersions[needUpdate.GetNamespace()+":"+needUpdate.Name])
		if err != nil {
			merr = multierr.Append(merr, fmt.Errorf("updating %q: %w", needUpdate.Name, err))
		} else if update != nil {
			fmt.Printf("Updated %q\n", update.Name)
		}
	}

	return merr
}

func getUpdates(ctx context.Context, menu *channels.AddonMenu, k8sClient kubernetes.Interface, cmClient certmanager.Interface, channelVersions map[string]*channels.ChannelVersion) ([]*channels.AddonUpdate, []*channels.Addon, error) {
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

func buildMenu(vfsContext *vfs.VFSContext, kubernetesVersion semver.Version, channelLocation string) (*channels.AddonMenu, error) {
	menu := channels.NewAddonMenu()

	location, err := url.Parse(channelLocation)
	if err != nil {
		return nil, fmt.Errorf("unable to parse argument %q as url", channelLocation)
	}
	if !location.IsAbs() {
		expanded := "https://raw.githubusercontent.com/kubernetes/kops/master/addons/" + channelLocation + "/addon.yaml"
		// Disallow the use of legacy addons from the "well-known" location starting Kubernetes 1.23:
		// https://raw.githubusercontent.com/kubernetes/kops/master/addons/<name>/addon.yaml
		return nil, fmt.Errorf("legacy addons are deprecated and unmaintained, use managed addons instead of %s", expanded)
	}
	o, err := channels.LoadAddons(vfsContext, channelLocation, location)
	if err != nil {
		return nil, fmt.Errorf("error loading channel %q: %v", location, err)
	}

	current, err := o.GetCurrent(kubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("error processing latest versions in %q: %v", location, err)
	}
	menu.MergeAddons(current)
	return menu, nil
}
