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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	apiutil "k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/util/pkg/tables"
)

func init() {
	if runtime.GOOS == "darwin" {
		// In order for  net.LookupHost(apiAddr.Host) to lookup our placeholder address on darwin, we have to
		os.Setenv("GODEBUG", "netdns=go")
	}
}

type ValidateClusterOptions struct {
	output string
}

func (o *ValidateClusterOptions) InitDefaults() {
	o.output = OutputTable
}

func NewCmdValidateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ValidateClusterOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   validateShort,
		Long:    validateLong,
		Example: validateExample,
		Run: func(cmd *cobra.Command, args []string) {
			result, err := RunValidateCluster(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
			// We want the validate command to exit non-zero if validation found a problem,
			// even if we didn't really hit an error during validation.
			if len(result.PodFailures) != 0 {
				os.Exit(2)
			}
		},
	}

	cmd.Flags().StringVarP(&options.output, "output", "o", options.output, "Ouput format. One of json|yaml|table.")

	return cmd
}

func RunValidateCluster(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *ValidateClusterOptions) (*validation.ValidationCluster, error) {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return nil, err
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return nil, err
	}

	clientSet, err := f.Clientset()
	if err != nil {
		return nil, err
	}

	list, err := clientSet.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get InstanceGroups for %q: %v", cluster.ObjectMeta.Name, err)
	}

	if options.output == OutputTable {
		fmt.Fprintf(out, "Validating cluster %v\n\n", cluster.ObjectMeta.Name)
	}

	var instanceGroups []api.InstanceGroup
	for _, ig := range list.Items {
		instanceGroups = append(instanceGroups, ig)
		glog.V(2).Infof("instance group: %#v\n\n", ig.Spec)
	}

	if len(instanceGroups) == 0 {
		return nil, fmt.Errorf("no InstanceGroup objects found")
	}

	// TODO: Refactor into util.Factory
	contextName := cluster.ObjectMeta.Name
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: contextName}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Cannot load kubecfg settings for %q: %v", contextName, err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Cannot build kubernetes api client for %q: %v", contextName, err)
	}

	// Do not use if we are running gossip
	if !dns.IsGossipHostname(cluster.ObjectMeta.Name) {
		// TODO we may want to return validation.ValidationCluster instead of building it later on
		hasPlaceHolderIPAddress, err := validation.HasPlaceHolderIP(contextName)
		if err != nil {
			return nil, err
		}

		if hasPlaceHolderIPAddress {
			message := "Validation Failed\n\n" +
				"The dns-controller Kubernetes deployment has not updated the Kubernetes cluster's API DNS entry to the correct IP address." +
				"  The API DNS IP address is the placeholder address that kops creates: 203.0.113.123." +
				"  Please wait about 5-10 minutes for a master to start, dns-controller to launch, and DNS to propagate." +
				"  The protokube container and dns-controller deployment logs may contain more diagnostic information." +
				"  Etcd and the API DNS entries must be updated for a kops Kubernetes cluster to start."
			validationCluster := &validation.ValidationCluster{
				ClusterName:  cluster.ObjectMeta.Name,
				ErrorMessage: message,
				Status:       validation.ClusterValidationFailed,
			}
			validationFailed := fmt.Errorf("\nCannot reach cluster's API server: unable to Validate Cluster: %s", cluster.ObjectMeta.Name)
			switch options.output {
			case OutputTable:
				fmt.Println(message)
				return validationCluster, validationFailed
			case OutputYaml:
				if err := validateClusterOutputYAML(validationCluster, validationFailed, out); err != nil {
					return nil, err
				}
			case OutputJSON:
				if err := validateClusterOutputJSON(validationCluster, validationFailed, out); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("Unknown output format: %q", options.output)
			}

			return validationCluster, validationFailed
		}
	}

	validationCluster, validationFailed := validation.ValidateCluster(cluster, list, k8sClient)

	if validationCluster == nil || validationCluster.NodeList == nil || validationCluster.NodeList.Items == nil {
		return validationCluster, validationFailed
	}

	switch options.output {
	case OutputTable:
		if err := validateClusterOutputTable(validationCluster, validationFailed, instanceGroups, out); err != nil {
			return nil, err
		}
	case OutputYaml:
		if err := validateClusterOutputYAML(validationCluster, validationFailed, out); err != nil {
			return nil, err
		}
	case OutputJSON:
		if err := validateClusterOutputJSON(validationCluster, validationFailed, out); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unknown output format: %q", options.output)
	}

	return validationCluster, validationFailed
}

func validateClusterOutputTable(validationCluster *validation.ValidationCluster, validationFailed error, instanceGroups []api.InstanceGroup, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("NAME", func(c api.InstanceGroup) string {
		return c.ObjectMeta.Name
	})
	t.AddColumn("ROLE", func(c api.InstanceGroup) string {
		return string(c.Spec.Role)
	})
	t.AddColumn("MACHINETYPE", func(c api.InstanceGroup) string {
		return c.Spec.MachineType
	})
	t.AddColumn("SUBNETS", func(c api.InstanceGroup) string {
		return strings.Join(c.Spec.Subnets, ",")
	})
	t.AddColumn("MIN", func(c api.InstanceGroup) string {
		return int32PointerToString(c.Spec.MinSize)
	})
	t.AddColumn("MAX", func(c api.InstanceGroup) string {
		return int32PointerToString(c.Spec.MaxSize)
	})

	fmt.Fprintln(out, "INSTANCE GROUPS")
	err := t.Render(instanceGroups, out, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "SUBNETS")

	if err != nil {
		return fmt.Errorf("cannot render nodes for %q: %v", validationCluster.ClusterName, err)
	}

	nodeTable := &tables.Table{}

	nodeTable.AddColumn("NAME", func(n v1.Node) string {
		return n.Name
	})

	nodeTable.AddColumn("READY", func(n v1.Node) v1.ConditionStatus {
		return validation.GetNodeReadyStatus(&n)
	})

	nodeTable.AddColumn("ROLE", func(n v1.Node) string {
		// TODO: Maybe print the instance group role instead?
		// TODO: Maybe include the instance group name?
		role := apiutil.GetNodeRole(&n)
		if role == "" {
			role = "node"
		}
		return role
	})

	fmt.Fprintln(out, "\nNODE STATUS")
	err = nodeTable.Render(validationCluster.NodeList.Items, out, "NAME", "ROLE", "READY")

	if err != nil {
		return fmt.Errorf("cannot render nodes for %q: %v", validationCluster.ClusterName, err)
	}

	if len(validationCluster.ComponentFailures) != 0 {
		componentFailuresTable := &tables.Table{}
		componentFailuresTable.AddColumn("NAME", func(s string) string {
			return s
		})

		fmt.Fprintln(out, "\nComponent Failures")
		err = componentFailuresTable.Render(validationCluster.ComponentFailures, out, "NAME")

		if err != nil {
			return fmt.Errorf("cannot render components for %q: %v", validationCluster.ClusterName, err)
		}
	}

	if len(validationCluster.PodFailures) != 0 {
		podFailuresTable := &tables.Table{}
		podFailuresTable.AddColumn("NAME", func(s string) string {
			return s
		})

		fmt.Fprintln(out, "\nPod Failures in kube-system")
		err = podFailuresTable.Render(validationCluster.PodFailures, out, "NAME")

		if err != nil {
			return fmt.Errorf("cannot render pods for %q: %v", validationCluster.ClusterName, err)
		}
	}

	if validationFailed == nil {
		fmt.Fprintf(out, "\nYour cluster %s is ready\n", validationCluster.ClusterName)
		return nil
	} else {
		// do we need to print which instance group is not ready?
		// nodes are going to be a pain
		fmt.Fprint(out, "\nValidation Failed\n")
		fmt.Fprintf(out, "Ready Master(s) %d out of %d.\n", len(validationCluster.MastersReadyArray), validationCluster.MastersCount)
		fmt.Fprintf(out, "Ready Node(s) %d out of %d.\n", len(validationCluster.NodesReadyArray), validationCluster.NodesCount)
		return validationFailed
	}
}

func validateClusterOutputYAML(validationCluster *validation.ValidationCluster, validationFailed error, out io.Writer) error {
	y, err := yaml.Marshal(validationCluster)
	if err != nil {
		return fmt.Errorf("unable to marshall YAML: %v\n", err)
	}
	return validateOutput(y, validationFailed, out)
}

func validateClusterOutputJSON(validationCluster *validation.ValidationCluster, validationFailed error, out io.Writer) error {
	j, err := json.Marshal(validationCluster)
	if err != nil {
		return fmt.Errorf("unable to marshall JSON: %v\n", err)
	}
	return validateOutput(j, validationFailed, out)
}

func validateOutput(b []byte, validationFailed error, out io.Writer) error {
	if _, err := out.Write(b); err != nil {
		return fmt.Errorf("unable to print data: %v\n", err)
	}
	return validationFailed
}
