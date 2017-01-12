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

	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// Default Machine types for various types of instance group machine
const (
	defaultNodeMachineTypeAWS = "t2.medium"
	defaultNodeMachineTypeGCE = "n1-standard-2"

	defaultBastionMachineTypeAWS = "t2.micro"
	defaultBastionMachineTypeGCE = "f1-micro"

	defaultMasterMachineTypeGCE = "n1-standard-1"
	defaultMasterMachineTypeAWS = "m3.medium"
)

var masterMachineTypeExceptions = map[string]string{
	// Some regions do not (currently) support the m3 family; the c4 large is the cheapest non-burstable instance
	"us-east-2":    "c4.large",
	"ca-central-1": "c4.large",
	"eu-west-2":    "c4.large",
}

// PopulateInstanceGroupSpec sets default values in the InstanceGroup
// The InstanceGroup is simpler than the cluster spec, so we just populate in place (like the rest of k8s)
func PopulateInstanceGroupSpec(cluster *api.Cluster, input *api.InstanceGroup, channel *api.Channel) (*api.InstanceGroup, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	ig := &api.InstanceGroup{}
	utils.JsonMergeStruct(ig, input)

	// TODO: Clean up
	if ig.IsMaster() {
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType = defaultMasterMachineType(cluster)
		}
		if ig.Spec.MinSize == nil {
			ig.Spec.MinSize = fi.Int(1)
		}
		if ig.Spec.MaxSize == nil {
			ig.Spec.MaxSize = fi.Int(1)
		}
	} else if ig.Spec.Role == api.InstanceGroupRoleBastion {
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType = defaultBastionMachineType(cluster)
		}
		if ig.Spec.MinSize == nil {
			ig.Spec.MinSize = fi.Int(1)
		}
		if ig.Spec.MaxSize == nil {
			ig.Spec.MaxSize = fi.Int(1)
		}
	} else {
		if ig.Spec.MachineType == "" {
			ig.Spec.MachineType = defaultNodeMachineType(cluster)
		}
		if ig.Spec.MinSize == nil {
			ig.Spec.MinSize = fi.Int(2)
		}
		if ig.Spec.MaxSize == nil {
			ig.Spec.MaxSize = fi.Int(2)
		}
	}

	if ig.Spec.AssociatePublicIP == nil {
		ig.Spec.AssociatePublicIP = fi.Bool(true)
	}

	if ig.Spec.Image == "" {
		ig.Spec.Image = defaultImage(cluster, channel)
	}

	if ig.IsMaster() {
		if len(ig.Spec.Subnets) == 0 {
			return nil, fmt.Errorf("Master InstanceGroup %s did not specify any Subnets", ig.ObjectMeta.Name)
		}
	} else if ig.Spec.Role == api.InstanceGroupRoleBastion {
		if len(ig.Spec.Subnets) == 0 {
			for _, subnet := range cluster.Spec.Subnets {
				if subnet.Type == api.SubnetTypeUtility {
					ig.Spec.Subnets = append(ig.Spec.Subnets, subnet.Name)
				}
			}
		}
	} else {
		if len(ig.Spec.Subnets) == 0 {
			for _, subnet := range cluster.Spec.Subnets {
				if subnet.Type != api.SubnetTypeUtility {
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

// defaultNodeMachineType returns the default MachineType for nodes, based on the cloudprovider
func defaultNodeMachineType(cluster *api.Cluster) string {
	switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
	case fi.CloudProviderAWS:
		return defaultNodeMachineTypeAWS
	case fi.CloudProviderGCE:
		return defaultNodeMachineTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

// defaultMasterMachineType returns the default MachineType for masters, based on the cloudprovider
func defaultMasterMachineType(cluster *api.Cluster) string {
	// TODO: We used to have logic like the following...
	//	{{ if gt .TotalNodeCount 500 }}
	//	MasterMachineType: n1-standard-32
	//	{{ else if gt .TotalNodeCount 250 }}
	//MasterMachineType: n1-standard-16
	//{{ else if gt .TotalNodeCount 100 }}
	//MasterMachineType: n1-standard-8
	//{{ else if gt .TotalNodeCount 10 }}
	//MasterMachineType: n1-standard-4
	//{{ else if gt .TotalNodeCount 5 }}
	//MasterMachineType: n1-standard-2
	//{{ else }}
	//MasterMachineType: n1-standard-1
	//{{ end }}
	//
	//{{ if gt TotalNodeCount 500 }}
	//MasterMachineType: c4.8xlarge
	//{{ else if gt TotalNodeCount 250 }}
	//MasterMachineType: c4.4xlarge
	//{{ else if gt TotalNodeCount 100 }}
	//MasterMachineType: m3.2xlarge
	//{{ else if gt TotalNodeCount 10 }}
	//MasterMachineType: m3.xlarge
	//{{ else if gt TotalNodeCount 5 }}
	//MasterMachineType: m3.large
	//{{ else }}
	//MasterMachineType: m3.medium
	//{{ end }}

	switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
	case fi.CloudProviderAWS:
		region, err := awsup.FindRegion(cluster)
		if err != nil {
			glog.Warningf("cannot determine region from cluster zones: %v", err)
		}
		// Check for special-cases
		masterMachineType := masterMachineTypeExceptions[region]
		if masterMachineType != "" {
			glog.Warningf("%q instance is not available in region %q, will set master to %q instead", defaultMasterMachineTypeAWS, region, masterMachineType)
			return masterMachineType
		}
		return defaultMasterMachineTypeAWS
	case fi.CloudProviderGCE:
		return defaultMasterMachineTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

// defaultBastionMachineType returns the default MachineType for bastions, based on the cloudprovider
func defaultBastionMachineType(cluster *api.Cluster) string {
	switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
	case fi.CloudProviderAWS:
		return defaultBastionMachineTypeAWS
	case fi.CloudProviderGCE:
		return defaultBastionMachineTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

// defaultImage returns the default Image, based on the cloudprovider
func defaultImage(cluster *api.Cluster, channel *api.Channel) string {
	if channel != nil {
		image := channel.FindImage(fi.CloudProviderID(cluster.Spec.CloudProvider))
		if image != nil {
			return image.Name
		}
	}

	glog.Infof("Cannot set default Image for CloudProvider=%q", cluster.Spec.CloudProvider)
	return ""
}
