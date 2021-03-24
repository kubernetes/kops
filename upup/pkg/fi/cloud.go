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

package fi

import (
	"regexp"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
)

type Cloud interface {
	ProviderID() kops.CloudProviderID

	DNS() (dnsprovider.Interface, error)

	// FindVPCInfo looks up the specified VPC by id, returning info if found, otherwise (nil, nil).
	FindVPCInfo(id string) (*VPCInfo, error)

	// DeleteInstance deletes a cloud instance.
	DeleteInstance(instance *cloudinstances.CloudInstance) error

	// DeleteGroup deletes the cloud resources that make up a CloudInstanceGroup, including the instances.
	DeleteGroup(group *cloudinstances.CloudInstanceGroup) error

	// DetachInstance causes a cloud instance to no longer be counted against the group's size limits.
	DetachInstance(instance *cloudinstances.CloudInstance) error

	// GetCloudGroups returns a map of cloud instances that back a kops cluster.
	// Detached instances must be returned in the NeedUpdate slice.
	GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error)

	// Region returns the cloud region bound to the cloud instance.
	// If the region concept does not apply, returns "".
	Region() string
}

type VPCInfo struct {
	// CIDR is the IP address range for the VPC
	CIDR string

	// Subnets is a list of subnets that are part of the VPC
	Subnets []*SubnetInfo
}

type SubnetInfo struct {
	ID   string
	Zone string
	CIDR string
}

// GuessCloudForZone tries to infer the cloudprovider from the zone name
// Ali has the same zoneNames as AWS in the regions outside China, so if use AliCloud to install k8s in the regions outside China,
// the users need to provide parameter "--cloud". But the regions inside China can be easily identified.
func GuessCloudForZone(zone string) (kops.CloudProviderID, bool) {

	structuredProviders := map[kops.CloudProviderID]string{
		kops.CloudProviderAWS: "^[a-z]{2}-[a-z\\-]+?-[0-9]{1,2}[a-z]$",
		kops.CloudProviderGCE: "^[a-z]+?-[a-z]+?[0-9]{1,2}-[a-z]$",
		kops.CloudProviderDO:  "^[a-z]{3}[0-9]+$",
		kops.CloudProviderALI: "^[a-z]{2}-[a-z]+-[a-z]+$",
	}

	unstructuredProviders := map[kops.CloudProviderID]string{
		kops.CloudProviderAzure: "^[a-z0-9]+$",
	}

	found := false
	var cloudProvider kops.CloudProviderID

	for provider, regex := range structuredProviders {
		if match, _ := regexp.MatchString(regex, zone); match {
			found = true
			cloudProvider = provider
			break
		}
	}

	//azure is more generic, so only test if not yet found
	if !found {
		for provider, regex := range unstructuredProviders {
			if match, _ := regexp.MatchString(regex, zone); match {
				found = true
				cloudProvider = provider
				break
			}
		}
	}

	return cloudProvider, found
}
