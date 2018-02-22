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
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/instancegroups"
	"k8s.io/kops/pkg/pretty"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/kutil"
	"k8s.io/kops/util/pkg/tables"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/kubectl/util/i18n"
)

var (
	rollingupdateLong = pretty.LongDesc(i18n.T(`
	This command updates a kubernetes cluster to match the cloud and kops specifications.

	To perform a rolling update, you need to update the cloud resources first with the command
	` + pretty.Bash("kops update cluster") + `.

	If rolling-update does not report that the cluster needs to be rolled, you can force the cluster to be
	rolled with the force flag.  Rolling update drains and validates the cluster by default.  A cluster is
	deemed validated when all required nodes are running and all pods in the kube-system namespace are operational.
	When a node is deleted, rolling-update sleeps the interval for the node type, and then tries for the same period
	of time for the cluster to be validated.  For instance, setting --master-interval=3m causes rolling-update
	to wait for 3 minutes after a master is rolled, and another 3 minutes for the cluster to stabilize and pass
	validation.

	Note: terraform users will need to run all of the following commands from the same directory
	` + pretty.Bash("kops update cluster --target=terraform") + ` then ` + pretty.Bash("terraform plan") + ` then
	` + pretty.Bash("terraform apply") + ` prior to running ` + pretty.Bash("kops rolling-update cluster") + `.`))

	rollingupdateExample = templates.Examples(i18n.T(`
		# Preview a rolling-update.
		kops rolling-update cluster

		# Roll the currently selected kops cluster with defaults.
		# Nodes will be drained and the cluster will be validated between node replacement.
		kops rolling-update cluster --yes

		# Roll the k8s-cluster.example.com kops cluster,
		# do not fail if the cluster does not validate,
		# wait 8 min to create new node, and wait at least
		# 8 min to validate the cluster.
		kops rolling-update cluster k8s-cluster.example.com --yes \
		  --fail-on-validate-error="false" \
		  --master-interval=8m \
		  --node-interval=8m

		# Roll the k8s-cluster.example.com kops cluster,
		# do not validate the cluster because of the cloudonly flag.
	    # Force the entire cluster to roll, even if rolling update
	    # reports that the cluster does not need to be rolled.
		kops rolling-update cluster k8s-cluster.example.com --yes \
	      --cloudonly \
		  --force

		# Roll the k8s-cluster.example.com kops cluster,
		# only roll the node instancegroup,
		# use the new drain an validate functionality.
		kops rolling-update cluster k8s-cluster.example.com --yes \
		  --fail-on-validate-error="false" \
		  --node-interval 8m \
		  --instance-group nodes
		`))

	rollingupdateShort = i18n.T(`Rolling update a cluster.`)
)

// RollingUpdateOptions is the command Object for a Rolling Update.
type RollingUpdateOptions struct {
	// BastionInterval is the minimum time to wait after stopping a bastion.  This does not include drain and validate time.
	BastionInterval time.Duration
	// Batch indicates the size of the batch to rollout i.e. rollout 2 at a time
	Batch int
	// CloudOnly indicate we only perform a cloud provider rollout, i.e. not kubernetes operation like draining
	CloudOnly bool
	// ClusterName is the name of the kops cluster
	ClusterName string
	// Count is an optional argument to perform an update on a subset of instances within an instancegroup
	Count int
	// Drain indicates we should drain the node between deleting it
	Drain *bool
	// DrainTimeout is the timeout for drain the node
	DrainTimeout time.Duration
	// FailOnDrainError fail rolling-update if drain errors.
	FailOnDrainError bool
	// FailOnValidate fail the cluster rolling-update when the cluster does not validate, after a validation period.
	FailOnValidate bool
	// Force indicates we should update the instance regardless whether we detect a change
	Force bool
	// InstanceGroups is the list of instance groups to rolling-update; if not specified, all instance groups will be updated
	InstanceGroups []string
	// InstanceGroupRoles is the list of roles we should rolling-update
	// if not specified, all instance groups will be updated
	InstanceGroupRoles []string
	// Interactive rolling-update prompts user to continue after each instances is updated.
	Interactive bool
	// MasterInterval is the minimum time to wait after stopping a master node.  This does not include drain and validate time.
	MasterInterval time.Duration
	// NodeBatch is the number of node instancegroups to rollout concurrently
	NodeBatch int
	// NodeInterval is the minimum time to wait after stopping a (non-master) node.  This does not include drain and validate time.
	NodeInterval time.Duration
	// PostDrainDelay is the duration of a pause after a drain operation
	PostDrainDelay time.Duration
	// ScaleTimeout is the duration to wait for a rescaling operation
	ScaleTimeout time.Duration
	// Strategy is the name of the rollout strategy to use
	Strategy string
	// ValidationTimeout is the timeout for validation to succeed after the drain and pause
	ValidationTimeout time.Duration
	// Yes is a confirmation option for rollout
	Yes bool
}

// InitDefaults are some sane defaults for the rollouts
func (o *RollingUpdateOptions) initDefaults() {
	o.BastionInterval = 5 * time.Minute
	o.Batch = 0
	o.CloudOnly = false
	o.Count = 0
	o.Drain = nil
	o.DrainTimeout = 5 * time.Minute
	o.FailOnDrainError = false
	o.FailOnValidate = true
	o.Force = false
	o.Interactive = false
	o.MasterInterval = 5 * time.Minute
	o.NodeBatch = 1
	o.NodeInterval = 4 * time.Minute
	o.PostDrainDelay = 20 * time.Second
	o.ScaleTimeout = 4 * time.Minute
	o.ValidationTimeout = 5 * time.Minute
	o.Yes = false
}

// NewCmdRollingUpdateCluster creates and returns the rollout command
func NewCmdRollingUpdateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	var options RollingUpdateOptions
	var drainOption bool

	options.initDefaults()

	cmd := &cobra.Command{
		Use:     "cluster",
		Example: rollingupdateExample,
		Long:    rollingupdateLong,
		Short:   rollingupdateShort,
	}

	cmd.Flags().BoolVar(&options.CloudOnly, "cloudonly", options.CloudOnly, "Perform rolling update without confirming progress with k8s")
	cmd.Flags().BoolVar(&drainOption, "drain", drainOption, "Indicates we should drain the node before terminating")
	cmd.Flags().BoolVar(&options.FailOnDrainError, "fail-on-drain-error", true, "The rolling-update will fail if draining a node fails.")
	cmd.Flags().BoolVar(&options.FailOnValidate, "fail-on-validate-error", true, "The rolling-update will fail if the cluster fails to validate.")
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force rolling update, even if no changes")
	cmd.Flags().BoolVarP(&options.Interactive, "interactive", "i", options.Interactive, "Prompt to continue after each instance is updated")
	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Perform rolling update immediately, without --yes rolling-update executes a dry-run")
	cmd.Flags().DurationVar(&options.BastionInterval, "bastion-interval", options.BastionInterval, "Time to wait between restarting bastions")
	cmd.Flags().DurationVar(&options.MasterInterval, "master-interval", options.MasterInterval, "Time to wait between restarting masters")
	cmd.Flags().DurationVar(&options.NodeInterval, "node-interval", options.NodeInterval, "Time to wait between restarting nodes")
	cmd.Flags().DurationVar(&options.ScaleTimeout, "scale-timeout", options.ScaleTimeout, "The max time to wait ")
	cmd.Flags().DurationVar(&options.PostDrainDelay, "post-drain-delay", options.PostDrainDelay, "Time to wait post draining a node to allow pods to settle")
	cmd.Flags().DurationVar(&options.ValidationTimeout, "valiation-timeout", options.ValidationTimeout, "Maxiumum time to wait for cluster validation")
	cmd.Flags().IntVar(&options.Batch, "batch", options.Batch, "Perform the rollout in batches, i.e. rollout to x at a time")
	cmd.Flags().IntVar(&options.Count, "count", options.Count, "Perform the rollout on only x instancegroup members, zero means all member")
	cmd.Flags().IntVar(&options.NodeBatch, "node-batch", options.NodeBatch, "The number of node instancegroup to run concurrently")
	cmd.Flags().StringSliceVar(&options.InstanceGroups, "instance-group", options.InstanceGroups, "List of instance groups to update (defaults to all if not specified)")
	cmd.Flags().StringSliceVar(&options.InstanceGroupRoles, "instance-group-roles", options.InstanceGroupRoles, "If specified, only instance groups of the specified role will be updated")
	cmd.Flags().StringVar(&options.Strategy, "strategy", options.Strategy, "The default rollout strategy to use when rolling out the cluster")

	// @check if the drain option was toggle as its the only way to find out if default false or actually set to false
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := rootCommand.ProcessArgs(args); err != nil {
			exitWithError(err)
			return
		}
		clusterName := rootCommand.ClusterName()
		if clusterName == "" {
			exitWithError(fmt.Errorf("--name is required"))
			return
		}
		options.ClusterName = clusterName
		if cmd.Flags().Changed("drain") {
			options.Drain = &drainOption
		}

		if err := RunRollingUpdateCluster(f, os.Stdout, &options); err != nil {
			exitWithError(err)
			return
		}
	}

	return cmd
}

// RunRollingUpdateCluster is responsible for performing an rolling update on the kops cluster
func RunRollingUpdateCluster(f *util.Factory, out io.Writer, options *RollingUpdateOptions) error {
	// @step: create a clientset for the kops cluster
	clientset, err := f.Clientset()
	if err != nil {
		return err
	}
	// retrieve the kops cluster configuration
	cluster, err := GetCluster(f, options.ClusterName)
	if err != nil {
		return err
	}

	contextName := cluster.ObjectMeta.Name
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: contextName}).ClientConfig()
	if err != nil {
		return fmt.Errorf("cannot load kubecfg settings for %q: %v", contextName, err)
	}

	var nodes []v1.Node
	var k8sClient kubernetes.Interface

	// @step: if we are not doing a cloudonly rollout, lets grab a list of nodes - it's debatable as to
	// whether this should be shifted into the rollout code itself.
	if !options.CloudOnly {
		k8sClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("cannot build kube client for %q: %v", contextName, err)
		}

		nodeList, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to reach the kubernetes API.\n")
			fmt.Fprintf(os.Stderr, "Use --cloudonly to do a rolling-update without confirming progress with the kubernetes API\n\n")
			return fmt.Errorf("error listing nodes in cluster: %v", err)
		}

		if nodeList != nil {
			nodes = nodeList.Items
		}
	}

	list, err := clientset.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	var instanceGroups []*api.InstanceGroup
	for i := range list.Items {
		instanceGroups = append(instanceGroups, &list.Items[i])
	}

	warnUnmatched := true

	if len(options.InstanceGroups) != 0 {
		var filtered []*api.InstanceGroup

		for _, instanceGroupName := range options.InstanceGroups {
			var found *api.InstanceGroup
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
		var filtered []*api.InstanceGroup

		for _, ig := range instanceGroups {
			for _, role := range options.InstanceGroupRoles {
				if ig.Spec.Role == api.InstanceGroupRole(role) {
					filtered = append(filtered, ig)
					continue
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

	// @NOTE: we need to add a warning for those choosing to validate but have a low node interval
	// as its unlikely kubernetes will detect the errors
	lowInterval := time.Second * 60
	logIntervalWaring := "Note: the %s interval on %s is low, kubernetes might not have time to detect node errors causing issues with validation"
	if options.MasterInterval < lowInterval {
		glog.Warningf(fmt.Sprintf(logIntervalWaring, "master", options.MasterInterval))
	}
	if options.NodeInterval < lowInterval {
		glog.Warningf(fmt.Sprintf(logIntervalWaring, "node", options.NodeInterval))
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

		columns := []string{"NAME", "STATUS", "NEEDUPDATE", "READY", "MIN", "MAX"}
		if !options.CloudOnly {
			columns = append(columns, "NODES")
		}

		if err := t.Render(l, out, columns...); err != nil {
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

	d := &instancegroups.RollingUpdateCluster{
		BastionInterval:       options.BastionInterval,
		Batch:                 options.Batch,
		Client:                k8sClient,
		ClientConfig:          kutil.NewClientConfig(config, "kube-system"),
		Clientset:             clientset,
		Cloud:                 cloud,
		CloudOnly:             options.CloudOnly,
		Cluster:               cluster,
		ClusterName:           options.ClusterName,
		Count:                 options.Count,
		Drain:                 options.Drain,
		DrainTimeout:          options.DrainTimeout,
		FailOnDrainError:      options.FailOnDrainError,
		FailOnValidate:        options.FailOnValidate,
		FailOnValidateTimeout: options.ValidationTimeout,
		Force:          options.Force,
		InstanceGroups: options.InstanceGroups,
		Interactive:    options.Interactive,
		MasterInterval: options.MasterInterval,
		NodeBatch:      options.NodeBatch,
		NodeInterval:   options.NodeInterval,
		PostDrainDelay: options.PostDrainDelay,
		ScaleTimeout:   options.ScaleTimeout,
		Strategy:       api.RolloutStrategy(options.Strategy),
	}

	ctx, cancel := context.WithCancel(context.Background())
	resultCh := d.RollingUpdate(ctx, &instancegroups.RollingUpdateOptions{
		InstanceGroups: groups,
		List:           list,
	})
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	for {
		select {
		case <-signalCh:
			glog.Infof("Recieved termination siganl, cancelling the rollout")
			cancel()
		case err := <-resultCh:
			if err != nil {
				return err
			}
			glog.Infof("Completed the kops cluster rollout")
			return nil
		}
	}

	return nil
}
