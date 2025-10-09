/*
Copyright 2025 The Kubernetes Authors.

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

package clusterapi

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	"k8s.io/kops/clusterapi/pkg/builders"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands/commandutils"
)

type GenerateMachineDeploymentOptions struct {
	ClusterName string

	// Name is the name of the MachineDeployment (and other objects) to create
	Name string
	// Namespace is the namespace for the MachineDeployment (and other objects) to create
	Namespace string

	// Replicas is the number of replicas for the MachineDeployment
	Replicas int

	// Zones is the set of zones for the MachineDeployments (also called failureDomain)
	Zones []string

	// InstanceType is the instance type for the MachineDeployment
	InstanceType string

	// Subnet is the subnet for the MachineDeployment
	Subnet string

	// Image is the image for the MachineDeployment
	Image string
}

func (o *GenerateMachineDeploymentOptions) InitDefaults() {
	o.Replicas = 1
	o.Namespace = "kube-system"
}

func BuildGenerateMachineDeploymentCommand(f commandutils.Factory, out io.Writer) *cobra.Command {
	options := &GenerateMachineDeploymentOptions{}
	options.InitDefaults()

	cmd := &cobra.Command{
		Use:   "machinedeployment [CLUSTER]",
		Short: i18n.T(`Generate a MachineDeployment configuration`),
		Long: templates.LongDesc(i18n.T(`
			Add nodes to a cluster by generating a MachineDeployment configuration.`)),
		Example: templates.Examples(i18n.T(`
			kops toolbox clusterapi generate machinedeployment --name k8s-cluster.example.com --name machinedeployment1 --replicas 2 | kubectl apply -f -
		`)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunGenerateMachineDeployment(cmd.Context(), f, out, options)
		},
	}

	cmd.Flags().StringVar(&options.ClusterName, "cluster", options.ClusterName, "Name of cluster to join")
	cmd.Flags().StringVar(&options.Name, "name", options.Name, "Name of MachineDeployment (and other objects) to create")
	cmd.Flags().StringVar(&options.Namespace, "namespace", options.Namespace, "Namespace for objects")
	cmd.Flags().IntVar(&options.Replicas, "replicas", options.Replicas, "Number of replicas for MachineDeployment")
	cmd.Flags().StringArrayVar(&options.Zones, "zones", options.Zones, "Zones for the MachineDeployment (if not specified, will be deployed to all discovered zones in the cluster)")
	cmd.Flags().StringVar(&options.InstanceType, "instance-type", options.InstanceType, "Instance type for the MachineDeployment (if not specified, an arbitrary instance type from the instance groups will be used)")
	cmd.Flags().StringVar(&options.Subnet, "subnet", options.Subnet, "Subnet for the MachineDeployment (if not specified, an arbitrary subnet from the instance groups will be used)")
	cmd.Flags().StringVar(&options.Image, "image", options.Image, "Image for the MachineDeployment (if not specified, an arbitrary image from the instance groups will be used)")
	return cmd
}

func RunGenerateMachineDeployment(ctx context.Context, f commandutils.Factory, out io.Writer, options *GenerateMachineDeploymentOptions) error {
	log := klog.FromContext(ctx)

	if options.ClusterName == "" {
		return fmt.Errorf("cluster is required")
	}
	if options.Name == "" {
		return fmt.Errorf("name is required")
	}

	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	cluster, err := clientset.GetCluster(ctx, options.ClusterName)
	if err != nil {
		return err
	}

	// TODO: Sync with LinkToSubnet and support ID
	subnet := options.Subnet
	if subnet == "" {
		subnets := sets.New[string]()
		for _, subnet := range cluster.Spec.Networking.Subnets {
			subnets.Insert(subnet.Name)
		}

		if subnets.Len() == 0 {
			return fmt.Errorf("no subnets found in cluster %q", options.ClusterName)
		}
		if subnets.Len() > 1 {
			klog.Warningf("multiple subnets found in cluster %q; using an arbitrary subnet", options.ClusterName)
		}
		subnet = sets.List(subnets)[0]
		log.Info("selected subnet", "subnet", subnet)
	}

	zones := options.Zones

	// TODO: When to use cluster zones, when to use IG zones?
	// if zone == "" {
	// 	for _, subnetCandidate := range cluster.Spec.Networking.Subnets {
	// 		if subnetCandidate.Name == subnet {
	// 			zone = subnetCandidate.Zone
	// 		}
	// 	}
	// 	if zone == "" {
	// 		return fmt.Errorf("unable to determine zone for subnet %q in cluster %q; please specify --zone", subnet, options.ClusterName)
	// 	}
	// 	log.Info("selected zone", "zone", zone)
	// }

	instanceType := options.InstanceType
	image := options.Image
	if len(zones) == 0 || instanceType == "" || image == "" {
		instanceGroups, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		if len(instanceGroups.Items) == 0 {
			return fmt.Errorf("no instance groups found in cluster %q", options.ClusterName)
		}

		if len(zones) == 0 {
			allZones := sets.New[string]()
			for _, ig := range instanceGroups.Items {
				for _, z := range ig.Spec.Zones {
					allZones.Insert(z)
				}
			}
			if allZones.Len() == 0 {
				return fmt.Errorf("no zones found in instance groups in cluster %q", options.ClusterName)
			}

			zones = sets.List(allZones)
			log.Info("selected zones", "zones", zones)
		}

		if instanceType == "" {
			instanceTypes := sets.New[string]()
			for _, ig := range instanceGroups.Items {
				if ig.Spec.MachineType != "" {
					instanceTypes.Insert(ig.Spec.MachineType)
				}
			}
			if instanceTypes.Len() == 0 {
				return fmt.Errorf("no instance types found in instance groups in cluster %q", options.ClusterName)
			}
			if instanceTypes.Len() > 1 {
				klog.Warningf("multiple instance types found in instance groups in cluster %q; using an arbitrary instance type", options.ClusterName)
			}
			instanceType = sets.List(instanceTypes)[0]
			log.Info("selected instance type", "instanceType", instanceType)
		}

		if image == "" {
			images := sets.New[string]()
			for _, ig := range instanceGroups.Items {
				if ig.Spec.Image != "" {
					images.Insert(ig.Spec.Image)
				}
			}
			if images.Len() == 0 {
				return fmt.Errorf("no images found in instance groups in cluster %q", options.ClusterName)
			}
			if images.Len() > 1 {
				klog.Warningf("multiple images found in instance groups in cluster %q; using an arbitrary image", options.ClusterName)
			}
			image = sets.List(images)[0]
			log.Info("selected image", "image", image)
		}
	}

	kubernetesVersion := cluster.Spec.KubernetesVersion

	role := kops.InstanceGroupRoleNode

	b := &builders.MachineDeploymentBuilder{
		ClusterName:       cluster.GetName(),
		Namespace:         options.Namespace,
		Name:              options.Name,
		Replicas:          options.Replicas,
		Zones:             zones,
		MachineType:       instanceType,
		Subnet:            subnet,
		Image:             image,
		KubernetesVersion: kubernetesVersion,
		Role:              role,
	}

	objects, err := b.BuildObjects(ctx)
	if err != nil {
		return err
	}

	for i, obj := range objects {
		b, err := yaml.Marshal(obj.Object)
		if err != nil {
			return fmt.Errorf("error marshalling %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}
		if i != 0 {
			fmt.Fprintf(out, "---\n")
		}
		out.Write(b)
	}

	return nil
}
