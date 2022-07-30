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
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	"k8s.io/klog/v2"

	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/kops/util/pkg/tables"

	"k8s.io/kops/pkg/apis/kops"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

var (
	getInstancesExample = templates.Examples(i18n.T(`
	# Display all instances.
	kops get instances
	`))

	getInstancesShort = i18n.T(`Display cluster instances.`)
)

type renderableCloudInstance struct {
	ID            string   `json:"id"`
	NodeName      string   `json:"nodeName,omitempty"`
	Status        string   `json:"status"`
	Roles         []string `json:"roles"`
	InternalIP    string   `json:"internalIP"`
	InstanceGroup string   `json:"instanceGroup"`
	MachineType   string   `json:"machineType"`
	State         string   `json:"state"`
}

func NewCmdGetInstances(f *util.Factory, out io.Writer, options *GetOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "instances [CLUSTER]",
		Short:             getInstancesShort,
		Example:           getInstancesExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGetInstances(context.TODO(), f, out, options)
		},
	}

	return cmd
}

func RunGetInstances(ctx context.Context, f *util.Factory, out io.Writer, options *GetOptions) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("cluster not found %q", options.ClusterName)
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	k8sClient, err := createK8sClient(cluster)
	if err != nil {
		return err
	}

	nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Warningf("cannot list node names. Kubernetes API unavailable: %v", err)
	}

	igList, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var instanceGroups []*kops.InstanceGroup
	for i := range igList.Items {
		instanceGroups = append(instanceGroups, &igList.Items[i])
	}

	var cloudInstances []*cloudinstances.CloudInstance

	cloudGroups, err := cloud.GetCloudGroups(cluster, instanceGroups, false, nodeList.Items)
	if err != nil {
		return err
	}

	for _, cg := range cloudGroups {
		cloudInstances = append(cloudInstances, cg.Ready...)
		cloudInstances = append(cloudInstances, cg.NeedUpdate...)
		cg.AdjustNeedUpdate()
	}

	switch options.Output {
	case OutputTable:
		return instanceOutputTable(cloudInstances, out)
	case OutputYaml:
		y, err := yaml.Marshal(asRenderable(cloudInstances))
		if err != nil {
			return fmt.Errorf("unable to marshal YAML: %v", err)
		}
		if _, err := out.Write(y); err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
		return nil
	case OutputJSON:
		j, err := json.Marshal(asRenderable(cloudInstances))
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %v", err)
		}
		if _, err := out.Write(j); err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %q", options.Output)
	}
}

func instanceOutputTable(instances []*cloudinstances.CloudInstance, out io.Writer) error {
	fmt.Println("")
	t := &tables.Table{}
	t.AddColumn("ID", func(i *cloudinstances.CloudInstance) string {
		return i.ID
	})
	t.AddColumn("NODE-NAME", func(i *cloudinstances.CloudInstance) string {
		node := i.Node
		if node == nil {
			return ""
		} else {
			return node.Name
		}
	})
	t.AddColumn("STATUS", func(i *cloudinstances.CloudInstance) string {
		return i.Status
	})
	t.AddColumn("ROLES", func(i *cloudinstances.CloudInstance) string {
		return strings.Join(i.Roles, ", ")
	})
	t.AddColumn("INTERNAL-IP", func(i *cloudinstances.CloudInstance) string {
		return i.PrivateIP
	})
	t.AddColumn("INSTANCE-GROUP", func(i *cloudinstances.CloudInstance) string {
		return i.CloudInstanceGroup.HumanName
	})
	t.AddColumn("MACHINE-TYPE", func(i *cloudinstances.CloudInstance) string {
		return i.MachineType
	})
	t.AddColumn("STATE", func(i *cloudinstances.CloudInstance) string {
		return string(i.State)
	})

	columns := []string{"ID", "NODE-NAME", "STATUS", "ROLES", "STATE", "INTERNAL-IP", "INSTANCE-GROUP", "MACHINE-TYPE"}
	return t.Render(instances, out, columns...)
}

func createK8sClient(cluster *kops.Cluster) (*kubernetes.Clientset, error) {
	contextName := cluster.ObjectMeta.Name
	clientGetter := genericclioptions.NewConfigFlags(true)
	clientGetter.Context = &contextName

	config, err := clientGetter.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot load kubecfg settings for %q: %v", contextName, err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot build kubernetes api client for %q: %v", contextName, err)
	}
	return k8sClient, nil
}

func asRenderable(instances []*cloudinstances.CloudInstance) []*renderableCloudInstance {
	arr := make([]*renderableCloudInstance, len(instances))
	for i, ci := range instances {
		arr[i] = &renderableCloudInstance{
			ID:            ci.ID,
			Status:        ci.Status,
			Roles:         ci.Roles,
			InternalIP:    ci.PrivateIP,
			InstanceGroup: ci.CloudInstanceGroup.HumanName,
			MachineType:   ci.MachineType,
			State:         string(ci.State),
		}
		if ci.Node != nil {
			arr[i].NodeName = ci.Node.Name
		}
	}
	return arr
}
