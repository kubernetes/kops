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

package cloudup

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// Default Machine types for various types of instance group machine
const (
	defaultNodeMachineTypeGCE     = "n1-standard-2"
	defaultNodeMachineTypeVSphere = "vsphere_node"

	defaultBastionMachineTypeGCE     = "f1-micro"
	defaultBastionMachineTypeVSphere = "vsphere_bastion"

	defaultMasterMachineTypeGCE     = "n1-standard-1"
	defaultMasterMachineTypeVSphere = "vsphere_master"

	defaultVSphereNodeImage = "kops_ubuntu_16_04.ova"
)

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
func PopulateInstanceGroupSpec(cluster *kops.Cluster, input *kops.InstanceGroup, channel *kops.Channel) (*kops.InstanceGroup, error) {
	err := validation.ValidateInstanceGroup(input)
	if err != nil {
		return nil, err
	}

	ig := &kops.InstanceGroup{}
	utils.JsonMergeStruct(ig, input)

	// TODO: Clean up
	if ig.IsMaster() {
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType, err = defaultMachineType(cluster, ig)
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
			ig.Spec.MachineType, err = defaultMachineType(cluster, ig)
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
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType, err = defaultMachineType(cluster, ig)
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
		ig.Spec.Image = defaultImage(cluster, channel)
	}

	if ig.Spec.Tenancy != "" && ig.Spec.Tenancy != "default" {
		switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
		case kops.CloudProviderAWS:
			if _, ok := awsDedicatedInstanceExceptions[ig.Spec.MachineType]; ok {
				return nil, fmt.Errorf("Invalid dedicated instance type: %s", ig.Spec.MachineType)
			}
		default:
			glog.Warning("Trying to set tenancy on non-AWS environment")
		}
	}

	if ig.IsMaster() {
		if len(ig.Spec.Subnets) == 0 {
			return nil, fmt.Errorf("Master InstanceGroup %s did not specify any Subnets", ig.ObjectMeta.Name)
		}
	} else if ig.Spec.Role == kops.InstanceGroupRoleBastion {
		if len(ig.Spec.Subnets) == 0 {
			for _, subnet := range cluster.Spec.Subnets {
				if subnet.Type == kops.SubnetTypeUtility {
					ig.Spec.Subnets = append(ig.Spec.Subnets, subnet.Name)
				}
			}
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

	return ig, nil
}

// defaultMachineType returns the default MachineType for the instance group, based on the cloudprovider
func defaultMachineType(cluster *kops.Cluster, ig *kops.InstanceGroup) (string, error) {
	switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
	case kops.CloudProviderAWS:
		cloud, err := BuildCloud(cluster)
		if err != nil {
			return "", fmt.Errorf("error building cloud for AWS cluster: %v", err)
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

	case kops.CloudProviderVSphere:
		switch ig.Spec.Role {
		case kops.InstanceGroupRoleMaster:
			return defaultMasterMachineTypeVSphere, nil

		case kops.InstanceGroupRoleNode:
			return defaultNodeMachineTypeVSphere, nil

		case kops.InstanceGroupRoleBastion:
			return defaultBastionMachineTypeVSphere, nil
		}
	}

	glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q, Role=%q", cluster.Spec.CloudProvider, ig.Spec.Role)
	return "", nil
}

// defaultImage returns the default Image, based on the cloudprovider
func defaultImage(cluster *kops.Cluster, channel *kops.Channel) string {
	if channel != nil {
		var kubernetesVersion *semver.Version
		if cluster.Spec.KubernetesVersion != "" {
			var err error
			kubernetesVersion, err = util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
			if err != nil {
				glog.Warningf("cannot parse KubernetesVersion %q in cluster", cluster.Spec.KubernetesVersion)
			}
		}
		if kubernetesVersion != nil {
			image := channel.FindImage(kops.CloudProviderID(cluster.Spec.CloudProvider), *kubernetesVersion)
			if image != nil {
				return image.Name
			}
		}
	} else if kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderVSphere {
		return defaultVSphereNodeImage
	}
	glog.Infof("Cannot set default Image for CloudProvider=%q", cluster.Spec.CloudProvider)
	return ""
}
