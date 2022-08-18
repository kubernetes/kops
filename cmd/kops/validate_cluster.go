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

	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
		`))

	validateClusterExample = templates.Examples(i18n.T(`
	# Validate the cluster set as the current context of the kube config.
	# Kops will try for 10 minutes to validate the cluster 3 times.
	kops validate cluster --wait 10m --count 3`))

	validateClusterShort = i18n.T(`Validate a kOps cluster.`)
)

type ValidateClusterOptions struct {
	ClusterName string
	output      string
	wait        time.Duration
	count       int
	kubeconfig  string
}

func (o *ValidateClusterOptions) InitDefaults() {
	o.output = OutputTable
}

func NewCmdValidateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ValidateClusterOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:               "cluster [CLUSTER]",
		Short:             validateClusterShort,
		Long:              validateClusterLong,
		Example:           validateClusterExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := RunValidateCluster(context.TODO(), f, out, options)
			if err != nil {
				return fmt.Errorf("Validation failed: %v", err)
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
	cmd.Flags().StringVar(&options.kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")

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

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	list, err := clientSet.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get InstanceGroups for %q: %v", cluster.ObjectMeta.Name, err)
	}

	if options.output == OutputTable {
		fmt.Fprintf(out, "Validating cluster %v\n\n", cluster.ObjectMeta.Name)
	}

	var instanceGroups []kopsapi.InstanceGroup
	for _, ig := range list.Items {
		instanceGroups = append(instanceGroups, ig)
		klog.V(2).Infof("instance group: %#v\n\n", ig.Spec)
	}

	if len(instanceGroups) == 0 {
		return nil, fmt.Errorf("no InstanceGroup objects found")
	}

	// TODO: Refactor into util.Factory
	contextName := cluster.ObjectMeta.Name
	configLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if options.kubeconfig != "" {
		configLoadingRules.ExplicitPath = options.kubeconfig
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{CurrentContext: contextName}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot load kubecfg settings for %q: %v", contextName, err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot build kubernetes api client for %q: %v", contextName, err)
	}

	timeout := time.Now().Add(options.wait)
	pollInterval := 10 * time.Second

	validator, err := validation.NewClusterValidator(cluster, cloud, list, config.Host, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating validatior: %v", err)
	}

	consecutive := 0
	for {
		if options.wait > 0 && time.Now().After(timeout) {
			return nil, fmt.Errorf("wait time exceeded during validation")
		}

		result, err := validator.Validate()
		if err != nil {
			consecutive = 0
			if options.wait > 0 {
				klog.Warningf("(will retry): unexpected error during validation: %v", err)
				time.Sleep(pollInterval)
				continue
			} else {
				return nil, fmt.Errorf("unexpected error during validation: %v", err)
			}
		}

		switch options.output {
		case OutputTable:
			if err := validateClusterOutputTable(result, cluster, instanceGroups, out); err != nil {
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
			return nil, fmt.Errorf("unknown output format: %q", options.output)
		}

		if len(result.Failures) == 0 {
			consecutive++
			if consecutive < options.count {
				klog.Infof("(will retry): cluster passed validation %d consecutive times", consecutive)
				if options.wait > 0 {
					time.Sleep(pollInterval)
					continue
				} else {
					return nil, fmt.Errorf("cluster passed validation %d consecutive times", consecutive)
				}
			} else {
				return result, nil
			}
		} else {
			if options.wait > 0 {
				klog.Warningf("(will retry): cluster not yet healthy")
				consecutive = 0
				time.Sleep(pollInterval)
				continue
			} else {
				return nil, fmt.Errorf("cluster not yet healthy")
			}
		}
	}
}

func validateClusterOutputTable(result *validation.ValidationCluster, cluster *kopsapi.Cluster, instanceGroups []kopsapi.InstanceGroup, out io.Writer) error {
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
		return fmt.Errorf("cannot render nodes for %q: %v", cluster.Name, err)
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
		fmt.Fprintf(out, "\nYour cluster %s is ready\n", cluster.Name)
	}

	return nil
}
