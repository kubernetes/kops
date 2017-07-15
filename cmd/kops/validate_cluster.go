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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	apiutil "k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/util/pkg/tables"
)

type ValidateClusterOptions struct {
	// No options yet
}

func NewCmdValidateCluster(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ValidateClusterOptions{}

	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   validate_short,
		Long:    validate_long,
		Example: validate_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunValidateCluster(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunValidateCluster(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *ValidateClusterOptions) error {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	cluster, err := rootCommand.Cluster()
	if err != nil {
		return err
	}

	clientSet, err := f.Clientset()
	if err != nil {
		return err
	}

	list, err := clientSet.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot get InstanceGroups for %q: %v", cluster.ObjectMeta.Name, err)
	}

	fmt.Fprintf(out, "Validating cluster %v\n\n", cluster.ObjectMeta.Name)

	var instanceGroups []api.InstanceGroup
	for _, ig := range list.Items {
		instanceGroups = append(instanceGroups, ig)
		glog.V(2).Infof("instance group: %#v\n\n", ig.Spec)
	}

	if len(instanceGroups) == 0 {
		return fmt.Errorf("no InstanceGroup objects found\n")
	}

	// TODO: Refactor into util.Factory
	contextName := cluster.ObjectMeta.Name
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: contextName}).ClientConfig()
	if err != nil {
		return fmt.Errorf("Cannot load kubecfg settings for %q: %v\n", contextName, err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("Cannot build kube api client for %q: %v\n", contextName, err)
	}

	validationCluster, validationFailed := validation.ValidateCluster(cluster.ObjectMeta.Name, list, k8sClient)

	if validationCluster == nil || validationCluster.NodeList == nil || validationCluster.NodeList.Items == nil {
		// validationFailed error is already formatted
		return validationFailed
	}

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
	err = t.Render(instanceGroups, out, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "SUBNETS")

	if err != nil {
		return fmt.Errorf("cannot render nodes for %q: %v", cluster.ObjectMeta.Name, err)
	}

	nodeTable := &tables.Table{}

	nodeTable.AddColumn("NAME", func(n v1.Node) string {
		return n.Name
	})

	nodeTable.AddColumn("READY", func(n v1.Node) v1.ConditionStatus {
		return validation.GetNodeConditionStatus(&n)
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
		return fmt.Errorf("cannot render nodes for %q: %v", cluster.ObjectMeta.Name, err)
	}

	if len(validationCluster.ComponentFailures) != 0 {
		componentFailuresTable := &tables.Table{}
		componentFailuresTable.AddColumn("NAME", func(s string) string {
			return s
		})

		fmt.Fprintln(out, "\nComponent Failures")
		err = componentFailuresTable.Render(validationCluster.ComponentFailures, out, "NAME")

		if err != nil {
			return fmt.Errorf("cannot render components for %q: %v", cluster.ObjectMeta.Name, err)
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
			return fmt.Errorf("cannot render pods for %q: %v", cluster.ObjectMeta.Name, err)
		}
	}

	if validationFailed == nil {
		fmt.Fprintf(out, "\nYour cluster %s is ready\n", cluster.ObjectMeta.Name)
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
