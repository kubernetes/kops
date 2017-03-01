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

import "k8s.io/kubernetes/federation/pkg/dnsprovider"

type CloudProviderID string

const CloudProviderAWS CloudProviderID = "aws"
const CloudProviderGCE CloudProviderID = "gce"

type Cloud interface {
	ProviderID() CloudProviderID

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
var zonesToCloud = map[string]CloudProviderID{
	"us-east-1a": CloudProviderAWS,
	"us-east-1b": CloudProviderAWS,
	"us-east-1c": CloudProviderAWS,
	"us-east-1d": CloudProviderAWS,
	"us-east-1e": CloudProviderAWS,

	"us-east-2a": CloudProviderAWS,
	"us-east-2b": CloudProviderAWS,
	"us-east-2c": CloudProviderAWS,
	"us-east-2d": CloudProviderAWS,
	"us-east-2e": CloudProviderAWS,

	"us-west-1a": CloudProviderAWS,
	"us-west-1b": CloudProviderAWS,
	"us-west-1c": CloudProviderAWS,
	"us-west-1d": CloudProviderAWS,
	"us-west-1e": CloudProviderAWS,

	"us-west-2a": CloudProviderAWS,
	"us-west-2b": CloudProviderAWS,
	"us-west-2c": CloudProviderAWS,
	"us-west-2d": CloudProviderAWS,
	"us-west-2e": CloudProviderAWS,

	"ca-central-1a": CloudProviderAWS,
	"ca-central-1b": CloudProviderAWS,

	"eu-west-1a": CloudProviderAWS,
	"eu-west-1b": CloudProviderAWS,
	"eu-west-1c": CloudProviderAWS,
	"eu-west-1d": CloudProviderAWS,
	"eu-west-1e": CloudProviderAWS,

	"eu-central-1a": CloudProviderAWS,
	"eu-central-1b": CloudProviderAWS,
	"eu-central-1c": CloudProviderAWS,
	"eu-central-1d": CloudProviderAWS,
	"eu-central-1e": CloudProviderAWS,

	"ap-south-1a": CloudProviderAWS,
	"ap-south-1b": CloudProviderAWS,
	"ap-south-1c": CloudProviderAWS,
	"ap-south-1d": CloudProviderAWS,
	"ap-south-1e": CloudProviderAWS,

	"ap-southeast-1a": CloudProviderAWS,
	"ap-southeast-1b": CloudProviderAWS,
	"ap-southeast-1c": CloudProviderAWS,
	"ap-southeast-1d": CloudProviderAWS,
	"ap-southeast-1e": CloudProviderAWS,

	"ap-southeast-2a": CloudProviderAWS,
	"ap-southeast-2b": CloudProviderAWS,
	"ap-southeast-2c": CloudProviderAWS,
	"ap-southeast-2d": CloudProviderAWS,
	"ap-southeast-2e": CloudProviderAWS,

	"ap-northeast-1a": CloudProviderAWS,
	"ap-northeast-1b": CloudProviderAWS,
	"ap-northeast-1c": CloudProviderAWS,
	"ap-northeast-1d": CloudProviderAWS,
	"ap-northeast-1e": CloudProviderAWS,

	"ap-northeast-2a": CloudProviderAWS,
	"ap-northeast-2b": CloudProviderAWS,
	"ap-northeast-2c": CloudProviderAWS,
	"ap-northeast-2d": CloudProviderAWS,
	"ap-northeast-2e": CloudProviderAWS,

	"sa-east-1a": CloudProviderAWS,
	"sa-east-1b": CloudProviderAWS,
	"sa-east-1c": CloudProviderAWS,
	"sa-east-1d": CloudProviderAWS,
	"sa-east-1e": CloudProviderAWS,

	"cn-north-1a": CloudProviderAWS,
	"cn-north-1b": CloudProviderAWS,

	// GCE
	"asia-east1-a": CloudProviderGCE,
	"asia-east1-b": CloudProviderGCE,
	"asia-east1-c": CloudProviderGCE,
	"asia-east1-d": CloudProviderGCE,

	"asia-northeast1-a": CloudProviderGCE,
	"asia-northeast1-b": CloudProviderGCE,
	"asia-northeast1-c": CloudProviderGCE,
	"asia-northeast1-d": CloudProviderGCE,

	"europe-west1-a": CloudProviderGCE,
	"europe-west1-b": CloudProviderGCE,
	"europe-west1-c": CloudProviderGCE,
	"europe-west1-d": CloudProviderGCE,
	"europe-west1-e": CloudProviderGCE,

	"us-central1-a": CloudProviderGCE,
	"us-central1-b": CloudProviderGCE,
	"us-central1-c": CloudProviderGCE,
	"us-central1-d": CloudProviderGCE,
	"us-central1-e": CloudProviderGCE,
	"us-central1-f": CloudProviderGCE,
	"us-central1-g": CloudProviderGCE,
	"us-central1-h": CloudProviderGCE,

	"us-east1-a": CloudProviderGCE,
	"us-east1-b": CloudProviderGCE,
	"us-east1-c": CloudProviderGCE,
	"us-east1-d": CloudProviderGCE,

	"us-west1-a": CloudProviderGCE,
	"us-west1-b": CloudProviderGCE,
	"us-west1-c": CloudProviderGCE,
	"us-west1-d": CloudProviderGCE,
}

// GuessCloudForZone tries to infer the cloudprovider from the zone name
func GuessCloudForZone(zone string) (CloudProviderID, bool) {
	c, found := zonesToCloud[zone]
	return c, found
}
