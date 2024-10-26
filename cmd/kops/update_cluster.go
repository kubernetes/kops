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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	updateClusterLong = templates.LongDesc(i18n.T(`
	Create or update cloud or cluster resources to match the current cluster and instance group definitions.
    If the cluster or cloud resources already exist this command may modify those resources.

	If, such as during a Kubernetes upgrade, nodes need updating, a rolling-update may
	be subsequently required.
	`))

	updateClusterExample = templates.Examples(i18n.T(`
	# After the cluster has been edited or upgraded, update the cloud resources with:
	kops update cluster k8s-cluster.example.com --yes --state=s3://my-state-store --yes
	`))

	updateClusterShort = i18n.T("Update a cluster.")
)

type UpdateClusterOptions struct {
	Yes                bool
	Target             string
	OutDir             string
	SSHPublicKey       string
	RunTasksOptions    fi.RunTasksOptions
	AllowKopsDowngrade bool
	// GetAssets is whether this is invoked from the CmdGetAssets.
	GetAssets bool

	ClusterName string

	CreateKubecfg bool
	admin         time.Duration
	user          string
	internal      bool

	// InstanceGroups is the list of instance groups to update;
	// if not specified, all instance groups will be updated
	InstanceGroups []string

	// InstanceGroupRoles is the list of roles we should update
	// if not specified, all instance groups will be updated
	InstanceGroupRoles []string

	Phase string

	// LifecycleOverrides is a slice of taskName=lifecycle name values.  This slice is used
	// to populate the LifecycleOverrides struct member in ApplyClusterCmd struct.
	LifecycleOverrides []string

	// Prune is true if we should clean up any old revisions of objects.
	// Typically this is done in after we have rolling-updated the cluster.
	// The goal is that the cluster can keep running even during more disruptive
	// infrastructure changes.
	Prune bool
}

func (o *UpdateClusterOptions) InitDefaults() {
	o.Yes = false
	o.Target = "direct"
	o.SSHPublicKey = ""
	o.OutDir = ""

	// By default we export a kubecfg, but it doesn't have a static/eternal credential in it any more.
	o.CreateKubecfg = true

	o.Prune = false

	o.RunTasksOptions.InitDefaults()
}

func NewCmdUpdateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &UpdateClusterOptions{}
	options.InitDefaults()

	allRoles := make([]string, 0, len(kops.AllInstanceGroupRoles))
	for _, r := range kops.AllInstanceGroupRoles {
		allRoles = append(allRoles, r.ToLowerString())
	}

	cmd := &cobra.Command{
		Use:               "cluster [CLUSTER]",
		Short:             updateClusterShort,
		Long:              updateClusterLong,
		Example:           updateClusterExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := RunUpdateCluster(cmd.Context(), f, out, options)
			return err
		},
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Create cloud resources, without --yes update is in dry run mode")
	cmd.Flags().StringVar(&options.Target, "target", options.Target, "Target - direct, terraform")
	cmd.RegisterFlagCompletionFunc("target", completeUpdateClusterTarget(f, options))
	cmd.Flags().StringVar(&options.SSHPublicKey, "ssh-public-key", options.SSHPublicKey, "SSH public key to use (deprecated: use kops create secret instead)")
	cmd.Flags().StringVar(&options.OutDir, "out", options.OutDir, "Path to write any local output")
	cmd.MarkFlagDirname("out")
	cmd.Flags().BoolVar(&options.CreateKubecfg, "create-kube-config", options.CreateKubecfg, "Will control automatically creating the kube config file on your local filesystem")
	cmd.Flags().DurationVar(&options.admin, "admin", options.admin, "Also export a cluster admin user credential with the specified lifetime and add it to the cluster context")
	cmd.Flags().Lookup("admin").NoOptDefVal = kubeconfig.DefaultKubecfgAdminLifetime.String()
	cmd.Flags().StringVar(&options.user, "user", options.user, "Existing user in kubeconfig file to use.  Implies --create-kube-config")
	cmd.RegisterFlagCompletionFunc("user", completeKubecfgUser)
	cmd.Flags().BoolVar(&options.internal, "internal", options.internal, "Use the cluster's internal DNS name. Implies --create-kube-config")
	cmd.Flags().BoolVar(&options.AllowKopsDowngrade, "allow-kops-downgrade", options.AllowKopsDowngrade, "Allow an older version of kOps to update the cluster than last used")
	cmd.Flags().StringSliceVar(&options.InstanceGroups, "instance-group", options.InstanceGroups, "Instance groups to update (defaults to all if not specified)")
	cmd.RegisterFlagCompletionFunc("instance-group", completeInstanceGroup(f, &options.InstanceGroups, &options.InstanceGroupRoles))
	cmd.Flags().StringSliceVar(&options.InstanceGroupRoles, "instance-group-roles", options.InstanceGroupRoles, "Instance group roles to update ("+strings.Join(allRoles, ",")+")")
	cmd.RegisterFlagCompletionFunc("instance-group-roles", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sets.NewString(allRoles...).Delete(options.InstanceGroupRoles...).List(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringVar(&options.Phase, "phase", options.Phase, "Subset of tasks to run: "+strings.Join(cloudup.Phases.List(), ", "))
	cmd.RegisterFlagCompletionFunc("phase", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return cloudup.Phases.List(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().StringSliceVar(&options.LifecycleOverrides, "lifecycle-overrides", options.LifecycleOverrides, "comma separated list of phase overrides, example: SecurityGroups=Ignore,InternetGateway=ExistsAndWarnIfChanges")
	viper.BindPFlag("lifecycle-overrides", cmd.Flags().Lookup("lifecycle-overrides"))
	viper.BindEnv("lifecycle-overrides", "KOPS_LIFECYCLE_OVERRIDES")
	cmd.RegisterFlagCompletionFunc("lifecycle-overrides", completeLifecycleOverrides)

	cmd.Flags().BoolVar(&options.Prune, "prune", options.Prune, "Delete old revisions of cloud resources that were needed during an upgrade")

	return cmd
}

type UpdateClusterResults struct {
	// Target is the fi.Target we will operate against.  This can be used to get dryrun results (primarily for tests)
	Target fi.CloudupTarget

	// TaskMap is the map of tasks that we built (output)
	TaskMap map[string]fi.CloudupTask

	// ImageAssets are the image assets we use (output).
	ImageAssets []*assets.ImageAsset
	// FileAssets are the file assets we use (output).
	FileAssets []*assets.FileAsset
	// Cluster is the cluster spec (output).
	Cluster *kops.Cluster
}

func RunUpdateCluster(ctx context.Context, f *util.Factory, out io.Writer, c *UpdateClusterOptions) (*UpdateClusterResults, error) {
	results := &UpdateClusterResults{}

	isDryrun := false
	targetName := c.Target

	if c.admin != 0 && c.user != "" {
		return nil, fmt.Errorf("cannot use both --admin and --user")
	}

	if c.admin != 0 && !c.CreateKubecfg {
		klog.Info("--admin implies --create-kube-config")
		c.CreateKubecfg = true
	}

	if c.user != "" && !c.CreateKubecfg {
		klog.Info("--user implies --create-kube-config")
		c.CreateKubecfg = true
	}

	if c.internal && !c.CreateKubecfg {
		klog.Info("--internal implies --create-kube-config")
		c.CreateKubecfg = true
	}

	// direct requires --yes (others do not, because they don't do anything!)
	if c.Target == cloudup.TargetDirect {
		if !c.Yes {
			isDryrun = true
			targetName = cloudup.TargetDryRun
		}
	}
	if c.Target == cloudup.TargetDryRun {
		isDryrun = true
		targetName = cloudup.TargetDryRun
	}

	if c.OutDir == "" {
		if c.Target == cloudup.TargetTerraform {
			c.OutDir = "out/terraform"
		} else {
			c.OutDir = "out"
		}
	}

	cluster, err := GetCluster(ctx, f, c.ClusterName)
	if err != nil {
		return results, err
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return results, err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return results, err
	}

	sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
	if err != nil {
		return results, err
	}

	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return results, err
	}

	if c.SSHPublicKey != "" {
		fmt.Fprintf(out, "--ssh-public-key on update is deprecated - please use `kops create secret --name %s sshpublickey admin -i ~/.ssh/id_rsa.pub` instead\n", cluster.ObjectMeta.Name)

		c.SSHPublicKey = utils.ExpandPath(c.SSHPublicKey)
		authorized, err := os.ReadFile(c.SSHPublicKey)
		if err != nil {
			return results, fmt.Errorf("error reading SSH key file %q: %v", c.SSHPublicKey, err)
		}
		err = sshCredentialStore.AddSSHPublicKey(ctx, authorized)
		if err != nil {
			return results, fmt.Errorf("error adding SSH public key: %v", err)
		}

		klog.Infof("Using SSH public key: %v\n", c.SSHPublicKey)
	}

	var phase cloudup.Phase
	if c.Phase != "" {
		switch strings.ToLower(c.Phase) {
		case string(cloudup.PhaseNetwork):
			phase = cloudup.PhaseNetwork
		case string(cloudup.PhaseSecurity), "iam": // keeping IAM for backwards compatibility
			phase = cloudup.PhaseSecurity
		case string(cloudup.PhaseCluster):
			phase = cloudup.PhaseCluster
		default:
			return results, fmt.Errorf("unknown phase %q, available phases: %s", c.Phase, strings.Join(cloudup.Phases.List(), ","))
		}
	}

	deletionProcessing := fi.DeletionProcessingModeDeleteIfNotDeferrred
	if c.Prune {
		deletionProcessing = fi.DeletionProcessingModeDeleteIncludingDeferred
	}

	lifecycleOverrideMap := make(map[string]fi.Lifecycle)

	for _, override := range c.LifecycleOverrides {
		values := strings.Split(override, "=")
		if len(values) != 2 {
			return results, fmt.Errorf("incorrect syntax for lifecyle-overrides, correct syntax is TaskName=lifecycleName, override provided: %q", override)
		}

		taskName := values[0]
		lifecycleName := values[1]

		lifecycleOverride, err := parseLifecycle(lifecycleName)
		if err != nil {
			return nil, err
		}

		lifecycleOverrideMap[taskName] = lifecycleOverride
	}

	list, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var allInstanceGroups []*kops.InstanceGroup
	for i := range list.Items {
		instanceGroup := &list.Items[i]
		allInstanceGroups = append(allInstanceGroups, instanceGroup)
	}

	var filteredInstanceGroups []*kops.InstanceGroup
	if len(c.InstanceGroups) != 0 {
		var filtered []*kops.InstanceGroup

		for _, instanceGroupName := range c.InstanceGroups {
			var found *kops.InstanceGroup
			for _, ig := range allInstanceGroups {
				if ig.ObjectMeta.Name == instanceGroupName {
					found = ig
					break
				}
			}
			if found == nil {
				return nil, fmt.Errorf("instance group %q not found", instanceGroupName)
			}

			filtered = append(filtered, found)
		}

		filteredInstanceGroups = filtered
	}
	if len(c.InstanceGroupRoles) != 0 {
		var filtered []*kops.InstanceGroup

		for _, role := range c.InstanceGroupRoles {
			s, f := kops.ParseInstanceGroupRole(role, true)
			if !f {
				return nil, fmt.Errorf("instance group role %q invalid", role)
			}
			for _, ig := range allInstanceGroups {
				if ig.Spec.Role == s {
					filtered = append(filtered, ig)
				}
			}
		}

		filteredInstanceGroups = filtered
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	applyCmd := &cloudup.ApplyClusterCmd{
		Cloud:                  cloud,
		Clientset:              clientset,
		Cluster:                cluster,
		DryRun:                 isDryrun,
		AllowKopsDowngrade:     c.AllowKopsDowngrade,
		RunTasksOptions:        &c.RunTasksOptions,
		OutDir:                 c.OutDir,
		FilteredInstanceGroups: filteredInstanceGroups,
		Phase:                  phase,
		TargetName:             targetName,
		LifecycleOverrides:     lifecycleOverrideMap,
		GetAssets:              c.GetAssets,
		DeletionProcessing:     deletionProcessing,
	}

	applyResults, err := applyCmd.Run(ctx)
	if err != nil {
		return results, err
	}

	results.Target = applyCmd.Target
	results.TaskMap = applyCmd.TaskMap
	results.ImageAssets = applyResults.AssetBuilder.ImageAssets
	results.FileAssets = applyResults.AssetBuilder.FileAssets
	results.Cluster = cluster

	if isDryrun && !c.GetAssets {
		target := applyCmd.Target.(*fi.CloudupDryRunTarget)
		if target.HasChanges() {
			fmt.Fprintf(out, "Must specify --yes to apply changes\n")
		} else {
			fmt.Fprintf(out, "No changes need to be applied\n")
		}
		return results, nil
	}

	firstRun := false

	if !isDryrun && c.CreateKubecfg {
		hasKubeconfig, err := clusterIsInKubeConfig(cluster.ObjectMeta.Name)
		if err != nil {
			klog.Warningf("error reading kubeconfig: %v", err)
			hasKubeconfig = true
		}
		firstRun = !hasKubeconfig

		klog.Infof("Exporting kubeconfig for cluster")

		// TODO: Another flag?
		useKopsAuthenticationPlugin := false
		conf, err := kubeconfig.BuildKubecfg(
			ctx,
			cluster,
			keyStore,
			secretStore,
			cloud,
			c.admin,
			c.user,
			c.internal,
			f.KopsStateStore(),
			useKopsAuthenticationPlugin)
		if err != nil {
			return nil, err
		}

		err = conf.WriteKubecfg(clientcmd.NewDefaultPathOptions())
		if err != nil {
			return nil, err
		}

		if c.admin == 0 && c.user == "" {
			klog.Warningf("Exported kubeconfig with no user authentication; use --admin, --user or --auth-plugin flags with `kops export kubeconfig`")
		}
	}

	if !isDryrun {
		sb := new(bytes.Buffer)

		if c.Target == cloudup.TargetTerraform {
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Terraform output has been placed into %s\n", c.OutDir)

			if firstRun {
				fmt.Fprintf(sb, "Run these commands to apply the configuration:\n")
				fmt.Fprintf(sb, "   cd %s\n", c.OutDir)
				fmt.Fprintf(sb, "   terraform plan\n")
				fmt.Fprintf(sb, "   terraform apply\n")
				fmt.Fprintf(sb, "\n")
			}
		} else if firstRun {
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Cluster is starting.  It should be ready in a few minutes.\n")
			fmt.Fprintf(sb, "\n")
		} else {
			// TODO: Different message if no changes were needed
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Cluster changes have been applied to the cloud.\n")
			fmt.Fprintf(sb, "\n")
		}

		// More suggestions on first run
		if firstRun {
			fmt.Fprintf(sb, "Suggestions:\n")
			fmt.Fprintf(sb, " * validate cluster: kops validate cluster --wait 10m\n")
			fmt.Fprintf(sb, " * list nodes: kubectl get nodes --show-labels\n")
			if !usesBastion(applyCmd.InstanceGroups) {
				fmt.Fprintf(sb, " * ssh to a control-plane node: ssh -i ~/.ssh/id_rsa ubuntu@%s\n", cluster.Spec.API.PublicName)
			} else {
				bastionPublicName := findBastionPublicName(cluster)
				if bastionPublicName != "" {
					fmt.Fprintf(sb, " * ssh to the bastion: ssh -A -i ~/.ssh/id_rsa ubuntu@%s\n", bastionPublicName)
				} else {
					fmt.Fprintf(sb, " * to ssh to the bastion, you probably want to configure a bastionPublicName.\n")
				}
			}
			fmt.Fprintf(sb, " * the ubuntu user is specific to Ubuntu. If not using Ubuntu please use the appropriate user based on your OS.\n")
			fmt.Fprintf(sb, " * read about installing addons at: https://kops.sigs.k8s.io/addons.\n")
			fmt.Fprintf(sb, "\n")
		}

		if !firstRun {
			// TODO: Detect if rolling-update is needed
			fmt.Fprintf(sb, "\n")
			fmt.Fprintf(sb, "Changes may require instances to restart: kops rolling-update cluster\n")
			fmt.Fprintf(sb, "\n")
		}

		_, err := out.Write(sb.Bytes())
		if err != nil {
			return nil, fmt.Errorf("error writing to output: %v", err)
		}
	}

	return results, nil
}

func parseLifecycle(lifecycle string) (fi.Lifecycle, error) {
	if v, ok := fi.LifecycleNameMap[lifecycle]; ok {
		return v, nil
	}
	return "", fmt.Errorf("unknown lifecycle %q, available lifecycle: %s", lifecycle, strings.Join(fi.Lifecycles.List(), ","))
}

func usesBastion(instanceGroups []*kops.InstanceGroup) bool {
	for _, ig := range instanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleBastion {
			return true
		}
	}

	return false
}

func findBastionPublicName(c *kops.Cluster) string {
	topology := c.Spec.Networking.Topology
	if topology == nil {
		return ""
	}
	bastion := topology.Bastion
	if bastion == nil {
		return ""
	}
	return bastion.PublicName
}

// clusterIsInKubeConfig checks if we have a context with the specified name (cluster name) in ~/.kube/config.
// It is used as a check to see if this is (likely) a new cluster.
func clusterIsInKubeConfig(contextName string) (bool, error) {
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return false, fmt.Errorf("error reading kubeconfig: %w", err)
	}

	for k := range config.Contexts {
		if k == contextName {
			return true, nil
		}
	}

	return false, nil
}

func completeUpdateClusterTarget(f commandutils.Factory, options *UpdateClusterOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := cmd.Context()

		commandutils.ConfigureKlogForCompletion()

		cluster, _, _, directive := GetClusterForCompletion(ctx, f, nil)
		if cluster == nil {
			return []string{
				cloudup.TargetDirect,
				cloudup.TargetDryRun,
				cloudup.TargetTerraform,
			}, directive
		}

		completions := []string{
			cloudup.TargetDirect,
			cloudup.TargetDryRun,
		}
		for _, cp := range cloudup.TerraformCloudProviders {
			if cluster.GetCloudProvider() == cp {
				completions = append(completions, cloudup.TargetTerraform)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func completeLifecycleOverrides(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	split := strings.SplitAfter(toComplete, "=")

	if len(split) < 2 {
		// providing completion for task names is too complicated
		return nil, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	}
	if len(split) > 2 {
		return commandutils.CompletionError("too many = characters", nil)
	}

	var completions []string
	for lifecycle := range fi.LifecycleNameMap {
		completions = append(completions, split[0]+lifecycle)
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
