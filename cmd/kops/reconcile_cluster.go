/*
Copyright 2024 The Kubernetes Authors.

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

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	reconcileClusterLong = templates.LongDesc(i18n.T(`
	Reconcile the cluster by updating and rolling the control plane and nodes sequentially.
	`))

	reconcileClusterExample = templates.Examples(i18n.T(`
	# After the cluster has been edited or upgraded, update the cloud resources with:
	kops reconcile cluster k8s-cluster.example.com --state=s3://my-state-store --yes
	`))

	reconcileClusterShort = i18n.T("Reconcile a cluster.")
)

type ReconcileClusterOptions struct {
	CoreUpdateClusterOptions
}

func NewCmdReconcileCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ReconcileClusterOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:               "cluster [CLUSTER]",
		Short:             reconcileClusterShort,
		Long:              reconcileClusterLong,
		Example:           reconcileClusterExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := RunReconcileCluster(cmd.Context(), f, out, &options.CoreUpdateClusterOptions)
			return err
		},
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Create cloud resources, without --yes reconcile is in dry run mode")

	// These flags from the update command are not obviously needed by reconcile, though we can add them if needed:
	//
	// cmd.Flags().StringVar(&options.Target, "target", options.Target, "Target - direct")
	// cmd.RegisterFlagCompletionFunc("target", completeUpdateClusterTarget(f, &options.CoreUpdateClusterOptions))
	// cmd.Flags().StringVar(&options.SSHPublicKey, "ssh-public-key", options.SSHPublicKey, "SSH public key to use (deprecated: use kops create secret instead)")
	// cmd.Flags().StringVar(&options.OutDir, "out", options.OutDir, "Path to write any local output")
	// cmd.MarkFlagDirname("out")

	// These flags from the update command are specified to kubeconfig creation
	//
	// cmd.Flags().BoolVar(&options.CreateKubecfg, "create-kube-config", options.CreateKubecfg, "Will control automatically creating the kube config file on your local filesystem")
	// cmd.Flags().DurationVar(&options.Admin, "admin", options.Admin, "Also export a cluster admin user credential with the specified lifetime and add it to the cluster context")
	// cmd.Flags().Lookup("admin").NoOptDefVal = kubeconfig.DefaultKubecfgAdminLifetime.String()
	// cmd.Flags().StringVar(&options.User, "user", options.User, "Existing user in kubeconfig file to use.  Implies --create-kube-config")
	// cmd.RegisterFlagCompletionFunc("user", completeKubecfgUser)
	// cmd.Flags().BoolVar(&options.Internal, "internal", options.Internal, "Use the cluster's internal DNS name. Implies --create-kube-config")

	cmd.Flags().BoolVar(&options.AllowKopsDowngrade, "allow-kops-downgrade", options.AllowKopsDowngrade, "Allow an older version of kOps to update the cluster than last used")

	// These flags from the update command are not obviously needed by reconcile, though we can add them if needed:
	//
	// cmd.Flags().StringSliceVar(&options.InstanceGroups, "instance-group", options.InstanceGroups, "Instance groups to update (defaults to all if not specified)")
	// cmd.RegisterFlagCompletionFunc("instance-group", completeInstanceGroup(f, &options.InstanceGroups, &options.InstanceGroupRoles))
	// cmd.Flags().StringSliceVar(&options.InstanceGroupRoles, "instance-group-roles", options.InstanceGroupRoles, "Instance group roles to update ("+strings.Join(allRoles, ",")+")")
	// cmd.RegisterFlagCompletionFunc("instance-group-roles", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 	return sets.NewString(allRoles...).Delete(options.InstanceGroupRoles...).List(), cobra.ShellCompDirectiveNoFileComp
	// })
	// cmd.Flags().StringVar(&options.Phase, "phase", options.Phase, "Subset of tasks to run: "+strings.Join(cloudup.Phases.List(), ", "))
	// cmd.RegisterFlagCompletionFunc("phase", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 	return cloudup.Phases.List(), cobra.ShellCompDirectiveNoFileComp
	// })
	// cmd.Flags().StringSliceVar(&options.LifecycleOverrides, "lifecycle-overrides", options.LifecycleOverrides, "comma separated list of phase overrides, example: SecurityGroups=Ignore,InternetGateway=ExistsAndWarnIfChanges")
	// viper.BindPFlag("lifecycle-overrides", cmd.Flags().Lookup("lifecycle-overrides"))
	// viper.BindEnv("lifecycle-overrides", "KOPS_LIFECYCLE_OVERRIDES")
	// cmd.RegisterFlagCompletionFunc("lifecycle-overrides", completeLifecycleOverrides)
	// cmd.Flags().BoolVar(&options.Prune, "prune", options.Prune, "Delete old revisions of cloud resources that were needed during an upgrade")
	// cmd.Flags().BoolVar(&options.IgnoreKubeletVersionSkew, "ignore-kubelet-version-skew", options.IgnoreKubeletVersionSkew, "Setting this to true will force updating the kubernetes version on all instance groups, regardles of which control plane version is running")

	// cmd.Flags().BoolVar(&options.Reconcile, "reconcile", options.Reconcile, "Reconcile the cluster by rolling the control plane and nodes sequentially")

	return cmd
}

// ReconcileCluster updates the cluster to the desired state, including rolling updates where necessary.
// To respect skew policy, it updates the control plane first, then updates the nodes.
// "update" is probably now smart enough to automatically not update the control plane if it is already at the desired version,
// but we do it explicitly here to be clearer / safer.
func RunReconcileCluster(ctx context.Context, f *util.Factory, out io.Writer, c *CoreUpdateClusterOptions) error {
	if c.Target == cloudup.TargetTerraform {
		return fmt.Errorf("reconcile is not supported with terraform")
	}

	if !c.Yes {
		// A reconcile without --yes is the same as a dry run
		opt := *c
		if _, err := RunCoreUpdateCluster(ctx, f, out, &opt); err != nil {
			return err
		}
		return nil
	}

	fmt.Fprintf(out, "Updating control plane configuration\n")
	{
		opt := *c
		opt.InstanceGroupRoles = []string{
			string(kops.InstanceGroupRoleAPIServer),
			string(kops.InstanceGroupRoleControlPlane),
		}
		opt.Prune = false // Do not prune until after the last rolling update
		if _, err := RunCoreUpdateCluster(ctx, f, out, &opt); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Doing rolling-update for control plane\n")
	{
		opt := &RollingUpdateOptions{}
		opt.InitDefaults()
		opt.ClusterName = c.ClusterName
		opt.InstanceGroupRoles = []string{
			string(kops.InstanceGroupRoleAPIServer),
			string(kops.InstanceGroupRoleControlPlane),
		}
		opt.Yes = c.Yes
		if err := RunRollingUpdateCluster(ctx, f, out, opt); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Updating node configuration\n")
	{
		opt := *c
		// Do all roles this time, though we only expect changes to node & bastion roles
		opt.InstanceGroupRoles = nil
		opt.Prune = false // Do not prune until after the last rolling update
		if _, err := RunCoreUpdateCluster(ctx, f, out, &opt); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Doing rolling-update for nodes\n")
	{
		opt := &RollingUpdateOptions{}
		opt.InitDefaults()
		opt.ClusterName = c.ClusterName
		// Do all roles this time, though we only expect changes to node & bastion roles
		opt.InstanceGroupRoles = nil
		opt.Yes = c.Yes
		if err := RunRollingUpdateCluster(ctx, f, out, opt); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Pruning old resources that are no longer used\n")
	{
		opt := *c
		opt.InstanceGroupRoles = nil
		opt.Prune = true
		if _, err := RunCoreUpdateCluster(ctx, f, out, &opt); err != nil {
			return err
		}
	}

	return nil
}
