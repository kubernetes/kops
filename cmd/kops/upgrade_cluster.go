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

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	kopsutil "k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/pretty"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	upgradeClusterLong = pretty.LongDesc(i18n.T(`
	Automates checking for and applying Kubernetes updates. This upgrades a cluster to the latest recommended
	production ready Kubernetes version. After this command is run, use ` + pretty.Bash("kops update cluster") + ` and ` + pretty.Bash("kops rolling-update cluster") + `
	to finish a cluster upgrade.
	`))

	upgradeClusterExample = templates.Examples(i18n.T(`
	# Upgrade a cluster's Kubernetes version.
	kops upgrade cluster k8s-cluster.example.com --yes --state=s3://my-state-store
	`))

	upgradeClusterShort = i18n.T("Upgrade a kubernetes cluster.")
)

type UpgradeClusterOptions struct {
	ClusterName string
	Yes         bool
	Channel     string
}

func NewCmdUpgradeCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &UpgradeClusterOptions{}

	cmd := &cobra.Command{
		Use:               "cluster [CLUSTER]",
		Short:             upgradeClusterShort,
		Long:              upgradeClusterLong,
		Example:           upgradeClusterExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()

			return RunUpgradeCluster(ctx, f, out, options)
		},
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", false, "Apply update")
	cmd.Flags().StringVar(&options.Channel, "channel", "", "Channel to use for upgrade")
	cmd.RegisterFlagCompletionFunc("channel", completeChannel)

	return cmd
}

type upgradeAction struct {
	Item     string
	Property string
	Old      string
	New      string

	apply func()
}

func RunUpgradeCluster(ctx context.Context, f *util.Factory, out io.Writer, options *UpgradeClusterOptions) error {
	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	instanceGroups, err := commands.ReadAllInstanceGroups(ctx, clientset, cluster)
	if err != nil {
		return err
	}

	if cluster.ObjectMeta.Annotations[kopsapi.AnnotationNameManagement] == kopsapi.AnnotationValueManagementImported {
		return fmt.Errorf("upgrade is not for use with imported clusters.)")
	}

	channelLocation := options.Channel
	if channelLocation == "" {
		channelLocation = cluster.Spec.Channel
	}
	if channelLocation == "" {
		channelLocation = kopsapi.DefaultChannel
	}

	var actions []*upgradeAction
	if channelLocation != cluster.Spec.Channel {
		actions = append(actions, &upgradeAction{
			Item:     "Cluster",
			Property: "Channel",
			Old:      cluster.Spec.Channel,
			New:      channelLocation,
			apply: func() {
				cluster.Spec.Channel = channelLocation
			},
		})
	}

	channel, err := kopsapi.LoadChannel(channelLocation)
	if err != nil {
		return fmt.Errorf("error loading channel %q: %v", channelLocation, err)
	}

	channelClusterSpec := channel.Spec.Cluster
	if channelClusterSpec == nil {
		// Just to prevent too much nil handling
		channelClusterSpec = &kopsapi.ClusterSpec{}
	}

	var currentKubernetesVersion *semver.Version
	{
		sv, err := kopsutil.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
		if err != nil {
			klog.Warningf("error parsing KubernetesVersion %q", cluster.Spec.KubernetesVersion)
		} else {
			currentKubernetesVersion = sv
		}
	}

	proposedKubernetesVersion := kopsapi.RecommendedKubernetesVersion(channel, kops.Version)

	// We won't propose a downgrade
	// TODO: What if a kubernetes version is bad?
	if currentKubernetesVersion != nil && proposedKubernetesVersion != nil && currentKubernetesVersion.GT(*proposedKubernetesVersion) {
		klog.Warningf("cluster version %q is greater than recommended version %q", *currentKubernetesVersion, *proposedKubernetesVersion)
		proposedKubernetesVersion = currentKubernetesVersion
	}

	if proposedKubernetesVersion != nil && currentKubernetesVersion != nil && currentKubernetesVersion.NE(*proposedKubernetesVersion) {
		actions = append(actions, &upgradeAction{
			Item:     "Cluster",
			Property: "KubernetesVersion",
			Old:      cluster.Spec.KubernetesVersion,
			New:      proposedKubernetesVersion.String(),
			apply: func() {
				cluster.Spec.KubernetesVersion = proposedKubernetesVersion.String()
			},
		})
	}

	// For further calculations, default to the current kubernetes version
	if proposedKubernetesVersion == nil {
		proposedKubernetesVersion = currentKubernetesVersion
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	// Prompt to upgrade image
	if proposedKubernetesVersion != nil {
		for _, ig := range instanceGroups {
			// Before kops v1.23, Spotinst uses the "ig.Spec.MachineType" field as a comma separated list to determine (among other) the allowed spot types.
			// A list is allowed only if all the items are with the same arch.
			// Therefore, it should be enough to check the arch for the first (possibly only) item
			machineType := strings.Split(ig.Spec.MachineType, ",")[0]

			architecture, err := cloudup.MachineArchitecture(cloud, machineType)
			if err != nil {
				klog.Warningf("Error finding architecture for machine type %q: %v", machineType, err)
				continue
			}
			image := channel.FindImage(cloud.ProviderID(), *proposedKubernetesVersion, architecture)
			if image == nil {
				klog.Warningf("No matching images specified in channel; cannot prompt for upgrade")
				continue
			}
			if channel.HasUpstreamImagePrefix(ig.Spec.Image) {
				if ig.Spec.Image != image.Name {
					target := ig
					actions = append(actions, &upgradeAction{
						Item:     "InstanceGroup/" + target.ObjectMeta.Name,
						Property: "Image",
						Old:      target.Spec.Image,
						New:      image.Name,
						apply: func() {
							target.Spec.Image = image.Name
						},
					})
				}
			} else {
				klog.Infof("Custom image (%s) has been provided for Instance Group %q; not updating image", ig.Spec.Image, ig.GetName())
			}
		}
	}

	// Prompt to upgrade to overlayfs
	if channelClusterSpec.Docker != nil {
		if cluster.Spec.Docker == nil {
			cluster.Spec.Docker = &kopsapi.DockerConfig{}
		}
		// TODO: make less hard-coded
		if channelClusterSpec.Docker.Storage != nil {
			dockerStorage := fi.StringValue(cluster.Spec.Docker.Storage)
			if dockerStorage != fi.StringValue(channelClusterSpec.Docker.Storage) {
				actions = append(actions, &upgradeAction{
					Item:     "Cluster",
					Property: "Docker.Storage",
					Old:      dockerStorage,
					New:      fi.StringValue(channelClusterSpec.Docker.Storage),
					apply: func() {
						cluster.Spec.Docker.Storage = channelClusterSpec.Docker.Storage
					},
				})
			}
		}
	}

	if len(actions) == 0 {
		// TODO: Allow --force option to force even if not needed?
		// Note stderr - we try not to print to stdout if no update is needed
		fmt.Fprintf(os.Stderr, "\nNo upgrade required\n")
		return nil
	}

	{
		t := &tables.Table{}
		t.AddColumn("ITEM", func(a *upgradeAction) string {
			return a.Item
		})
		t.AddColumn("PROPERTY", func(a *upgradeAction) string {
			return a.Property
		})
		t.AddColumn("OLD", func(a *upgradeAction) string {
			return a.Old
		})
		t.AddColumn("NEW", func(a *upgradeAction) string {
			return a.New
		})

		err := t.Render(actions, out, "ITEM", "PROPERTY", "OLD", "NEW")
		if err != nil {
			return err
		}
	}

	if !options.Yes {
		fmt.Printf("\nMust specify --yes to perform upgrade\n")
		return nil
	}
	for _, action := range actions {
		action.apply()
	}

	if err := commands.UpdateCluster(ctx, clientset, cluster, instanceGroups); err != nil {
		return err
	}

	for _, g := range instanceGroups {
		_, err := clientset.InstanceGroupsFor(cluster).Update(ctx, g, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("error writing InstanceGroup %q: %v", g.ObjectMeta.Name, err)
		}
	}

	fmt.Printf("\nUpdates applied to configuration.\n")

	// TODO: automate this step
	fmt.Printf("You can now apply these changes, using `kops update cluster %s`\n", cluster.ObjectMeta.Name)

	return nil
}

func completeChannel(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO implement completion against VFS
	return []string{"alpha", "stable"}, cobra.ShellCompDirectiveNoFileComp
}
