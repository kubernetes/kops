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

	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"k8s.io/kops/upup/pkg/fi"

	"k8s.io/kops/pkg/client/simple"

	"k8s.io/kops/pkg/resources/openstack"

	"k8s.io/klog/v2"

	"k8s.io/client-go/kubernetes"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/util/pkg/tables"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resources/aws"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi/cloudup"

	osCloudup "k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

func NewCmdGetInstances(f *util.Factory, out io.Writer, options *GetOptions) *cobra.Command {
	getInstancesShort := i18n.T(`Display cluster instances.`)

	getInstancesLong := templates.LongDesc(i18n.T(`
	Display cluster instances.`))

	getInstancesExample := templates.Examples(i18n.T(`
	# Display all instances.
	kops get instances
	`))

	cmd := &cobra.Command{
		Use:     "instances",
		Short:   getInstancesShort,
		Long:    getInstancesLong,
		Example: getInstancesExample,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.TODO()

			if err := rootCommand.ProcessArgs(args); err != nil {
				exitWithError(err)
			}

			err := RunGetInstances(ctx, f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	return cmd
}

func RunGetInstances(ctx context.Context, f *util.Factory, out io.Writer, options *GetOptions) error {

	clientset, err := f.Clientset()
	if err != nil {
		return err
	}

	clusterName := rootCommand.ClusterName()
	options.clusterName = clusterName
	if clusterName == "" {
		return fmt.Errorf("--name is required")
	}

	cluster, err := clientset.GetCluster(ctx, options.clusterName)
	if err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("cluster not found %q", options.clusterName)
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return err
	}

	k8sClient, err := createK8sClient(cluster)
	if err != nil {
		return err
	}

	var status map[string]string
	nodeList, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.V(2).Infof("error listing nodes: %v", err)
	} else {
		status, _ = getNodeStatus(ctx, cloud, clientset, cluster, nodeList.Items)
	}

	var instances []*resources.Instance

	switch cloud.ProviderID() {
	case kops.CloudProviderAWS:
		rs, _ := aws.ListInstances(cloud, options.clusterName)
		for _, r := range rs {
			instances = append(instances, aws.GetInstanceFromResource(r))
		}
	case kops.CloudProviderOpenstack:
		rs, _ := openstack.ListResources(cloud.(osCloudup.OpenstackCloud), options.clusterName)
		for _, r := range rs {
			if r.Type == "Instance" {
				instances = append(instances, openstack.GetInstanceFromResource(r))
			}
		}
	default:
		return fmt.Errorf("cloud provider not supported")
	}

	switch options.output {
	case OutputTable:
		return instanceOutputTable(instances, status, out)
	default:
		return fmt.Errorf("Unsupported output format: %q", options.output)
	}
}

func instanceOutputTable(instances []*resources.Instance, status map[string]string, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("ID", func(i *resources.Instance) string {
		return i.ID
	})
	t.AddColumn("NAME", func(i *resources.Instance) string {
		return i.Name
	})
	t.AddColumn("STATUS", func(i *resources.Instance) string {
		s := status[i.ID]
		if s == "" {
			return "NotJoined"
		} else {
			return s
		}
	})
	t.AddColumn("ROLES", func(i *resources.Instance) string {
		return strings.Join(i.Roles, ", ")
	})
	t.AddColumn("INTERNAL-IP", func(i *resources.Instance) string {
		return i.PrivateAddress
	})
	t.AddColumn("INSTANCE-GROUP", func(i *resources.Instance) string {
		return i.InstanceGroup
	})
	t.AddColumn("MACHINE-TYPE", func(i *resources.Instance) string {
		return i.MachineType
	})
	return t.Render(instances, os.Stdout, "ID", "NAME", "STATUS", "ROLES", "INTERNAL-IP", "INSTANCE-GROUP", "MACHINE-TYPE")
}

func getNodeStatus(ctx context.Context, cloud fi.Cloud, clientset simple.Clientset, cluster *kops.Cluster, nodes []v1.Node) (map[string]string, error) {
	status := make(map[string]string)
	igList, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var instanceGroups []*kops.InstanceGroup
	for i := range igList.Items {
		instanceGroups = append(instanceGroups, &igList.Items[i])
	}
	igs := cloudinstances.GetNodeMap(nodes, cluster)

	cloudGroups, err := cloud.GetCloudGroups(cluster, instanceGroups, false, nodes)
	if err != nil {
		return nil, err
	}

	for _, cg := range cloudGroups {
		for _, instance := range cg.Ready {
			if instance.Detached {
				status[instance.ID] = "Detached"
			} else {
				if igs[instance.ID] != nil {
					status[instance.ID] = "Ready"
				} else {
					status[instance.ID] = "NotJoined"
				}
			}
		}
	}

	for _, cg := range cloudGroups {
		for _, node := range cg.NeedUpdate {
			if node.Detached {
				status[node.ID] = "Detached"
			} else {
				status[node.ID] = "NeedsUpdate"
			}
		}
	}
	return status, nil
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
