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
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/instancegroups"
	"k8s.io/kops/pkg/pretty"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	rollingupdateLong = pretty.LongDesc(i18n.T(`
	This command updates a kubernetes cluster to match the cloud and kOps specifications.

	To perform a rolling update, you need to update the cloud resources first with the command
	` + pretty.Bash("kops update cluster --yes") + `. Nodes may be additionally marked for update by placing a
	` + pretty.Bash("kops.k8s.io/needs-update") + ` annotation on them.

	If rolling-update does not report that the cluster needs to be updated, you can force the cluster to be
	updated with the --force flag.  Rolling update drains and validates the cluster by default.  A cluster is
	deemed validated when all required nodes are running and all pods with a critical priority are operational.

	Note: terraform users will need to run all of the following commands from the same directory
	` + pretty.Bash("kops update cluster --target=terraform") + ` then ` + pretty.Bash("terraform plan") + ` then
	` + pretty.Bash("terraform apply") + ` prior to running ` + pretty.Bash("kops rolling-update cluster") + `.`))

	rollingupdateExample = templates.Examples(i18n.T(`
		# Preview a rolling update.
		kops rolling-update cluster

		# Update the currently selected kOps cluster with defaults.
		# Nodes will be drained and the cluster will be validated between node replacement.
		kops rolling-update cluster --yes

		# Update the k8s-cluster.example.com kOps cluster.
		# Do not fail if the cluster does not validate.
		kops rolling-update cluster k8s-cluster.example.com --yes \
		  --fail-on-validate-error="false"

		# Update the k8s-cluster.example.com kOps cluster.
		# Do not validate the cluster.
	    # Force the entire cluster to update, even if rolling update
	    # reports that the cluster does not need to be updated.
		kops rolling-update cluster k8s-cluster.example.com --yes \
	      --cloudonly \
		  --force

		# Update only the "nodes-1a" instance group of the k8s-cluster.example.com kOps cluster.
		kops rolling-update cluster k8s-cluster.example.com --yes \
		  --instance-group nodes-1a
		`))

	rollingupdateShort = i18n.T(`Rolling update a cluster.`)
)

// RollingUpdateOptions is the command Object for a Rolling Update.
type RollingUpdateOptions struct {
	Yes       bool
	Force     bool
	CloudOnly bool

	// The following two variables are when kOps is validating a cluster
	// during a rolling update.

	// FailOnDrainError fail rolling-update if drain errors.
	FailOnDrainError bool

	// FailOnValidate fail the cluster rolling-update when the cluster
	// does not validate, after a validation period.
	FailOnValidate bool

	// DrainTimeout is the maximum time to wait while draining a node.
	DrainTimeout time.Duration

	// PostDrainDelay is the duration of a pause after a drain operation
	PostDrainDelay time.Duration

	// ValidationTimeout is the timeout for validation to succeed after the drain and pause
	ValidationTimeout time.Duration

	// ValidateCount is the amount of time that a cluster needs to be validated after single node update
	ValidateCount int32

	// MasterInterval is the minimum time to wait after stopping a master node.  This does not include drain and validate time.
	MasterInterval time.Duration

	// NodeInterval is the minimum time to wait after stopping a (non-master) node.  This does not include drain and validate time.
	NodeInterval time.Duration

	// BastionInterval is the minimum time to wait after stopping a bastion.  This does not include drain and validate time.
	BastionInterval time.Duration

	// Interactive rolling-update prompts user to continue after each instances is updated.
	Interactive bool

	ClusterName string

	// InstanceGroups is the list of instance groups to rolling-update;
	// if not specified, all instance groups will be updated
	InstanceGroups []string

	// InstanceGroupRoles is the list of roles we should rolling-update
	// if not specified, all instance groups will be updated
	InstanceGroupRoles []string

	// TODO: Move more/all above options to RollingUpdateOptions
	instancegroups.RollingUpdateOptions
}

func (o *RollingUpdateOptions) InitDefaults() {
	o.Yes = false
	o.Force = false
	o.CloudOnly = false
	o.FailOnDrainError = false
	o.FailOnValidate = true

	o.MasterInterval = 15 * time.Second
	o.NodeInterval = 15 * time.Second
	o.BastionInterval = 15 * time.Second
	o.Interactive = false

	o.PostDrainDelay = 5 * time.Second
	o.ValidationTimeout = 15 * time.Minute
	o.ValidateCount = 2

	o.DrainTimeout = 15 * time.Minute

	o.RollingUpdateOptions.InitDefaults()
}

func NewCmdRollingUpdateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	var options RollingUpdateOptions
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:               "cluster [CLUSTER]",
		Short:             rollingupdateShort,
		Long:              rollingupdateLong,
		Example:           rollingupdateExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunRollingUpdateCluster(context.TODO(), f, out, &options)
		},
	}

	allRoles := make([]string, 0, len(kopsapi.AllInstanceGroupRoles))
	for _, r := range kopsapi.AllInstanceGroupRoles {
		allRoles = append(allRoles, strings.ToLower(string(r)))
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Perform rolling update immediately; without --yes rolling-update executes a dry-run")
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force rolling update, even if no changes")
	cmd.Flags().BoolVar(&options.CloudOnly, "cloudonly", options.CloudOnly, "Perform rolling update without confirming progress with Kubernetes")

	cmd.Flags().DurationVar(&options.ValidationTimeout, "validation-timeout", options.ValidationTimeout, "Maximum time to wait for a cluster to validate")
	cmd.Flags().DurationVar(&options.DrainTimeout, "drain-timeout", options.DrainTimeout, "Maximum time to wait for a node to drain")
	cmd.Flags().Int32Var(&options.ValidateCount, "validate-count", options.ValidateCount, "Number of times that a cluster needs to be validated after single node update")
	cmd.Flags().DurationVar(&options.MasterInterval, "master-interval", options.MasterInterval, "Time to wait between restarting control plane nodes")
	cmd.Flags().DurationVar(&options.NodeInterval, "node-interval", options.NodeInterval, "Time to wait between restarting worker nodes")
	cmd.Flags().DurationVar(&options.BastionInterval, "bastion-interval", options.BastionInterval, "Time to wait between restarting bastions")
	cmd.Flags().DurationVar(&options.PostDrainDelay, "post-drain-delay", options.PostDrainDelay, "Time to wait after draining each node")
	cmd.Flags().BoolVarP(&options.Interactive, "interactive", "i", options.Interactive, "Prompt to continue after each instance is updated")
	cmd.Flags().StringSliceVar(&options.InstanceGroups, "instance-group", options.InstanceGroups, "Instance groups to update (defaults to all if not specified)")
	cmd.RegisterFlagCompletionFunc("instance-group", completeInstanceGroup(f, &options.InstanceGroups, &options.InstanceGroupRoles))
	cmd.Flags().StringSliceVar(&options.InstanceGroupRoles, "instance-group-roles", options.InstanceGroupRoles, "Instance group roles to update ("+strings.Join(allRoles, ",")+")")
	cmd.RegisterFlagCompletionFunc("instance-group-roles", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sets.NewString(allRoles...).Delete(options.InstanceGroupRoles...).List(), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.Flags().BoolVar(&options.FailOnDrainError, "fail-on-drain-error", true, "Fail if draining a node fails")
	cmd.Flags().BoolVar(&options.FailOnValidate, "fail-on-validate-error", true, "Fail if the cluster fails to validate")

	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "ig", "instance-groups":
			name = "instance-group"
		case "role", "roles", "instance-group-role":
			name = "instance-group-roles"
		}
		return pflag.NormalizedName(name)
	})

	return cmd
}

func RunRollingUpdateCluster(ctx context.Context, f *util.Factory, out io.Writer, options *RollingUpdateOptions) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	contextName := cluster.ObjectMeta.Name
	clientGetter := genericclioptions.NewConfigFlags(true)
	clientGetter.Context = &contextName

	config, err := clientGetter.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("cannot load kubecfg settings for %q: %v", contextName, err)
	}

	var nodes []v1.Node
	var k8sClient kubernetes.Interface
	if !options.CloudOnly {
		k8sClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("cannot build kube client for %q: %v", contextName, err)
		}

		nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to reach the kubernetes API.\n")
			fmt.Fprintf(os.Stderr, "Use --cloudonly to do a rolling-update without confirming progress with the k8s API\n\n")
			return fmt.Errorf("error listing nodes in cluster: %v", err)
		}

		if nodeList != nil {
			nodes = nodeList.Items
		}
	}

	list, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	countByRole := make(map[kopsapi.InstanceGroupRole]int32)
	var instanceGroups []*kopsapi.InstanceGroup
	for i := range list.Items {
		instanceGroup := &list.Items[i]
		instanceGroups = append(instanceGroups, instanceGroup)

		minSize := int32(1)
		if instanceGroup.Spec.MinSize != nil {
			minSize = *instanceGroup.Spec.MinSize
		}
		countByRole[instanceGroup.Spec.Role] = countByRole[instanceGroup.Spec.Role] + minSize
	}
	if countByRole[kopsapi.InstanceGroupRoleAPIServer]+countByRole[kopsapi.InstanceGroupRoleMaster] <= 1 {
		fmt.Fprintf(out, "Detected single-control-plane cluster; won't detach before draining\n")
		options.DeregisterControlPlaneNodes = false
	}

	warnUnmatched := true

	if len(options.InstanceGroups) != 0 {
		var filtered []*kopsapi.InstanceGroup

		for _, instanceGroupName := range options.InstanceGroups {
			var found *kopsapi.InstanceGroup
			for _, ig := range instanceGroups {
				if ig.ObjectMeta.Name == instanceGroupName {
					found = ig
					break
				}
			}
			if found == nil {
				return fmt.Errorf("InstanceGroup %q not found", instanceGroupName)
			}

			filtered = append(filtered, found)
		}

		instanceGroups = filtered

		// Don't warn if we find more ASGs than IGs
		warnUnmatched = false
	}

	if len(options.InstanceGroupRoles) != 0 {
		var filtered []*kopsapi.InstanceGroup

		for _, role := range options.InstanceGroupRoles {
			s, f := kopsapi.ParseInstanceGroupRole(role, true)
			if !f {
				return fmt.Errorf("invalid instance group role %q", role)
			}
			for _, ig := range instanceGroups {
				if ig.Spec.Role == s {
					filtered = append(filtered, ig)
				}
			}
		}

		instanceGroups = filtered

		// Don't warn if we find more ASGs than IGs
		warnUnmatched = false
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	groups, err := cloud.GetCloudGroups(cluster, instanceGroups, warnUnmatched, nodes)
	if err != nil {
		return err
	}

	d := &instancegroups.RollingUpdateCluster{
		Clientset:         clientset,
		Ctx:               ctx,
		Cluster:           cluster,
		MasterInterval:    options.MasterInterval,
		NodeInterval:      options.NodeInterval,
		BastionInterval:   options.BastionInterval,
		Interactive:       options.Interactive,
		Force:             options.Force,
		Cloud:             cloud,
		K8sClient:         k8sClient,
		FailOnDrainError:  options.FailOnDrainError,
		FailOnValidate:    options.FailOnValidate,
		CloudOnly:         options.CloudOnly,
		ClusterName:       options.ClusterName,
		PostDrainDelay:    options.PostDrainDelay,
		ValidationTimeout: options.ValidationTimeout,
		ValidateCount:     int(options.ValidateCount),
		DrainTimeout:      options.DrainTimeout,
		// TODO should we expose this to the UI?
		ValidateTickDuration:    30 * time.Second,
		ValidateSuccessDuration: 10 * time.Second,

		// TODO: Move more of the passthrough options here, instead of duplicating them.
		Options: options.RollingUpdateOptions,
	}

	err = d.AdjustNeedUpdate(groups)
	if err != nil {
		return err
	}

	{
		t := &tables.Table{}
		t.AddColumn("NAME", func(r *cloudinstances.CloudInstanceGroup) string {
			return r.InstanceGroup.ObjectMeta.Name
		})
		t.AddColumn("STATUS", func(r *cloudinstances.CloudInstanceGroup) string {
			return r.Status()
		})
		t.AddColumn("NEEDUPDATE", func(r *cloudinstances.CloudInstanceGroup) string {
			return strconv.Itoa(len(r.NeedUpdate))
		})
		t.AddColumn("READY", func(r *cloudinstances.CloudInstanceGroup) string {
			return strconv.Itoa(len(r.Ready))
		})
		t.AddColumn("MIN", func(r *cloudinstances.CloudInstanceGroup) string {
			return strconv.Itoa(r.MinSize)
		})
		t.AddColumn("TARGET", func(r *cloudinstances.CloudInstanceGroup) string {
			return strconv.Itoa(r.TargetSize)
		})
		t.AddColumn("MAX", func(r *cloudinstances.CloudInstanceGroup) string {
			return strconv.Itoa(r.MaxSize)
		})
		t.AddColumn("NODES", func(r *cloudinstances.CloudInstanceGroup) string {
			var nodes []*v1.Node
			for _, i := range r.Ready {
				if i.Node != nil {
					nodes = append(nodes, i.Node)
				}
			}
			for _, i := range r.NeedUpdate {
				if i.Node != nil {
					nodes = append(nodes, i.Node)
				}
			}
			return strconv.Itoa(len(nodes))
		})
		var l []*cloudinstances.CloudInstanceGroup
		for _, v := range groups {
			l = append(l, v)
		}

		columns := []string{"NAME", "STATUS", "NEEDUPDATE", "READY", "MIN", "TARGET", "MAX"}
		if !options.CloudOnly {
			columns = append(columns, "NODES")
		}
		err := t.Render(l, out, columns...)
		if err != nil {
			return err
		}
	}

	needUpdate := false
	for _, group := range groups {
		if len(group.NeedUpdate) != 0 {
			needUpdate = true
		}
	}

	if !needUpdate && !options.Force {
		fmt.Printf("\nNo rolling-update required.\n")
		return nil
	}

	if !options.Yes {
		fmt.Printf("\nMust specify --yes to rolling-update.\n")
		return nil
	}

	var clusterValidator validation.ClusterValidator
	if !options.CloudOnly {
		clusterValidator, err = validation.NewClusterValidator(cluster, cloud, list, config.Host, k8sClient)
		if err != nil {
			return fmt.Errorf("cannot create cluster validator: %v", err)
		}
	}
	d.ClusterValidator = clusterValidator

	return d.RollingUpdate(groups, list)
}

func completeInstanceGroup(f commandutils.Factory, selectedInstanceGroups *[]string, selectedInstanceGroupRoles *[]string) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		commandutils.ConfigureKlogForCompletion()
		ctx := context.TODO()

		cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, f, args)
		if cluster == nil {
			return completions, directive
		}

		list, err := clientSet.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
		if err != nil {
			return commandutils.CompletionError("listing instance groups", err)
		}

		alreadySelected := sets.NewString()
		if selectedInstanceGroups != nil {
			alreadySelected = alreadySelected.Insert(*selectedInstanceGroups...)
		}
		alreadySelectedRoles := sets.NewString()
		if selectedInstanceGroupRoles != nil {
			alreadySelectedRoles = alreadySelectedRoles.Insert(*selectedInstanceGroupRoles...)
		}
		var igs []string
		for _, ig := range list.Items {
			if !alreadySelected.Has(ig.Name) && !alreadySelectedRoles.Has(strings.ToLower(string(ig.Spec.Role))) {
				igs = append(igs, ig.Name)
			}
		}

		return igs, cobra.ShellCompDirectiveNoFileComp
	}
}
