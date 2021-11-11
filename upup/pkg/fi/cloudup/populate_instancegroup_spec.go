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

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/blang/semver/v4"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/reflectutils"
)

// Default Machine types for various types of instance group machine
const (
	defaultNodeMachineTypeGCE   = "n1-standard-2"
	defaultNodeMachineTypeDO    = "s-2vcpu-4gb"
	defaultNodeMachineTypeALI   = "ecs.n2.medium"
	defaultNodeMachineTypeAzure = "Standard_B2ms"

	defaultBastionMachineTypeGCE   = "f1-micro"
	defaultBastionMachineTypeALI   = "ecs.n2.small"
	defaultBastionMachineTypeAzure = "Standard_B2ms"

	defaultMasterMachineTypeGCE   = "n1-standard-1"
	defaultMasterMachineTypeDO    = "s-2vcpu-4gb"
	defaultMasterMachineTypeALI   = "ecs.n2.medium"
	defaultMasterMachineTypeAzure = "Standard_B2ms"

	defaultDONodeImage  = "ubuntu-20-04-x64"
	defaultALINodeImage = "centos_7_04_64_20G_alibase_201701015.vhd"
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
	var err error
	err = validation.ValidateInstanceGroup(input, nil).ToAggregate()
	if err != nil {
		return nil, err
	}

	ig := &kops.InstanceGroup{}
	reflectutils.JSONMergeStruct(ig, input)

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
		switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
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
	} else {
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

	if cluster.Spec.Containerd != nil && cluster.Spec.Containerd.NvidiaGPU != nil && fi.BoolValue(cluster.Spec.Containerd.NvidiaGPU.Enabled) {
		switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
		case kops.CloudProviderAWS:
			mt, err := awsup.GetMachineTypeInfo(cloud.(awsup.AWSCloud), ig.Spec.MachineType)
			if err != nil {
				return ig, fmt.Errorf("error looking up machine type info: %v", err)
			}
			if mt.GPU {
				if ig.Spec.NodeLabels == nil {
					ig.Spec.NodeLabels = make(map[string]string)
				}
				ig.Spec.NodeLabels["kops.k8s.io/gpu"] = "1"
				ig.Spec.Taints = append(ig.Spec.Taints, "nvidia.com/gpu:NoSchedule")
			}
		}
	}
	return ig, nil
}

// defaultMachineType returns the default MachineType for the instance group, based on the cloudprovider
func defaultMachineType(cloud fi.Cloud, cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:

		instanceType, err := cloud.(awsup.AWSCloud).DefaultInstanceType(cluster, ig)
		if err != nil {
			return "", fmt.Errorf("error finding default machine type: %v", err)
		}
		return instanceType, nil

	case kops.CloudProviderGCE:
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleAPIServer:
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

	case kops.CloudProviderOpenstack:
		instanceType, err := cloud.(openstack.OpenstackCloud).DefaultInstanceType(cluster, ig)
		if err != nil {
			return "", fmt.Errorf("error finding default machine type: %v", err)
		}
		return instanceType, nil

	case kops.CloudProviderALI:
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleMaster:
			return defaultMasterMachineTypeALI, nil

		case kops.InstanceGroupRoleNode:
			return defaultNodeMachineTypeALI, nil

		case kops.InstanceGroupRoleBastion:
			return defaultBastionMachineTypeALI, nil
		}

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

	klog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q, Role=%q", cluster.Spec.CloudProvider, ig.Spec.Role)
	return "", nil
}

// defaultImage returns the default Image, based on the cloudprovider
func defaultImage(cluster *kops.Cluster, channel *kops.Channel, architecture architectures.Architecture) string {
	if channel != nil {
		var kubernetesVersion *semver.Version
		if cluster.Spec.KubernetesVersion != "" {
			var err error
			kubernetesVersion, err = util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
			if err != nil {
				klog.Warningf("cannot parse KubernetesVersion %q in cluster", cluster.Spec.KubernetesVersion)
			}
		}
		if kubernetesVersion != nil {
			image := channel.FindImage(kops.CloudProviderID(cluster.Spec.CloudProvider), *kubernetesVersion, architecture)
			if image != nil {
				return image.Name
			}
		}
	}

	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderDO:
		return defaultDONodeImage
	case kops.CloudProviderALI:
		return defaultALINodeImage
	}
	klog.Infof("Cannot set default Image for CloudProvider=%q", cluster.Spec.CloudProvider)
	return ""
}

func MachineArchitecture(cloud fi.Cloud, machineType string) (architectures.Architecture, error) {
	switch cloud.ProviderID() {
	case kops.CloudProviderAWS:
		info, err := cloud.(awsup.AWSCloud).DescribeInstanceType(machineType)
		if err != nil {
			return "", fmt.Errorf("error finding instance info for instance type %q: %v", machineType, err)
		}
		if info.ProcessorInfo == nil || len(info.ProcessorInfo.SupportedArchitectures) == 0 {
			return "", fmt.Errorf("error finding architecture info for instance type %q", machineType)
		}
		var unsupported []string
		for _, arch := range info.ProcessorInfo.SupportedArchitectures {
			// Return the first found supported architecture, in order of popularity
			switch fi.StringValue(arch) {
			case ec2.ArchitectureTypeX8664:
				return architectures.ArchitectureAmd64, nil
			case ec2.ArchitectureTypeArm64:
				return architectures.ArchitectureArm64, nil
			default:
				unsupported = append(unsupported, fi.StringValue(arch))
			}
		}
		return "", fmt.Errorf("unsupported architecture for instance type %q: %v", machineType, unsupported)
	default:
		// No other clouds are known to support any other architectures at this time
		return architectures.ArchitectureAmd64, nil
	}
}
