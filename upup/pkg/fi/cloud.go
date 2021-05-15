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

	// FindClusterStatus discovers the status of the cluster, by inspecting the cloud objects
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)

	GetApiIngressStatus(cluster *kops.Cluster) ([]ApiIngressStatus, error)
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

// ApiIngressStatus represents the status of an ingress point:
// traffic intended for the service should be sent to an ingress point.
type ApiIngressStatus struct {
	// IP is set for load-balancer ingress points that are IP based
	// (typically GCE or OpenStack load-balancers)
	// +optional
	IP string `json:"ip,omitempty" protobuf:"bytes,1,opt,name=ip"`

	// Hostname is set for load-balancer ingress points that are DNS based
	// (typically AWS load-balancers)
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,2,opt,name=hostname"`
}

// zonesToCloud allows us to infer from certain well-known zones to a cloud
// Note it is safe to "overmap" zones that don't exist: we'll check later if the zones actually exist
var zonesToCloud = map[string]kops.CloudProviderID{
	"us-east-1a": kops.CloudProviderAWS,
	"us-east-1b": kops.CloudProviderAWS,
	"us-east-1c": kops.CloudProviderAWS,
	"us-east-1d": kops.CloudProviderAWS,
	"us-east-1e": kops.CloudProviderAWS,
	"us-east-1f": kops.CloudProviderAWS,

	"us-east-2a": kops.CloudProviderAWS,
	"us-east-2b": kops.CloudProviderAWS,
	"us-east-2c": kops.CloudProviderAWS,
	"us-east-2d": kops.CloudProviderAWS,
	"us-east-2e": kops.CloudProviderAWS,
	"us-east-2f": kops.CloudProviderAWS,

	"us-west-1a": kops.CloudProviderAWS,
	"us-west-1b": kops.CloudProviderAWS,
	"us-west-1c": kops.CloudProviderAWS,
	"us-west-1d": kops.CloudProviderAWS,
	"us-west-1e": kops.CloudProviderAWS,
	"us-west-1f": kops.CloudProviderAWS,

	"us-west-2a": kops.CloudProviderAWS,
	"us-west-2b": kops.CloudProviderAWS,
	"us-west-2c": kops.CloudProviderAWS,
	"us-west-2d": kops.CloudProviderAWS,
	"us-west-2e": kops.CloudProviderAWS,
	"us-west-2f": kops.CloudProviderAWS,

	"ca-central-1a": kops.CloudProviderAWS,
	"ca-central-1b": kops.CloudProviderAWS,
	"ca-central-1c": kops.CloudProviderAWS,
	"ca-central-1d": kops.CloudProviderAWS,

	"eu-north-1a": kops.CloudProviderAWS,
	"eu-north-1b": kops.CloudProviderAWS,
	"eu-north-1c": kops.CloudProviderAWS,

	"eu-west-1a": kops.CloudProviderAWS,
	"eu-west-1b": kops.CloudProviderAWS,
	"eu-west-1c": kops.CloudProviderAWS,
	"eu-west-1d": kops.CloudProviderAWS,
	"eu-west-1e": kops.CloudProviderAWS,

	"eu-west-2a": kops.CloudProviderAWS,
	"eu-west-2b": kops.CloudProviderAWS,
	"eu-west-2c": kops.CloudProviderAWS,

	"eu-west-3a": kops.CloudProviderAWS,
	"eu-west-3b": kops.CloudProviderAWS,
	"eu-west-3c": kops.CloudProviderAWS,

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

	"ap-northeast-3a": kops.CloudProviderAWS,
	"ap-northeast-3b": kops.CloudProviderAWS,
	"ap-northeast-3c": kops.CloudProviderAWS,
	"ap-northeast-3d": kops.CloudProviderAWS,
	"ap-northeast-3e": kops.CloudProviderAWS,

	"ap-east-1a": kops.CloudProviderAWS,
	"ap-east-1b": kops.CloudProviderAWS,
	"ap-east-1c": kops.CloudProviderAWS,
	"ap-east-1d": kops.CloudProviderAWS,
	"ap-east-1e": kops.CloudProviderAWS,

	"sa-east-1a": kops.CloudProviderAWS,
	"sa-east-1b": kops.CloudProviderAWS,
	"sa-east-1c": kops.CloudProviderAWS,
	"sa-east-1d": kops.CloudProviderAWS,
	"sa-east-1e": kops.CloudProviderAWS,

	"cn-north-1a": kops.CloudProviderAWS,
	"cn-north-1b": kops.CloudProviderAWS,

	"cn-northwest-1a": kops.CloudProviderAWS,
	"cn-northwest-1b": kops.CloudProviderAWS,
	"cn-northwest-1c": kops.CloudProviderAWS,

	"me-south-1a": kops.CloudProviderAWS,
	"me-south-1b": kops.CloudProviderAWS,
	"me-south-1c": kops.CloudProviderAWS,

	"us-gov-east-1a": kops.CloudProviderAWS,
	"us-gov-east-1b": kops.CloudProviderAWS,
	"us-gov-east-1c": kops.CloudProviderAWS,

	"us-gov-west-1a": kops.CloudProviderAWS,
	"us-gov-west-1b": kops.CloudProviderAWS,
	"us-gov-west-1c": kops.CloudProviderAWS,

	"af-south-1a": kops.CloudProviderAWS,
	"af-south-1b": kops.CloudProviderAWS,
	"af-south-1c": kops.CloudProviderAWS,

	"eu-south-1a": kops.CloudProviderAWS,
	"eu-south-1b": kops.CloudProviderAWS,
	"eu-south-1c": kops.CloudProviderAWS,

	// GCE
	"asia-east1-a": kops.CloudProviderGCE,
	"asia-east1-b": kops.CloudProviderGCE,
	"asia-east1-c": kops.CloudProviderGCE,
	"asia-east1-d": kops.CloudProviderGCE,

	"asia-east2-a": kops.CloudProviderGCE,
	"asia-east2-b": kops.CloudProviderGCE,
	"asia-east2-c": kops.CloudProviderGCE,

	"asia-northeast1-a": kops.CloudProviderGCE,
	"asia-northeast1-b": kops.CloudProviderGCE,
	"asia-northeast1-c": kops.CloudProviderGCE,
	"asia-northeast1-d": kops.CloudProviderGCE,

	"asia-northeast2-a": kops.CloudProviderGCE,
	"asia-northeast2-b": kops.CloudProviderGCE,
	"asia-northeast2-c": kops.CloudProviderGCE,

	"asia-northeast3-a": kops.CloudProviderGCE,
	"asia-northeast3-b": kops.CloudProviderGCE,
	"asia-northeast3-c": kops.CloudProviderGCE,

	"asia-south1-a": kops.CloudProviderGCE,
	"asia-south1-b": kops.CloudProviderGCE,
	"asia-south1-c": kops.CloudProviderGCE,

	"asia-southeast1-a": kops.CloudProviderGCE,
	"asia-southeast1-b": kops.CloudProviderGCE,

	"asia-southeast2-a": kops.CloudProviderGCE,
	"asia-southeast2-b": kops.CloudProviderGCE,
	"asia-southeast2-c": kops.CloudProviderGCE,

	"australia-southeast1-a": kops.CloudProviderGCE,
	"australia-southeast1-b": kops.CloudProviderGCE,
	"australia-southeast1-c": kops.CloudProviderGCE,

	"europe-north1-a": kops.CloudProviderGCE,
	"europe-north1-b": kops.CloudProviderGCE,
	"europe-north1-c": kops.CloudProviderGCE,

	"europe-west1-a": kops.CloudProviderGCE,
	"europe-west1-b": kops.CloudProviderGCE,
	"europe-west1-c": kops.CloudProviderGCE,
	"europe-west1-d": kops.CloudProviderGCE,
	"europe-west1-e": kops.CloudProviderGCE,

	"europe-west2-a": kops.CloudProviderGCE,
	"europe-west2-b": kops.CloudProviderGCE,
	"europe-west2-c": kops.CloudProviderGCE,

	"europe-west3-a": kops.CloudProviderGCE,
	"europe-west3-b": kops.CloudProviderGCE,
	"europe-west3-c": kops.CloudProviderGCE,

	"europe-west4-a": kops.CloudProviderGCE,
	"europe-west4-b": kops.CloudProviderGCE,
	"europe-west4-c": kops.CloudProviderGCE,

	"europe-west6-a": kops.CloudProviderGCE,
	"europe-west6-b": kops.CloudProviderGCE,
	"europe-west6-c": kops.CloudProviderGCE,

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

	"us-east4-a": kops.CloudProviderGCE,
	"us-east4-b": kops.CloudProviderGCE,
	"us-east4-c": kops.CloudProviderGCE,

	"us-west1-a": kops.CloudProviderGCE,
	"us-west1-b": kops.CloudProviderGCE,
	"us-west1-c": kops.CloudProviderGCE,
	"us-west1-d": kops.CloudProviderGCE,

	"us-west2-a": kops.CloudProviderGCE,
	"us-west2-b": kops.CloudProviderGCE,
	"us-west2-c": kops.CloudProviderGCE,

	"us-west3-a": kops.CloudProviderGCE,
	"us-west3-b": kops.CloudProviderGCE,
	"us-west3-c": kops.CloudProviderGCE,

	"us-west4-a": kops.CloudProviderGCE,
	"us-west4-b": kops.CloudProviderGCE,
	"us-west4-c": kops.CloudProviderGCE,

	"northamerica-northeast1-a": kops.CloudProviderGCE,
	"northamerica-northeast1-b": kops.CloudProviderGCE,
	"northamerica-northeast1-c": kops.CloudProviderGCE,

	"southamerica-east1-a": kops.CloudProviderGCE,
	"southamerica-east1-b": kops.CloudProviderGCE,
	"southamerica-east1-c": kops.CloudProviderGCE,

	"nyc1": kops.CloudProviderDO,
	"nyc2": kops.CloudProviderDO,
	"nyc3": kops.CloudProviderDO,

	"sfo1": kops.CloudProviderDO,
	"sfo2": kops.CloudProviderDO,
	"sfo3": kops.CloudProviderDO,

	"ams2": kops.CloudProviderDO,
	"ams3": kops.CloudProviderDO,

	"tor1": kops.CloudProviderDO,

	"sgp1": kops.CloudProviderDO,

	"lon1": kops.CloudProviderDO,

	"fra1": kops.CloudProviderDO,

	"blr1": kops.CloudProviderDO,

	"cn-qingdao-b": kops.CloudProviderALI,
	"cn-qingdao-c": kops.CloudProviderALI,

	"cn-beijing-a": kops.CloudProviderALI,
	"cn-beijing-b": kops.CloudProviderALI,
	"cn-beijing-c": kops.CloudProviderALI,
	"cn-beijing-d": kops.CloudProviderALI,
	"cn-beijing-e": kops.CloudProviderALI,

	"cn-zhangjiakou-a": kops.CloudProviderALI,

	"cn-huhehaote-a": kops.CloudProviderALI,

	"cn-hangzhou-b": kops.CloudProviderALI,
	"cn-hangzhou-c": kops.CloudProviderALI,
	"cn-hangzhou-d": kops.CloudProviderALI,
	"cn-hangzhou-e": kops.CloudProviderALI,
	"cn-hangzhou-f": kops.CloudProviderALI,
	"cn-hangzhou-g": kops.CloudProviderALI,

	"cn-shanghai-a": kops.CloudProviderALI,
	"cn-shanghai-b": kops.CloudProviderALI,
	"cn-shanghai-c": kops.CloudProviderALI,
	"cn-shanghai-d": kops.CloudProviderALI,

	"cn-shenzhen-a": kops.CloudProviderALI,
	"cn-shenzhen-b": kops.CloudProviderALI,
	"cn-shenzhen-c": kops.CloudProviderALI,

	"cn-hongkong-a": kops.CloudProviderALI,
	"cn-hongkong-b": kops.CloudProviderALI,
	"cn-hongkong-c": kops.CloudProviderALI,

	"asia":                kops.CloudProviderAzure,
	"asiapacific":         kops.CloudProviderAzure,
	"australia":           kops.CloudProviderAzure,
	"australiacentral":    kops.CloudProviderAzure,
	"australiacentral2":   kops.CloudProviderAzure,
	"australiaeast":       kops.CloudProviderAzure,
	"australiasoutheast":  kops.CloudProviderAzure,
	"brazil":              kops.CloudProviderAzure,
	"brazilsouth":         kops.CloudProviderAzure,
	"brazilsoutheast":     kops.CloudProviderAzure,
	"canada":              kops.CloudProviderAzure,
	"canadacentral":       kops.CloudProviderAzure,
	"canadaeast":          kops.CloudProviderAzure,
	"centralindia":        kops.CloudProviderAzure,
	"centralus":           kops.CloudProviderAzure,
	"centraluseuap":       kops.CloudProviderAzure,
	"centralusstage":      kops.CloudProviderAzure,
	"eastasia":            kops.CloudProviderAzure,
	"eastasiastage":       kops.CloudProviderAzure,
	"eastus":              kops.CloudProviderAzure,
	"eastus2":             kops.CloudProviderAzure,
	"eastus2euap":         kops.CloudProviderAzure,
	"eastus2stage":        kops.CloudProviderAzure,
	"eastusstage":         kops.CloudProviderAzure,
	"europe":              kops.CloudProviderAzure,
	"francecentral":       kops.CloudProviderAzure,
	"francesouth":         kops.CloudProviderAzure,
	"germanynorth":        kops.CloudProviderAzure,
	"germanywestcentral":  kops.CloudProviderAzure,
	"global":              kops.CloudProviderAzure,
	"india":               kops.CloudProviderAzure,
	"japan":               kops.CloudProviderAzure,
	"japaneast":           kops.CloudProviderAzure,
	"japanwest":           kops.CloudProviderAzure,
	"koreacentral":        kops.CloudProviderAzure,
	"koreasouth":          kops.CloudProviderAzure,
	"northcentralus":      kops.CloudProviderAzure,
	"northcentralusstage": kops.CloudProviderAzure,
	"northeurope":         kops.CloudProviderAzure,
	"norwayeast":          kops.CloudProviderAzure,
	"norwaywest":          kops.CloudProviderAzure,
	"southafricanorth":    kops.CloudProviderAzure,
	"southafricawest":     kops.CloudProviderAzure,
	"southcentralus":      kops.CloudProviderAzure,
	"southcentralusstage": kops.CloudProviderAzure,
	"southeastasia":       kops.CloudProviderAzure,
	"southeastasiastage":  kops.CloudProviderAzure,
	"southindia":          kops.CloudProviderAzure,
	"switzerlandnorth":    kops.CloudProviderAzure,
	"switzerlandwest":     kops.CloudProviderAzure,
	"uaecentral":          kops.CloudProviderAzure,
	"uaenorth":            kops.CloudProviderAzure,
	"uk":                  kops.CloudProviderAzure,
	"uksouth":             kops.CloudProviderAzure,
	"ukwest":              kops.CloudProviderAzure,
	"unitedstates":        kops.CloudProviderAzure,
	"westcentralus":       kops.CloudProviderAzure,
	"westeurope":          kops.CloudProviderAzure,
	"westindia":           kops.CloudProviderAzure,
	"westus":              kops.CloudProviderAzure,
	"westus2":             kops.CloudProviderAzure,
	"westus2stage":        kops.CloudProviderAzure,
	"westusstage":         kops.CloudProviderAzure,
}

// GuessCloudForZone tries to infer the cloudprovider from the zone name
// Ali has the same zoneNames as AWS in the regions outside China, so if use AliCloud to install k8s in the regions outside China,
// the users need to provide parameter "--cloud". But the regions inside China can be easily identified.
func GuessCloudForZone(zone string) (kops.CloudProviderID, bool) {
	c, found := zonesToCloud[zone]
	return c, found
}
