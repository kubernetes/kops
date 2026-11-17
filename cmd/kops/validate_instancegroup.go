/*
Copyright 2026 The Kubernetes Authors.

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
	"strings"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	validateInstanceGroupLong = templates.LongDesc(i18n.T(`
		This command validates only the specified instance groups. Unlike
		"kops validate cluster", nodes and pods that do not belong to these
		instance groups cannot fail the validation. This makes the command
		suitable for waiting on an instance group in a very large cluster, where
		there is often an unrelated node or pod that is not ready.

		The following is validated:

		1. Each instance group has the expected number of instances.
		2. All of its instances have joined the cluster and have "Ready" status.
		3. All of its control plane nodes have the expected pods.
		4. All pods with a critical priority running on its nodes have "Ready" status.
		`))

	validateInstanceGroupExample = templates.Examples(i18n.T(`
	# Wait up to 10 minutes for the instance group "addons" to become ready.
	kops validate instancegroup --name k8s-cluster.example.com addons --wait 10m

	# Validate two instance groups, requiring 3 consecutive successful validations.
	kops validate instancegroup --name k8s-cluster.example.com nodes-us-east-1a nodes-us-east-1b --wait 10m --count 3`))

	validateInstanceGroupShort = i18n.T(`Validate instance groups of a kOps cluster.`)
)

type ValidateInstanceGroupOptions struct {
	ClusterName        string
	InstanceGroupNames []string

	output   string
	wait     time.Duration
	count    int
	interval time.Duration

	MaxUnreadyNodes int

	kubeconfig.CreateKubecfgOptions
}

func (o *ValidateInstanceGroupOptions) InitDefaults() {
	o.output = OutputTable
	o.interval = 10 * time.Second
	o.MaxUnreadyNodes = 0
}

func NewCmdValidateInstanceGroup(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ValidateInstanceGroupOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "instancegroup INSTANCE_GROUP...",
		Aliases: []string{"instancegroups", "ig"},
		Short:   validateInstanceGroupShort,
		Long:    validateInstanceGroupLong,
		Example: validateInstanceGroupExample,
		Args: func(cmd *cobra.Command, args []string) error {
			options.ClusterName = rootCommand.ClusterName(true)
			if options.ClusterName == "" {
				return fmt.Errorf("--name is required")
			}

			if len(args) == 0 {
				return fmt.Errorf("must specify the name of at least one instance group to validate")
			}

			options.InstanceGroupNames = args
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return completeInstanceGroup(f, &args, nil)(cmd, nil, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := RunValidateInstanceGroup(cmd.Context(), f, out, options); err != nil {
				return fmt.Errorf("validation failed: %v", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&options.output, "output", "o", options.output, "Output format. One of json|yaml|table.")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml", "table"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().DurationVar(&options.wait, "wait", options.wait, "Amount of time to wait for the instance groups to become ready")
	cmd.Flags().IntVar(&options.count, "count", options.count, "Number of consecutive successful validations required")
	cmd.Flags().DurationVar(&options.interval, "interval", options.interval, "Time in duration to wait between validation attempts")
	cmd.Flags().IntVar(&options.MaxUnreadyNodes, "max-unready-nodes", options.MaxUnreadyNodes, "The maximum number of non-ready worker nodes tolerated during validation")

	options.CreateKubecfgOptions.AddCommonFlags(cmd.Flags())

	return cmd
}

func RunValidateInstanceGroup(ctx context.Context, f *util.Factory, out io.Writer, options *ValidateInstanceGroupOptions) (*validation.ValidationCluster, error) {
	clientSet, err := f.KopsClient()
	if err != nil {
		return nil, err
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return nil, err
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	list, err := clientSet.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get InstanceGroups for %q: %v", cluster.ObjectMeta.Name, err)
	}

	// A repeated name would make the cloud providers match one cloud group to several instance groups.
	var names []string
	seen := make(map[string]bool)
	for _, name := range options.InstanceGroupNames {
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("must specify the name of at least one instance group to validate")
	}

	instanceGroups, err := filterInstanceGroupsByName(names, list.Items)
	if err != nil {
		return nil, err
	}

	subject := fmt.Sprintf("instance group %s", names[0])
	readyMessage := fmt.Sprintf("Your %s is ready", subject)
	if len(names) > 1 {
		subject = fmt.Sprintf("instance groups %s", strings.Join(names, ", "))
		readyMessage = fmt.Sprintf("Your %s are ready", subject)
	}

	if options.output == OutputTable {
		fmt.Fprintf(out, "Validating %s in cluster %v\n\n", subject, cluster.ObjectMeta.Name)
	}

	restConfig, k8sClient, err := buildValidationClient(ctx, f, cluster, options.CreateKubecfgOptions)
	if err != nil {
		return nil, err
	}

	validator, err := validation.NewInstanceGroupValidator(cluster, cloud, instanceGroups, options.MaxUnreadyNodes, restConfig, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating validator: %v", err)
	}

	var tableInstanceGroups []kopsapi.InstanceGroup
	for _, ig := range instanceGroups {
		tableInstanceGroups = append(tableInstanceGroups, *ig)
	}

	loop := &validationLoop{
		subject:  subject,
		output:   options.output,
		wait:     options.wait,
		count:    options.count,
		interval: options.interval,
		outputTable: func(result *validation.ValidationCluster) error {
			return validateClusterOutputTable(result, cluster, tableInstanceGroups, readyMessage, out)
		},
	}
	return loop.run(ctx, validator, out)
}
