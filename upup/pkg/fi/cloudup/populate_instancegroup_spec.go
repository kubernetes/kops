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
// The InstanceGroup is simpler than the cluster spec, so we just populate in place (like the rest of k8s)
func PopulateInstanceGroupSpec(cluster *kops.Cluster, input *kops.InstanceGroup, cloud fi.Cloud, channel *kops.Channel) (*kops.InstanceGroup, error) {
	klog.Infof("Populating instance group spec for %q", input.GetName())

	var err error
	err = validation.ValidateInstanceGroup(input, nil, false).ToAggregate()
	if err != nil {
		return nil, fmt.Errorf("failed validating input specs: %w", err)
	}

	ig := &kops.InstanceGroup{}
	reflectutils.JSONMergeStruct(ig, input)

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
