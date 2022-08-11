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

package cloudup

import (
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/util/pkg/reflectutils"
)

// Default Machine types for various types of instance group machine
const (
	defaultNodeMachineTypeGCE     = "n1-standard-2"
	defaultNodeMachineTypeDO      = "s-2vcpu-4gb"
	defaultNodeMachineTypeAzure   = "Standard_B2ms"
	defaultNodeMachineTypeHetzner = "cx21"

	defaultBastionMachineTypeGCE     = "f1-micro"
	defaultBastionMachineTypeAzure   = "Standard_B2ms"
	defaultBastionMachineTypeHetzner = "cx11"

	defaultMasterMachineTypeGCE     = "n1-standard-1"
	defaultMasterMachineTypeDO      = "s-2vcpu-4gb"
	defaultMasterMachineTypeAzure   = "Standard_B2ms"
	defaultMasterMachineTypeHetzner = "cx21"

	defaultDONodeImage = "ubuntu-20-04-x64"
)

// TODO: this hardcoded list can be replaced with DescribeInstanceTypes' DedicatedHostsSupported field
var awsDedicatedInstanceExceptions = map[string]bool{
	"t2.nano":   true,
	"t2.micro":  true,
	"t2.small":  true,
	"t2.medium": true,
	"t2.large":  true,
	"t2.xlarge": true,
}

// PopulateInstanceGroupSpec sets default values in the InstanceGroup
func PopulateInstanceGroupSpec(cluster *kops.Cluster, input *kops.InstanceGroup, cloud fi.Cloud, channel *kops.Channel) (*kops.InstanceGroup, error) {
	klog.Infof("Populating instance group spec for %q", input.GetName())

	var err error
	err = validation.ValidateInstanceGroup(input, nil, false).ToAggregate()
	if err != nil {
		return nil, fmt.Errorf("failed validating input specs: %w", err)
	}

	ig := &kops.InstanceGroup{}
	reflectutils.JSONMergeStruct(ig, input)

	spec := &ig.Spec

	// TODO: Clean up
	if ig.IsMaster() {
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType, err = defaultMachineType(cloud, cluster, ig)
			if err != nil {
				return nil, fmt.Errorf("error assigning default machine type for masters: %v", err)
			}

		}
		if ig.Spec.MinSize == nil {
			ig.Spec.MinSize = fi.Int32(1)
		}
		if ig.Spec.MaxSize == nil {
			ig.Spec.MaxSize = fi.Int32(1)
		}
	} else if ig.Spec.Role == kops.InstanceGroupRoleBastion {
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType, err = defaultMachineType(cloud, cluster, ig)
			if err != nil {
				return nil, fmt.Errorf("error assigning default machine type for bastions: %v", err)
			}
		}
		if ig.Spec.MinSize == nil {
			ig.Spec.MinSize = fi.Int32(1)
		}
		if ig.Spec.MaxSize == nil {
			ig.Spec.MaxSize = fi.Int32(1)
		}
	} else {
		if ig.IsAPIServerOnly() && !featureflag.APIServerNodes.Enabled() {
			return nil, fmt.Errorf("apiserver nodes requires the APIServerNodes feature flag to be enabled")
		}
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType, err = defaultMachineType(cloud, cluster, ig)
			if err != nil {
				return nil, fmt.Errorf("error assigning default machine type for nodes: %v", err)
			}
		}
		if ig.Spec.MinSize == nil {
			ig.Spec.MinSize = fi.Int32(2)
		}
		if ig.Spec.MaxSize == nil {
			ig.Spec.MaxSize = fi.Int32(2)
		}
	}

	if ig.Spec.Image == "" {
		architecture, err := MachineArchitecture(cloud, ig.Spec.MachineType)
		if err != nil {
			return nil, fmt.Errorf("unable to determine machine architecture for InstanceGroup %q: %v", ig.ObjectMeta.Name, err)
		}
		ig.Spec.Image = defaultImage(cluster, channel, architecture)
		if ig.Spec.Image == "" {
			return nil, fmt.Errorf("unable to determine default image for InstanceGroup %s", ig.ObjectMeta.Name)
		}
	}

	if ig.Spec.Tenancy != "" && ig.Spec.Tenancy != "default" {
		switch cluster.Spec.GetCloudProvider() {
		case kops.CloudProviderAWS:
			if _, ok := awsDedicatedInstanceExceptions[ig.Spec.MachineType]; ok {
				return nil, fmt.Errorf("invalid dedicated instance type: %s", ig.Spec.MachineType)
			}
		default:
			klog.Warning("Trying to set tenancy on non-AWS environment")
		}
	}

	if ig.IsMaster() {
		if len(ig.Spec.Subnets) == 0 {
			return nil, fmt.Errorf("master InstanceGroup %s did not specify any Subnets", ig.ObjectMeta.Name)
		}
	} else if ig.IsAPIServerOnly() && cluster.Spec.IsIPv6Only() {
		if len(ig.Spec.Subnets) == 0 {
			for _, subnet := range cluster.Spec.Subnets {
				if subnet.Type != kops.SubnetTypePrivate && subnet.Type != kops.SubnetTypeUtility {
					ig.Spec.Subnets = append(ig.Spec.Subnets, subnet.Name)
				}
			}
		}
	} else {
		if len(ig.Spec.Subnets) == 0 {
			for _, subnet := range cluster.Spec.Subnets {
				if subnet.Type != kops.SubnetTypeDualStack && subnet.Type != kops.SubnetTypeUtility {
					ig.Spec.Subnets = append(ig.Spec.Subnets, subnet.Name)
				}
			}
		}

		if len(ig.Spec.Subnets) == 0 {
			for _, subnet := range cluster.Spec.Subnets {
				if subnet.Type != kops.SubnetTypeUtility {
					ig.Spec.Subnets = append(ig.Spec.Subnets, subnet.Name)
				}
			}
		}
	}

	if len(ig.Spec.Subnets) == 0 {
		return nil, fmt.Errorf("unable to infer any Subnets for InstanceGroup %s ", ig.ObjectMeta.Name)
	}

	hasGPU := false
	clusterNvidia := false
	if cluster.Spec.Containerd != nil && cluster.Spec.Containerd.NvidiaGPU != nil && fi.BoolValue(cluster.Spec.Containerd.NvidiaGPU.Enabled) {
		clusterNvidia = true
	}
	igNvidia := false
	if ig.Spec.Containerd != nil && ig.Spec.Containerd.NvidiaGPU != nil && fi.BoolValue(ig.Spec.Containerd.NvidiaGPU.Enabled) {
		igNvidia = true
	}

	switch cluster.Spec.GetCloudProvider() {
	case kops.CloudProviderAWS:
		if clusterNvidia || igNvidia {
			mt, err := awsup.GetMachineTypeInfo(cloud.(awsup.AWSCloud), ig.Spec.MachineType)
			if err != nil {
				return ig, fmt.Errorf("error looking up machine type info: %v", err)
			}
			hasGPU = mt.GPU
		}
	case kops.CloudProviderOpenstack:
		if igNvidia {
			hasGPU = true
		}
	}

	if hasGPU {
		if ig.Spec.NodeLabels == nil {
			ig.Spec.NodeLabels = make(map[string]string)
		}
		ig.Spec.NodeLabels["kops.k8s.io/gpu"] = "1"
		hasNvidiaTaint := false
		for _, taint := range ig.Spec.Taints {
			if strings.HasPrefix(taint, "nvidia.com/gpu") {
				hasNvidiaTaint = true
			}
		}
		if !hasNvidiaTaint {
			ig.Spec.Taints = append(ig.Spec.Taints, "nvidia.com/gpu:NoSchedule")
		}
	}

	if ig.Spec.Manager == "" {
		ig.Spec.Manager = kops.InstanceManagerCloudGroup
	}

	if spec.Kubelet == nil {
		spec.Kubelet = &kops.KubeletConfigSpec{}
	}

	kubeletConfig := spec.Kubelet

	// We include the NodeLabels in the userdata even for Kubernetes 1.16 and later so that
	// rolling update will still replace nodes when they change.
	kubeletConfig.NodeLabels = nodelabels.BuildNodeLabels(cluster, ig)

	kubeletConfig.Taints = append(kubeletConfig.Taints, spec.Taints...)

	if ig.IsMaster() {
		reflectutils.JSONMergeStruct(kubeletConfig, cluster.Spec.MasterKubelet)

		// A few settings in Kubelet override those in MasterKubelet. I'm not sure why.
		if cluster.Spec.Kubelet != nil && cluster.Spec.Kubelet.AnonymousAuth != nil && !*cluster.Spec.Kubelet.AnonymousAuth {
			kubeletConfig.AnonymousAuth = fi.Bool(false)
		}
	} else {
		reflectutils.JSONMergeStruct(kubeletConfig, cluster.Spec.Kubelet)
	}

	if ig.Spec.Kubelet != nil {
		useSecureKubelet := fi.BoolValue(kubeletConfig.AnonymousAuth)

		reflectutils.JSONMergeStruct(kubeletConfig, spec.Kubelet)

		if useSecureKubelet {
			kubeletConfig.AnonymousAuth = fi.Bool(false)
		}
	}

	return ig, nil
}

// defaultMachineType returns the default MachineType for the instance group, based on the cloudprovider
func defaultMachineType(cloud fi.Cloud, cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	switch cluster.Spec.GetCloudProvider() {
	case kops.CloudProviderAWS:
		if ig.Spec.Manager == kops.InstanceManagerKarpenter {
			return "", nil
		}

		instanceType, err := cloud.(awsup.AWSCloud).DefaultInstanceType(cluster, ig)
		if err != nil {
			return "", fmt.Errorf("error finding default machine type: %v", err)
		}
		return instanceType, nil

	case kops.CloudProviderGCE:
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleMaster:
			return defaultMasterMachineTypeGCE, nil

		case kops.InstanceGroupRoleNode:
			return defaultNodeMachineTypeGCE, nil

		case kops.InstanceGroupRoleBastion:
			return defaultBastionMachineTypeGCE, nil
		}

	case kops.CloudProviderDO:
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleMaster:
			return defaultMasterMachineTypeDO, nil

		case kops.InstanceGroupRoleNode:
			return defaultNodeMachineTypeDO, nil

		}

	case kops.CloudProviderHetzner:
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleMaster:
			return defaultMasterMachineTypeHetzner, nil

		case kops.InstanceGroupRoleNode:
			return defaultNodeMachineTypeHetzner, nil

		case kops.InstanceGroupRoleBastion:
			return defaultBastionMachineTypeHetzner, nil
		}

	case kops.CloudProviderOpenstack:
		instanceType, err := cloud.(openstack.OpenstackCloud).DefaultInstanceType(cluster, ig)
		if err != nil {
			return "", fmt.Errorf("error finding default machine type: %v", err)
		}
		return instanceType, nil

	case kops.CloudProviderAzure:
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleMaster:
			return defaultMasterMachineTypeAzure, nil

		case kops.InstanceGroupRoleNode:
			return defaultNodeMachineTypeAzure, nil

		case kops.InstanceGroupRoleBastion:
			return defaultBastionMachineTypeAzure, nil
		}
	}

	klog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q, Role=%q", cluster.Spec.GetCloudProvider(), ig.Spec.Role)
	return "", nil
}
