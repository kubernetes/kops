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

package fi

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

type Cloud interface {
	ProviderID() kops.CloudProviderID

	DNS() (dnsprovider.Interface, error)

	// FindVPCInfo looks up the specified VPC by id, returning info if found, otherwise (nil, nil)
	FindVPCInfo(id string) (*VPCInfo, error)
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

// zonesToCloud allows us to infer from certain well-known zones to a cloud
// Note it is safe to "overmap" zones that don't exist: we'll check later if the zones actually exist
var zonesToCloud = map[string]kops.CloudProviderID{
	"us-east-1a": kops.CloudProviderAWS,
	"us-east-1b": kops.CloudProviderAWS,
	"us-east-1c": kops.CloudProviderAWS,
	"us-east-1d": kops.CloudProviderAWS,
	"us-east-1e": kops.CloudProviderAWS,

	"us-east-2a": kops.CloudProviderAWS,
	"us-east-2b": kops.CloudProviderAWS,
	"us-east-2c": kops.CloudProviderAWS,
	"us-east-2d": kops.CloudProviderAWS,
	"us-east-2e": kops.CloudProviderAWS,

	"us-west-1a": kops.CloudProviderAWS,
	"us-west-1b": kops.CloudProviderAWS,
	"us-west-1c": kops.CloudProviderAWS,
	"us-west-1d": kops.CloudProviderAWS,
	"us-west-1e": kops.CloudProviderAWS,

	"us-west-2a": kops.CloudProviderAWS,
	"us-west-2b": kops.CloudProviderAWS,
	"us-west-2c": kops.CloudProviderAWS,
	"us-west-2d": kops.CloudProviderAWS,
	"us-west-2e": kops.CloudProviderAWS,

	"ca-central-1a": kops.CloudProviderAWS,
	"ca-central-1b": kops.CloudProviderAWS,

	"eu-west-1a": kops.CloudProviderAWS,
	"eu-west-1b": kops.CloudProviderAWS,
	"eu-west-1c": kops.CloudProviderAWS,
	"eu-west-1d": kops.CloudProviderAWS,
	"eu-west-1e": kops.CloudProviderAWS,
	"eu-west-2a": kops.CloudProviderAWS,
	"eu-west-2b": kops.CloudProviderAWS,

	"eu-central-1a": kops.CloudProviderAWS,
	"eu-central-1b": kops.CloudProviderAWS,
	"eu-central-1c": kops.CloudProviderAWS,
	"eu-central-1d": kops.CloudProviderAWS,
	"eu-central-1e": kops.CloudProviderAWS,

	"ap-south-1a": kops.CloudProviderAWS,
	"ap-south-1b": kops.CloudProviderAWS,
	"ap-south-1c": kops.CloudProviderAWS,
	"ap-south-1d": kops.CloudProviderAWS,
	"ap-south-1e": kops.CloudProviderAWS,

	"ap-southeast-1a": kops.CloudProviderAWS,
	"ap-southeast-1b": kops.CloudProviderAWS,
	"ap-southeast-1c": kops.CloudProviderAWS,
	"ap-southeast-1d": kops.CloudProviderAWS,
	"ap-southeast-1e": kops.CloudProviderAWS,

	"ap-southeast-2a": kops.CloudProviderAWS,
	"ap-southeast-2b": kops.CloudProviderAWS,
	"ap-southeast-2c": kops.CloudProviderAWS,
	"ap-southeast-2d": kops.CloudProviderAWS,
	"ap-southeast-2e": kops.CloudProviderAWS,

	"ap-northeast-1a": kops.CloudProviderAWS,
	"ap-northeast-1b": kops.CloudProviderAWS,
	"ap-northeast-1c": kops.CloudProviderAWS,
	"ap-northeast-1d": kops.CloudProviderAWS,
	"ap-northeast-1e": kops.CloudProviderAWS,

	"ap-northeast-2a": kops.CloudProviderAWS,
	"ap-northeast-2b": kops.CloudProviderAWS,
	"ap-northeast-2c": kops.CloudProviderAWS,
	"ap-northeast-2d": kops.CloudProviderAWS,
	"ap-northeast-2e": kops.CloudProviderAWS,

	"sa-east-1a": kops.CloudProviderAWS,
	"sa-east-1b": kops.CloudProviderAWS,
	"sa-east-1c": kops.CloudProviderAWS,
	"sa-east-1d": kops.CloudProviderAWS,
	"sa-east-1e": kops.CloudProviderAWS,

	"cn-north-1a": kops.CloudProviderAWS,
	"cn-north-1b": kops.CloudProviderAWS,

	"us-gov-west-1a": kops.CloudProviderAWS,
	"us-gov-west-1b": kops.CloudProviderAWS,

	// GCE
	"asia-east1-a": kops.CloudProviderGCE,
	"asia-east1-b": kops.CloudProviderGCE,
	"asia-east1-c": kops.CloudProviderGCE,
	"asia-east1-d": kops.CloudProviderGCE,

	"asia-northeast1-a": kops.CloudProviderGCE,
	"asia-northeast1-b": kops.CloudProviderGCE,
	"asia-northeast1-c": kops.CloudProviderGCE,
	"asia-northeast1-d": kops.CloudProviderGCE,

	"europe-west1-a": kops.CloudProviderGCE,
	"europe-west1-b": kops.CloudProviderGCE,
	"europe-west1-c": kops.CloudProviderGCE,
	"europe-west1-d": kops.CloudProviderGCE,
	"europe-west1-e": kops.CloudProviderGCE,

	"us-central1-a": kops.CloudProviderGCE,
	"us-central1-b": kops.CloudProviderGCE,
	"us-central1-c": kops.CloudProviderGCE,
	"us-central1-d": kops.CloudProviderGCE,
	"us-central1-e": kops.CloudProviderGCE,
	"us-central1-f": kops.CloudProviderGCE,
	"us-central1-g": kops.CloudProviderGCE,
	"us-central1-h": kops.CloudProviderGCE,

	"us-east1-a": kops.CloudProviderGCE,
	"us-east1-b": kops.CloudProviderGCE,
	"us-east1-c": kops.CloudProviderGCE,
	"us-east1-d": kops.CloudProviderGCE,

	"us-west1-a": kops.CloudProviderGCE,
	"us-west1-b": kops.CloudProviderGCE,
	"us-west1-c": kops.CloudProviderGCE,
	"us-west1-d": kops.CloudProviderGCE,

	"nyc1": kops.CloudProviderDO,
	"nyc2": kops.CloudProviderDO,
	"nyc3": kops.CloudProviderDO,

	"sfo1": kops.CloudProviderDO,
	"sfo2": kops.CloudProviderDO,

	"ams2": kops.CloudProviderDO,
	"ams3": kops.CloudProviderDO,

	"tor1": kops.CloudProviderDO,

	"sgp1": kops.CloudProviderDO,

	"lon1": kops.CloudProviderDO,

	"fra1": kops.CloudProviderDO,

	"blr1": kops.CloudProviderDO,
}

// GuessCloudForZone tries to infer the cloudprovider from the zone name
func GuessCloudForZone(zone string) (kops.CloudProviderID, bool) {
	c, found := zonesToCloud[zone]
	return c, found
}
