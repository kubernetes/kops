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

	// // DeregisterInstance drains a cloud instance and loadbalancers.
	DeregisterInstance(instance *cloudinstances.CloudInstance) error

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
	// (typically GCP or OpenStack load-balancers)
	// +optional
	IP string `json:"ip,omitempty" protobuf:"bytes,1,opt,name=ip"`

	// Hostname is set for load-balancer ingress points that are DNS based
	// (typically AWS load-balancers)
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,2,opt,name=hostname"`
}
