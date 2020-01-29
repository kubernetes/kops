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
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
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
		# use the new drain and validate functionality.
		kops rolling-update cluster k8s-cluster.example.com --yes \
		  --fail-on-validate-error="false" \
		  --node-interval 8m \
		  --instance-group nodes
		`))

	rollingupdateShort = i18n.T(`Rolling update a cluster.`)
)

// RollingUpdateOptions is the command Object for a Rolling Update.
type RollingUpdateOptions struct {
	Yes       bool
	Force     bool
	CloudOnly bool

	// The following two variables are when kops is validating a cluster
	// during a rolling update.

	// FailOnDrainError fail rolling-update if drain errors.
	FailOnDrainError bool

	// FailOnValidate fail the cluster rolling-update when the cluster
	// does not validate, after a validation period.
	FailOnValidate bool

	// PostDrainDelay is the duration of a pause after a drain operation
	PostDrainDelay time.Duration

	// ValidationTimeout is the timeout for validation to succeed after the drain and pause
	ValidationTimeout time.Duration

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
}

func NewCmdRollingUpdateCluster(f *util.Factory, out io.Writer) *cobra.Command {

	var options RollingUpdateOptions
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   rollingupdateShort,
		Long:    rollingupdateLong,
		Example: rollingupdateExample,
	}

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Perform rolling update immediately, without --yes rolling-update executes a dry-run")
	cmd.Flags().BoolVar(&options.Force, "force", options.Force, "Force rolling update, even if no changes")
	cmd.Flags().BoolVar(&options.CloudOnly, "cloudonly", options.CloudOnly, "Perform rolling update without confirming progress with k8s")

	cmd.Flags().DurationVar(&options.ValidationTimeout, "validation-timeout", options.ValidationTimeout, "Maximum time to wait for a cluster to validate")
	cmd.Flags().DurationVar(&options.MasterInterval, "master-interval", options.MasterInterval, "Time to wait between restarting masters")
	cmd.Flags().DurationVar(&options.NodeInterval, "node-interval", options.NodeInterval, "Time to wait between restarting nodes")
	cmd.Flags().DurationVar(&options.BastionInterval, "bastion-interval", options.BastionInterval, "Time to wait between restarting bastions")
	cmd.Flags().DurationVar(&options.PostDrainDelay, "post-drain-delay", options.PostDrainDelay, "Time to wait after draining each node")
	cmd.Flags().BoolVarP(&options.Interactive, "interactive", "i", options.Interactive, "Prompt to continue after each instance is updated")
	cmd.Flags().StringSliceVar(&options.InstanceGroups, "instance-group", options.InstanceGroups, "List of instance groups to update (defaults to all if not specified)")
	cmd.Flags().StringSliceVar(&options.InstanceGroupRoles, "instance-group-roles", options.InstanceGroupRoles, "If specified, only instance groups of the specified role will be updated (e.g. Master,Node,Bastion)")

	cmd.Flags().BoolVar(&options.FailOnDrainError, "fail-on-drain-error", true, "The rolling-update will fail if draining a node fails.")
	cmd.Flags().BoolVar(&options.FailOnValidate, "fail-on-validate-error", true, "The rolling-update will fail if the cluster fails to validate.")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := rootCommand.ProcessArgs(args)
		if err != nil {
			exitWithError(err)
			return
		}

		clusterName := rootCommand.ClusterName()
		if clusterName == "" {
			exitWithError(fmt.Errorf("--name is required"))
			return
		}

		options.ClusterName = clusterName

		err = RunRollingUpdateCluster(f, os.Stdout, &options)
		if err != nil {
			exitWithError(err)
			return
		}

	}

	return cmd
}

func RunRollingUpdateCluster(f *util.Factory, out io.Writer, options *RollingUpdateOptions) error {

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	cluster, err := GetCluster(f, options.ClusterName)
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

		nodeList, err := k8sClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to reach the kubernetes API.\n")
			fmt.Fprintf(os.Stderr, "Use --cloudonly to do a rolling-update without confirming progress with the k8s API\n\n")
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
				if ig.Spec.Role == api.InstanceGroupRole(strings.Title(strings.ToLower(role))) {
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
		clusterValidator, err = validation.NewClusterValidator(cluster, cloud, list, k8sClient)
		if err != nil {
			return fmt.Errorf("cannot create cluster validator: %v", err)
		}
	}
	d := &instancegroups.RollingUpdateCluster{
		MasterInterval:    options.MasterInterval,
		NodeInterval:      options.NodeInterval,
		BastionInterval:   options.BastionInterval,
		Interactive:       options.Interactive,
		Force:             options.Force,
		Cloud:             cloud,
		K8sClient:         k8sClient,
		ClusterValidator:  clusterValidator,
		FailOnDrainError:  options.FailOnDrainError,
		FailOnValidate:    options.FailOnValidate,
		CloudOnly:         options.CloudOnly,
		ClusterName:       options.ClusterName,
		PostDrainDelay:    options.PostDrainDelay,
		ValidationTimeout: options.ValidationTimeout,
		// TODO should we expose this to the UI?
		ValidateTickDuration:    30 * time.Second,
		ValidateSuccessDuration: 10 * time.Second,
	}
	return d.RollingUpdate(groups, cluster, list)
}
