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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/util/pkg/tables"
	"sigs.k8s.io/yaml"
)

var (
	validateClusterLong = templates.LongDesc(i18n.T(`
		This commands validates the following components:
	
		1. All control plane nodes are running and have "Ready" status.
		2. All worker nodes are running and have "Ready" status.
		3. All control plane nodes have the expected pods.
		4. All pods with a critical priority are running and have "Ready" status.

		When --instance-group or --instance-group-roles is specified, only the matching
		instance groups, their nodes, and the critical pods scheduled on those nodes
		are validated. Nodes and pods outside those instance groups cannot fail the
		validation.
		`))

	validateClusterExample = templates.Examples(i18n.T(`
	# Validate the cluster set as the current context of the kube config.
	# Kops will try for 10 minutes to validate the cluster 3 times.
	kops validate cluster --wait 10m --count 3

	# Validate only the "addons" instance group, waiting up to 10 minutes.
	kops validate cluster --instance-group addons --wait 10m

	# Validate all control-plane instance groups.
	kops validate cluster --instance-group-roles control-plane --wait 10m`))

	validateClusterShort = i18n.T(`Validate a kOps cluster.`)
)

type ValidateClusterOptions struct {
	ClusterName        string
	InstanceGroups     []string
	InstanceGroupRoles []string
	output             string
	wait               time.Duration
	count              int
	interval           time.Duration
	kubeconfig         string

	MaxUnreadyNodes int

	// filterInstanceGroups is a function that returns true if the instance group should be validated
	filterInstanceGroups func(ig *kopsapi.InstanceGroup) bool

	// filterPodsForValidation is a function that returns true if the pod should be validated
	filterPodsForValidation func(pod *v1.Pod) bool

	ExportKubeconfigOptions
}

func (o *ValidateClusterOptions) InitDefaults() {
	o.output = OutputTable
	o.interval = 10 * time.Second
	o.MaxUnreadyNodes = 0
}

func NewCmdValidateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ValidateClusterOptions{}
	options.InitDefaults()

	allRoles := make([]string, 0, len(kopsapi.AllInstanceGroupRoles))
	for _, r := range kopsapi.AllInstanceGroupRoles {
		allRoles = append(allRoles, r.ToLowerString())
	}

	cmd := &cobra.Command{
		Use:               "cluster [CLUSTER]",
		Short:             validateClusterShort,
		Long:              validateClusterLong,
		Example:           validateClusterExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := RunValidateCluster(cmd.Context(), f, out, options)
			if err != nil {
				return fmt.Errorf("validation failed: %v", err)
			}

			// We want the validate command to exit non-zero if validation found a problem,
			// even if we didn't really hit an error during validation.
			if len(result.Failures) != 0 {
				os.Exit(2)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&options.output, "output", "o", options.output, "Output format. One of json|yaml|table.")
	cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml", "table"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.Flags().DurationVar(&options.wait, "wait", options.wait, "Amount of time to wait for the cluster to become ready")
	cmd.Flags().IntVar(&options.count, "count", options.count, "Number of consecutive successful validations required")
	cmd.Flags().DurationVar(&options.interval, "interval", options.interval, "Time in duration to wait between validation attempts")
	cmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
	cmd.Flags().IntVar(&options.MaxUnreadyNodes, "max-unready-nodes", options.MaxUnreadyNodes, "The maximum number of non-ready worker nodes tolerated during validation")
	cmd.Flags().StringSliceVar(&options.InstanceGroups, "instance-group", options.InstanceGroups, "Instance groups to validate (defaults to all if not specified)")
	cmd.RegisterFlagCompletionFunc("instance-group", completeInstanceGroup(f, &options.InstanceGroups, &options.InstanceGroupRoles))
	cmd.Flags().StringSliceVar(&options.InstanceGroupRoles, "instance-group-roles", options.InstanceGroupRoles, "Instance group roles to validate ("+strings.Join(allRoles, ",")+")")
	cmd.RegisterFlagCompletionFunc("instance-group-roles", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sets.List(sets.New(allRoles...).Delete(options.InstanceGroupRoles...)), cobra.ShellCompDirectiveNoFileComp
	})

	options.CreateKubecfgOptions.AddCommonFlags(cmd.Flags())

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

func RunValidateCluster(ctx context.Context, f *util.Factory, out io.Writer, options *ValidateClusterOptions) (*validation.ValidationCluster, error) {
	clientSet, err := f.KopsClient()
	if err != nil {
		return nil, err
	}

	cluster, err := GetCluster(ctx, f, options.ClusterName)
	if err != nil {
		return nil, err
	}

	list, err := clientSet.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get InstanceGroups for %q: %v", cluster.ObjectMeta.Name, err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no InstanceGroup objects found")
	}

	for _, ig := range list.Items {
		klog.V(2).Infof("instance group: %#v\n\n", ig.Spec)
	}

	scoped := len(options.InstanceGroups) != 0 || len(options.InstanceGroupRoles) != 0

	var instanceGroups []kopsapi.InstanceGroup
	var selectedInstanceGroups []*kopsapi.InstanceGroup
	subject := "cluster"
	readyMessage := fmt.Sprintf("Your cluster %s is ready", cluster.Name)

	if scoped {
		selectedInstanceGroups, err = selectInstanceGroupsToValidate(list.Items, options.InstanceGroups, options.InstanceGroupRoles)
		if err != nil {
			return nil, err
		}

		names := make([]string, 0, len(selectedInstanceGroups))
		for _, ig := range selectedInstanceGroups {
			names = append(names, ig.ObjectMeta.Name)
			instanceGroups = append(instanceGroups, *ig)
		}

		subject = fmt.Sprintf("instance group %s", names[0])
		readyMessage = fmt.Sprintf("Your %s is ready", subject)
		if len(names) > 1 {
			subject = fmt.Sprintf("instance groups %s", strings.Join(names, ", "))
			readyMessage = fmt.Sprintf("Your %s are ready", subject)
		}

		if options.output == OutputTable {
			fmt.Fprintf(out, "Validating %s in cluster %v\n\n", subject, cluster.ObjectMeta.Name)
		}
	} else {
		instanceGroups = append(instanceGroups, list.Items...)
		if options.output == OutputTable {
			fmt.Fprintf(out, "Validating cluster %v\n\n", cluster.ObjectMeta.Name)
		}
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	restConfig, k8sClient, err := buildValidationClient(ctx, f, cluster, options.CreateKubecfgOptions)
	if err != nil {
		return nil, err
	}

	var validator validation.ClusterValidator
	if scoped {
		validator, err = validation.NewInstanceGroupValidator(cluster, cloud, selectedInstanceGroups, options.MaxUnreadyNodes, restConfig, k8sClient)
	} else {
		validator, err = validation.NewClusterValidator(cluster, cloud, list, options.filterInstanceGroups, options.filterPodsForValidation, options.MaxUnreadyNodes, restConfig, k8sClient)
	}
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating validator: %v", err)
	}

	loop := &validationLoop{
		subject:  subject,
		output:   options.output,
		wait:     options.wait,
		count:    options.count,
		interval: options.interval,
		outputTable: func(result *validation.ValidationCluster) error {
			return validateClusterOutputTable(result, cluster, instanceGroups, readyMessage, out)
		},
	}
	return loop.run(ctx, validator, out)
}

// selectInstanceGroupsToValidate resolves --instance-group and --instance-group-roles
// into a deduplicated slice of instance groups. A repeated group would make cloud
// providers match a single cloud group to multiple instance groups.
func selectInstanceGroupsToValidate(items []kopsapi.InstanceGroup, instanceGroupNames []string, instanceGroupRoles []string) ([]*kopsapi.InstanceGroup, error) {
	byName := make(map[string]*kopsapi.InstanceGroup, len(items))
	for i := range items {
		ig := &items[i]
		byName[ig.ObjectMeta.Name] = ig
	}

	var selected []*kopsapi.InstanceGroup
	seen := make(map[string]bool)

	for _, name := range instanceGroupNames {
		ig := byName[name]
		if ig == nil {
			return nil, fmt.Errorf("InstanceGroup %q not found", name)
		}
		if !seen[name] {
			seen[name] = true
			selected = append(selected, ig)
		}
	}

	if len(instanceGroupRoles) != 0 {
		roles := make(map[kopsapi.InstanceGroupRole]bool, len(instanceGroupRoles))
		for _, role := range instanceGroupRoles {
			parsed, ok := kopsapi.ParseInstanceGroupRole(role, true)
			if !ok {
				return nil, fmt.Errorf("invalid instance group role %q", role)
			}
			roles[parsed] = true
		}

		for i := range items {
			ig := &items[i]
			if roles[ig.Spec.Role] && !seen[ig.ObjectMeta.Name] {
				seen[ig.ObjectMeta.Name] = true
				selected = append(selected, ig)
			}
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no matching InstanceGroup objects found")
	}

	return selected, nil
}

// buildValidationClient builds the client used to reach the API server of the cluster being validated.
func buildValidationClient(ctx context.Context, f *util.Factory, cluster *kopsapi.Cluster, options kubeconfig.CreateKubecfgOptions) (*rest.Config, kubernetes.Interface, error) {
	restConfig, err := f.RESTConfig(ctx, cluster, options)
	if err != nil {
		return nil, nil, fmt.Errorf("getting rest config: %w", err)
	}

	httpClient, err := f.HTTPClient(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("getting http client: %w", err)
	}

	k8sClient, err := kubernetes.NewForConfigAndClient(restConfig, httpClient)
	if err != nil {
		return nil, nil, fmt.Errorf("building kubernetes client: %w", err)
	}

	return restConfig, k8sClient, nil
}

// validationLoop validates repeatedly, until the required number of consecutive
// validations pass or the wait time is exceeded.
type validationLoop struct {
	// subject is what is being validated, as named in the messages to the user.
	subject  string
	output   string
	wait     time.Duration
	count    int
	interval time.Duration

	// outputTable writes a validation result in the table output format.
	outputTable func(result *validation.ValidationCluster) error
}

func (l *validationLoop) run(ctx context.Context, validator validation.ClusterValidator, out io.Writer) (*validation.ValidationCluster, error) {
	timeout := time.Now().Add(l.wait)

	consecutive := 0
	for {
		if l.wait > 0 && time.Now().After(timeout) && consecutive == 0 {
			return nil, fmt.Errorf("wait time exceeded during validation")
		}

		result, err := validator.Validate(ctx)
		if err != nil {
			consecutive = 0
			if l.wait > 0 {
				klog.Warningf("(will retry): unexpected error during validation: %v", err)
				time.Sleep(l.interval)
				continue
			} else {
				return nil, fmt.Errorf("unexpected error during validation: %v", err)
			}
		}

		switch l.output {
		case OutputTable:
			if err := l.outputTable(result); err != nil {
				return nil, err
			}
		case OutputYaml:
			y, err := yaml.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal YAML: %v", err)
			}
			if _, err := out.Write(y); err != nil {
				return nil, fmt.Errorf("error writing to output: %v", err)
			}
		case OutputJSON:
			j, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("unable to marshal JSON: %v", err)
			}
			if _, err := out.Write(j); err != nil {
				return nil, fmt.Errorf("error writing to output: %v", err)
			}
		default:
			return nil, fmt.Errorf("unknown output format: %q", l.output)
		}

		if len(result.Failures) == 0 {
			consecutive++
			if consecutive < l.count {
				klog.Infof("(will retry): %s passed validation %d consecutive times", l.subject, consecutive)
				if l.wait > 0 {
					time.Sleep(l.interval)
					continue
				} else {
					return nil, fmt.Errorf("%s passed validation %d consecutive times", l.subject, consecutive)
				}
			} else {
				return result, nil
			}
		} else {
			if l.wait > 0 {
				klog.Warningf("(will retry): %s not yet healthy", l.subject)
				consecutive = 0
				time.Sleep(l.interval)
				continue
			} else {
				return nil, fmt.Errorf("%s not yet healthy", l.subject)
			}
		}
	}
}

func validateClusterOutputTable(result *validation.ValidationCluster, cluster *kopsapi.Cluster, instanceGroups []kopsapi.InstanceGroup, readyMessage string, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("NAME", func(c kopsapi.InstanceGroup) string {
		return c.ObjectMeta.Name
	})
	t.AddColumn("ROLE", func(c kopsapi.InstanceGroup) string {
		return string(c.Spec.Role)
	})
	t.AddColumn("MACHINETYPE", func(c kopsapi.InstanceGroup) string {
		return c.Spec.MachineType
	})
	t.AddColumn("SUBNETS", func(c kopsapi.InstanceGroup) string {
		return strings.Join(c.Spec.Subnets, ",")
	})
	t.AddColumn("MIN", func(c kopsapi.InstanceGroup) string {
		return int32PointerToString(c.Spec.MinSize)
	})
	t.AddColumn("MAX", func(c kopsapi.InstanceGroup) string {
		return int32PointerToString(c.Spec.MaxSize)
	})

	fmt.Fprintln(out, "INSTANCE GROUPS")
	err := t.Render(instanceGroups, out, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "SUBNETS")
	if err != nil {
		return fmt.Errorf("cannot render instance groups for %q: %w", cluster.Name, err)
	}

	{
		nodeTable := &tables.Table{}
		nodeTable.AddColumn("NAME", func(n *validation.ValidationNode) string {
			return n.Name
		})

		nodeTable.AddColumn("READY", func(n *validation.ValidationNode) v1.ConditionStatus {
			return n.Status
		})

		nodeTable.AddColumn("ROLE", func(n *validation.ValidationNode) string {
			return n.Role
		})

		fmt.Fprintln(out, "\nNODE STATUS")
		if err := nodeTable.Render(result.Nodes, out, "NAME", "ROLE", "READY"); err != nil {
			return fmt.Errorf("cannot render nodes for %q: %v", cluster.Name, err)
		}
	}

	if len(result.Failures) != 0 {
		failuresTable := &tables.Table{}
		failuresTable.AddColumn("KIND", func(e *validation.ValidationError) string {
			return e.Kind
		})
		failuresTable.AddColumn("NAME", func(e *validation.ValidationError) string {
			return e.Name
		})
		failuresTable.AddColumn("MESSAGE", func(e *validation.ValidationError) string {
			return e.Message
		})

		fmt.Fprintln(out, "\nVALIDATION ERRORS")
		if err := failuresTable.Render(result.Failures, out, "KIND", "NAME", "MESSAGE"); err != nil {
			return fmt.Errorf("error rendering failures table: %v", err)
		}

		fmt.Fprintf(out, "\nValidation Failed\n")
	} else {
		fmt.Fprintf(out, "\n%s\n", readyMessage)
	}

	return nil
}
