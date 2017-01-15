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

package v1alpha2

import (
	"k8s.io/kubernetes/pkg/api/v1"
	meta_v1 "k8s.io/kubernetes/pkg/apis/meta/v1"
)

type Cluster struct {
	meta_v1.TypeMeta `json:",inline"`
	ObjectMeta       v1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSpec `json:"spec,omitempty"`
}

type ClusterList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata,omitempty"`

	Items []Cluster `json:"items"`
}

type ClusterSpec struct {
	// The Channel we are following
	Channel string `json:"channel,omitempty"`

	// ConfigBase is the path where we store configuration for the cluster
	// This might be different that the location when the cluster spec itself is stored,
	// both because this must be accessible to the cluster,
	// and because it might be on a different cloud or storage system (etcd vs S3)
	ConfigBase string `json:"configBase,omitempty"`

	// The CloudProvider to use (aws or gce)
	CloudProvider string `json:"cloudProvider,omitempty"`

	// The version of kubernetes to install (optional, and can be a "spec" like stable)
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	//
	//// The Node initializer technique to use: cloudinit or nodeup
	//NodeInit                      string `json:",omitempty"`

	// Configuration of subnets we are targeting
	Subnets []ClusterSubnetSpec `json:"subnets,omitempty"`

	// Project is the cloud project we should use, required on GCE
	Project string `json:"project,omitempty"`

	// MasterPublicName is the external DNS name for the master nodes
	MasterPublicName string `json:"masterPublicName,omitempty"`
	// MasterInternalName is the internal DNS name for the master nodes
	MasterInternalName string `json:"masterInternalName,omitempty"`

	// The CIDR used for the AWS VPC / GCE Network, or otherwise allocated to k8s
	// This is a real CIDR, not the internal k8s network
	NetworkCIDR string `json:"networkCIDR,omitempty"`

	// NetworkID is an identifier of a network, if we want to reuse/share an existing network (e.g. an AWS VPC)
	NetworkID string `json:"networkID,omitempty"`

	// Topology defines the type of network topology to use on the cluster - default public
	// This is heavily weighted towards AWS for the time being, but should also be agnostic enough
	// to port out to GCE later if needed
	Topology *TopologySpec `json:"topology,omitempty"`

	// SecretStore is the VFS path to where secrets are stored
	SecretStore string `json:"secretStore,omitempty"`
	// KeyStore is the VFS path to where SSL keys and certificates are stored
	KeyStore string `json:"keyStore,omitempty"`
	// ConfigStore is the VFS path to where the configuration (Cluster, InstanceGroups etc) is stored
	ConfigStore string `json:"configStore,omitempty"`

	// DNSZone is the DNS zone we should use when configuring DNS
	// This is because some clouds let us define a managed zone foo.bar, and then have
	// kubernetes.dev.foo.bar, without needing to define dev.foo.bar as a hosted zone.
	// DNSZone will probably be a suffix of the MasterPublicName and MasterInternalName
	// Note that DNSZone can either by the host name of the zone (containing dots),
	// or can be an identifier for the zone.
	DNSZone string `json:"dnsZone,omitempty"`

	// ClusterDNSDomain is the suffix we use for internal DNS names (normally cluster.local)
	ClusterDNSDomain string `json:"clusterDNSDomain,omitempty"`

	//InstancePrefix                string `json:",omitempty"`

	// ClusterName is a unique identifier for the cluster, and currently must be a DNS name
	//ClusterName       string `json:",omitempty"`

	//ClusterIPRange                string `json:",omitempty"`

	// ServiceClusterIPRange is the CIDR, from the internal network, where we allocate IPs for services
	ServiceClusterIPRange string `json:"serviceClusterIPRange,omitempty"`
	//MasterIPRange                 string `json:",omitempty"`

	// NonMasqueradeCIDR is the CIDR for the internal k8s network (on which pods & services live)
	// It cannot overlap ServiceClusterIPRange
	NonMasqueradeCIDR string `json:"nonMasqueradeCIDR,omitempty"`

	// SSHAccess determines the permitted access to SSH
	// Currently only a single CIDR is supported (though a richer grammar could be added in future)
	SSHAccess []string `json:"sshAccess,omitempty"`

	// KubernetesAPIAccess determines the permitted access to the API endpoints (master HTTPS)
	// Currently only a single CIDR is supported (though a richer grammar could be added in future)
	KubernetesAPIAccess []string `json:"kubernetesApiAccess,omitempty"`

	// IsolatesMasters determines whether we should lock down masters so that they are not on the pod network.
	// true is the kube-up behaviour, but it is very surprising: it means that daemonsets only work on the master
	// if they have hostNetwork=true.
	// false is now the default, and it will:
	//  * give the master a normal PodCIDR
	//  * run kube-proxy on the master
	//  * enable debugging handlers on the master, so kubectl logs works
	IsolateMasters *bool `json:"isolateMasters,omitempty"`

	// UpdatePolicy determines the policy for applying upgrades automatically.
	// Valid values:
	//   'external' do not apply updates automatically - they are applied manually or by an external system
	//   missing: default policy (currently OS security upgrades that do not require a reboot)
	UpdatePolicy *string `json:"updatePolicy,omitempty"`

	// EtcdClusters stores the configuration for each cluster
	EtcdClusters []*EtcdClusterSpec `json:"etcdClusters,omitempty"`

	// Component configurations
	Docker                *DockerConfig                `json:"docker,omitempty"`
	KubeDNS               *KubeDNSConfig               `json:"kubeDNS,omitempty"`
	KubeAPIServer         *KubeAPIServerConfig         `json:"kubeAPIServer,omitempty"`
	KubeControllerManager *KubeControllerManagerConfig `json:"kubeControllerManager,omitempty"`
	KubeScheduler         *KubeSchedulerConfig         `json:"kubeScheduler,omitempty"`
	KubeProxy             *KubeProxyConfig             `json:"kubeProxy,omitempty"`
	Kubelet               *KubeletConfigSpec           `json:"kubelet,omitempty"`
	MasterKubelet         *KubeletConfigSpec           `json:"masterKubelet,omitempty"`

	// Networking configuration
	Networking *NetworkingSpec `json:"networking,omitempty"`

	// API field controls how the API is exposed outside the cluster
	API *AccessSpec `json:"api,omitempty"`
}

type AccessSpec struct {
	DNS          *DNSAccessSpec          `json:"dns,omitempty"`
	LoadBalancer *LoadBalancerAccessSpec `json:"loadBalancer,omitempty"`
}

func (s *AccessSpec) IsEmpty() bool {
	return s.DNS == nil && s.LoadBalancer == nil
}

type DNSAccessSpec struct {
}

// LoadBalancerType string describes LoadBalancer types (public, internal)
type LoadBalancerType string

const (
	LoadBalancerTypePublic   LoadBalancerType = "Public"
	LoadBalancerTypeInternal LoadBalancerType = "Internal"
)

type LoadBalancerAccessSpec struct {
	Type LoadBalancerType `json:"type,omitempty"`
}

type KubeDNSConfig struct {
	// Image is the name of the docker image to run
	Image string `json:"image,omitempty"`

	Replicas int    `json:"replicas,omitempty"`
	Domain   string `json:"domain,omitempty"`
	ServerIP string `json:"serverIP,omitempty"`
}

type EtcdClusterSpec struct {
	// Name is the name of the etcd cluster (main, events etc)
	Name string `json:"name,omitempty"`

	// EtcdMember stores the configurations for each member of the cluster (including the data volume)
	Members []*EtcdMemberSpec `json:"etcdMembers,omitempty"`
}

type EtcdMemberSpec struct {
	// Name is the name of the member within the etcd cluster
	Name          string  `json:"name,omitempty"`
	InstanceGroup *string `json:"instanceGroup,omitempty"`

	VolumeType      *string `json:"volumeType,omitempty"`
	VolumeSize      *int32  `json:"volumeSize,omitempty"`
	KmsKeyId        *string `json:"kmsKeyId,omitempty"`
	EncryptedVolume *bool   `json:"encryptedVolume,omitempty"`
}

// SubnetType string describes subnet types (public, private, utility)
type SubnetType string

const (
	SubnetTypePublic  SubnetType = "Public"
	SubnetTypePrivate SubnetType = "Private"
	SubnetTypeUtility SubnetType = "Utility"
)

type ClusterSubnetSpec struct {
	Name string `json:"name,omitempty"`

	Zone string `json:"zone,omitempty"`

	CIDR string `json:"cidr,omitempty"`

	// ProviderID is the cloud provider id for the objects associated with the zone (the subnet on AWS)
	ProviderID string `json:"id,omitempty"`

	Type SubnetType `json:"type,omitempty"`
}
