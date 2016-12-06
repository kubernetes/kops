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

const DefaultNodeMachineTypeAWS = "t2.medium"
const DefaultNodeMachineTypeGCE = "n1-standard-2"

const DefaultMasterMachineTypeAWS = "m3.medium"

// us-east-2 does not (currently) support the m3 family; the c4 large is the cheapest non-burstable instance
const DefaultMasterMachineTypeAWS_USEAST2 = "c4.large"

const DefaultMasterMachineTypeGCE = "n1-standard-1"

// Default Machine type for bastion hosts
const DefaultBastionMachineTypeAWS = "t2.medium"
const DefaultBastionMasterMachineTypeGCE = "n1-standard-1"

// Default LoadBalancing IdleTimeout for bastion hosts
const DefaultBastionIdleTimeoutAWS = 120
const DefaultBastionIdleTimeoutGCE = 120

// PopulateInstanceGroupSpec sets default values in the InstanceGroup
// The InstanceGroup is simpler than the cluster spec, so we just populate in place (like the rest of k8s)
func PopulateInstanceGroupSpec(cluster *api.Cluster, input *api.InstanceGroup, channel *api.Channel) (*api.InstanceGroup, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	ig := &api.InstanceGroup{}
	utils.JsonMergeStruct(ig, input)

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
		if len(ig.Spec.Zones) == 0 {
			return nil, fmt.Errorf("Master InstanceGroup %s did not specify any Zones", ig.ObjectMeta.Name)
		}
	} else {
		if len(ig.Spec.Zones) == 0 {
			for _, z := range cluster.Spec.Zones {
				ig.Spec.Zones = append(ig.Spec.Zones, z.Name)
			}
		}
	}

	return ig, nil
}

// defaultNodeMachineType returns the default MachineType for nodes, based on the cloudprovider
func defaultNodeMachineType(cluster *api.Cluster) string {
	switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
	case fi.CloudProviderAWS:
		return DefaultNodeMachineTypeAWS
	case fi.CloudProviderGCE:
		return DefaultNodeMachineTypeGCE
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
		if region == "us-east-2" {
			glog.Warningf("%q instance is not available in region %q, will set master to %q instead", DefaultMasterMachineTypeAWS, region, DefaultMasterMachineTypeAWS_USEAST2)
			return DefaultMasterMachineTypeAWS_USEAST2
		}
		return DefaultMasterMachineTypeAWS
	case fi.CloudProviderGCE:
		return DefaultMasterMachineTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

// defaultBastionMachineType returns the default MachineType for bastion host, based on the cloudprovider
func DefaultBastionMachineType(cluster *api.Cluster) string {
	switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
	case fi.CloudProviderAWS:
		return DefaultBastionMachineTypeAWS
	case fi.CloudProviderGCE:
		return DefaultBastionMasterMachineTypeGCE
	default:
		glog.V(2).Infof("Cannot set default MachineType for CloudProvider=%q", cluster.Spec.CloudProvider)
		return ""
	}
}

// defaultIdleTimeout returns the default Idletimeout for bastion loadbalancer, based on the cloudprovider
func DefaultBastionIdleTimeout(cluster *api.Cluster) int {
	switch fi.CloudProviderID(cluster.Spec.CloudProvider) {
	case fi.CloudProviderAWS:
		return DefaultBastionIdleTimeoutAWS
	case fi.CloudProviderGCE:
		return DefaultBastionIdleTimeoutGCE
	default:
		glog.V(2).Infof("Cannot set default IdleTimeout for CloudProvider=%q", cluster.Spec.CloudProvider)
		return 0
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
