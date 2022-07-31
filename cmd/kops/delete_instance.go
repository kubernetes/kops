/*
Copyright 2020 The Kubernetes Authors.

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
	"time"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/instancegroups"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// DeleteInstanceOptions is the command Object for an instance deletion.
type DeleteInstanceOptions struct {
	Yes       bool
	CloudOnly bool

	// The following two variables are when kOps is validating a cluster
	// between detach and deletion.

	// FailOnDrainError fail deletion if drain errors.
	FailOnDrainError bool

	// FailOnValidate fail the deletion when the cluster
	// does not validate, after a validation period.
	FailOnValidate bool

	// PostDrainDelay is the duration of a pause after a drain operation
	PostDrainDelay time.Duration

	// ValidationTimeout is the timeout for validation to succeed after the drain and pause
	ValidationTimeout time.Duration

	// ValidateCount is the amount of time that a cluster needs to be validated between drain and deletion
	ValidateCount int32

	ClusterName string

	InstanceID string

	Surge bool
}

func (o *DeleteInstanceOptions) initDefaults() {
	d := &RollingUpdateOptions{}
	d.InitDefaults()

	o.CloudOnly = false
	o.FailOnDrainError = false
	o.FailOnValidate = true

	o.PostDrainDelay = d.PostDrainDelay
	o.ValidationTimeout = d.ValidationTimeout
	o.ValidateCount = d.ValidateCount

	o.Surge = true
}

func NewCmdDeleteInstance(f *util.Factory, out io.Writer) *cobra.Command {
	deleteInstanceLong := templates.LongDesc(i18n.T(`
		Delete an instance. By default, it will detach the instance from 
		the instance group, drain it, then terminate it.`))

	deleteInstanceExample := templates.Examples(i18n.T(`
		# Delete an instance from the currently active cluster.
		kops delete instance i-0a5ed581b862d3425 --yes

		# Delete an instance from the currently active cluster using node name.
		kops delete instance ip-xx.xx.xx.xx.ec2.internal --yes

		# Delete an instance from the currently active cluster without
		validation or draining.
		kops delete instance --cloudonly i-0a5ed581b862d3425 --yes
		`))

	deleteInstanceShort := i18n.T(`Delete an instance.`)

	var options DeleteInstanceOptions
	options.initDefaults()

	cmd := &cobra.Command{
		Use:     "instance INSTANCE|NODE",
		Short:   deleteInstanceShort,
		Long:    deleteInstanceLong,
		Example: deleteInstanceExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)
			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify ID of instance or name of node to delete")
			}
			options.InstanceID = args[0]

			if len(args) != 1 {
				return fmt.Errorf("can only delete one instance at a time")
			}

			return nil
		},
		ValidArgsFunction: completeInstanceOrNode(f, &options),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDeleteInstance(context.TODO(), f, out, &options)
		},
	}

	cmd.Flags().BoolVar(&options.CloudOnly, "cloudonly", options.CloudOnly, "Perform deletion update without confirming progress with Kubernetes")
	cmd.Flags().BoolVar(&options.Surge, "surge", options.Surge, "Surge by detaching the node from the ASG before deletion")

	cmd.Flags().DurationVar(&options.ValidationTimeout, "validation-timeout", options.ValidationTimeout, "Maximum time to wait for a cluster to validate")
	cmd.Flags().Int32Var(&options.ValidateCount, "validate-count", options.ValidateCount, "Number of times that a cluster needs to be validated after single node update")
	cmd.Flags().DurationVar(&options.PostDrainDelay, "post-drain-delay", options.PostDrainDelay, "Time to wait after draining each node")

	cmd.Flags().BoolVar(&options.FailOnDrainError, "fail-on-drain-error", true, "Fail if draining a node fails")
	cmd.Flags().BoolVar(&options.FailOnValidate, "fail-on-validate-error", true, "Fail if the cluster fails to validate")

	cmd.Flags().BoolVarP(&options.Yes, "yes", "y", options.Yes, "Specify --yes to immediately delete the instance")

	return cmd
}

func RunDeleteInstance(ctx context.Context, f *util.Factory, out io.Writer, options *DeleteInstanceOptions) error {
	clientSet, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return err
	}

	var nodes []v1.Node
	var k8sClient kubernetes.Interface
	var host string
	if !options.CloudOnly {
		k8sClient, host, nodes, err = getNodes(ctx, cluster, true)
		if err != nil {
			return err
		}
	}

	list, err := clientSet.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var instanceGroups []*kopsapi.InstanceGroup
	for i := range list.Items {
		instanceGroups = append(instanceGroups, &list.Items[i])
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	groups, err := cloud.GetCloudGroups(cluster, instanceGroups, false, nodes)
	if err != nil {
		return err
	}

	cloudMember := findDeletionNode(groups, options)

	if cloudMember == nil {
		return fmt.Errorf("could not find instance %v", options.InstanceID)
	}

	if options.CloudOnly {
		fmt.Fprintf(out, "Instance %v found for deletion\n", cloudMember.ID)
	} else {
		if cloudMember.Node != nil {
			fmt.Fprintf(out, "Instance %v (%v) found for deletion\n", cloudMember.ID, cloudMember.Node.Name)
		} else {
			fmt.Fprintf(os.Stderr, "Instance is not a member of the cluster\n")
			fmt.Fprintf(os.Stderr, "Use --cloudonly to do a deletion without confirming progress with the k8s API\n\n")
			return fmt.Errorf("error finding node name for instance: %v", cloudMember.ID)
		}
	}

	if !options.Yes {
		fmt.Fprintf(out, "\nMust specify --yes to delete instance\n")
		return nil
	}

	d := &instancegroups.RollingUpdateCluster{
		Clientset:         clientSet,
		Cluster:           cluster,
		Ctx:               ctx,
		MasterInterval:    0,
		NodeInterval:      0,
		BastionInterval:   0,
		Interactive:       false,
		Force:             true,
		Cloud:             cloud,
		K8sClient:         k8sClient,
		FailOnDrainError:  options.FailOnDrainError,
		FailOnValidate:    options.FailOnValidate,
		CloudOnly:         options.CloudOnly,
		ClusterName:       options.ClusterName,
		PostDrainDelay:    options.PostDrainDelay,
		ValidationTimeout: options.ValidationTimeout,
		ValidateCount:     int(options.ValidateCount),
		// TODO should we expose this to the UI?
		ValidateTickDuration:    30 * time.Second,
		ValidateSuccessDuration: 10 * time.Second,
	}

	var clusterValidator validation.ClusterValidator
	if !options.CloudOnly {
		clusterValidator, err = validation.NewClusterValidator(cluster, cloud, list, host, k8sClient)
		if err != nil {
			return fmt.Errorf("cannot create cluster validator: %v", err)
		}
	}
	d.ClusterValidator = clusterValidator

	return d.UpdateSingleInstance(cloudMember, options.Surge)
}

func getNodes(ctx context.Context, cluster *kopsapi.Cluster, verbose bool) (kubernetes.Interface, string, []v1.Node, error) {
	var nodes []v1.Node
	var k8sClient kubernetes.Interface

	contextName := cluster.ObjectMeta.Name
	clientGetter := genericclioptions.NewConfigFlags(true)
	clientGetter.Context = &contextName

	config, err := clientGetter.ToRESTConfig()
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot load kubecfg settings for %q: %v", contextName, err)
	}

	k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", nil, fmt.Errorf("cannot build kube client for %q: %v", contextName, err)
	}

	nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Unable to reach the kubernetes API.\n")
			fmt.Fprintf(os.Stderr, "Use --cloudonly to do a deletion without confirming progress with the k8s API\n\n")
		}
		return nil, "", nil, fmt.Errorf("listing nodes in cluster: %v", err)
	}

	if nodeList != nil {
		nodes = nodeList.Items
	}
	return k8sClient, config.Host, nodes, nil
}

func deleteNodeMatch(cloudMember *cloudinstances.CloudInstance, options *DeleteInstanceOptions) bool {
	return cloudMember.ID == options.InstanceID ||
		(!options.CloudOnly && cloudMember.Node != nil && cloudMember.Node.Name == options.InstanceID)
}

func findDeletionNode(groups map[string]*cloudinstances.CloudInstanceGroup, options *DeleteInstanceOptions) *cloudinstances.CloudInstance {
	for _, group := range groups {
		for _, r := range group.Ready {
			if deleteNodeMatch(r, options) {
				return r
			}
		}
		for _, r := range group.NeedUpdate {
			if deleteNodeMatch(r, options) {
				return r
			}
		}
	}
	return nil
}

func completeInstanceOrNode(f commandutils.Factory, options *DeleteInstanceOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		commandutils.ConfigureKlogForCompletion()
		ctx := context.TODO()

		cluster, clientSet, completions, directive := GetClusterForCompletion(ctx, f, nil)
		if cluster == nil {
			return completions, directive
		}

		var nodes []v1.Node
		var err error
		if !options.CloudOnly {
			_, _, nodes, err = getNodes(ctx, cluster, false)
			if err != nil {
				cobra.CompErrorln(err.Error())
			}
		}

		list, err := clientSet.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
		if err != nil {
			return commandutils.CompletionError("listing instance groups", err)
		}

		var instanceGroups []*kopsapi.InstanceGroup
		for i := range list.Items {
			instanceGroups = append(instanceGroups, &list.Items[i])
		}

		cloud, err := cloudup.BuildCloud(cluster)
		if err != nil {
			return commandutils.CompletionError("initializing cloud", err)
		}

		groups, err := cloud.GetCloudGroups(cluster, instanceGroups, false, nodes)
		if err != nil {
			return commandutils.CompletionError("listing instances", err)
		}

		completions = nil
		longestGroup := 0
		for _, group := range groups {
			if group.InstanceGroup != nil && longestGroup < len(group.InstanceGroup.Name) {
				longestGroup = len(group.InstanceGroup.Name)
			}
		}
		for _, group := range groups {
			for _, instance := range group.Ready {
				completions = appendInstance(completions, instance, longestGroup)
			}
			for _, instance := range group.NeedUpdate {
				completions = appendInstance(completions, instance, longestGroup)
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func appendInstance(completions []string, instance *cloudinstances.CloudInstance, longestGroup int) []string {
	completion := instance.ID
	if instance.CloudInstanceGroup.InstanceGroup != nil {
		completion += "\t" + instance.CloudInstanceGroup.InstanceGroup.Name

		if instance.Node != nil {
			padding := strings.Repeat(" ", longestGroup+1-len(instance.CloudInstanceGroup.InstanceGroup.Name))
			completion += padding + instance.Node.Name
			completions = append(completions, instance.Node.Name+"\t"+instance.CloudInstanceGroup.InstanceGroup.Name+padding+instance.ID)
		}
	}
	return append(completions, completion)
}
