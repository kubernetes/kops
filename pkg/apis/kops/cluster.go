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

package kops

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient=true

// Cluster is a specific cluster wrapper
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSpec `json:"spec,omitempty"`
}

// ClusterList is a list of clusters
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Cluster `json:"items"`
}

// ClusterSpec defines the configuration for a cluster
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
	// ServiceClusterIPRange is the CIDR, from the internal network, where we allocate IPs for services
	ServiceClusterIPRange string `json:"serviceClusterIPRange,omitempty"`
	// NonMasqueradeCIDR is the CIDR for the internal k8s network (on which pods & services live)
	// It cannot overlap ServiceClusterIPRange
	NonMasqueradeCIDR string `json:"nonMasqueradeCIDR,omitempty"`
	// SSHAccess is a list of the CIDRs that can access SSH.
	SSHAccess []string `json:"sshAccess,omitempty"`
	// HTTPProxy defines connection information to support use of a private cluster behind an forward HTTP Proxy
	EgressProxy *EgressProxySpec `json:"egressProxy,omitempty"`

	// KubernetesAPIAccess is a list of the CIDRs that can access the Kubernetes API endpoint (master HTTPS)
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
	// Additional policies to add for roles
	AdditionalPolicies *map[string]string `json:"additionalPolicies,omitempty"`
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
	CloudConfig           *CloudConfiguration          `json:"cloudConfig,omitempty"`

	// Networking configuration
	Networking *NetworkingSpec `json:"networking,omitempty"`

	// API field controls how the API is exposed outside the cluster
	API *AccessSpec `json:"api,omitempty"`
	// Authentication field controls how the cluster is configured for authentication
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
	// Authorization field controls how the cluster is configured for authorization
	Authorization *AuthorizationSpec `json:"authorization,omitempty"`
	// Tags for AWS instance groups
	CloudLabels map[string]string `json:"cloudLabels,omitempty"`
	// Hooks for custom actions e.g. on first installation
	Hooks []HookSpec `json:"hooks,omitempty"`
	// Alternative locations for files and containers
	// This API component is under contruction, will remove this comment
	// once this API is fully functional.
	Assets *Assets `json:"assets,omitempty"`
}

type Assets struct {
	ContainerRegistry *string `json:"containerRegistry,omitempty"`
	FileRepository    *string `json:"fileRepository,omitempty"`
}

type HookSpec struct {
	ExecContainer *ExecContainerAction `json:"execContainer,omitempty"`
}

type ExecContainerAction struct {
	// Docker image name.
	Image string `json:"image,omitempty" `

	Command []string `json:"command,omitempty"`
}

type AuthenticationSpec struct {
	Kopeio *KopeioAuthenticationSpec `json:"kopeio,omitempty"`
}

func (s *AuthenticationSpec) IsEmpty() bool {
	return s.Kopeio == nil
}

type KopeioAuthenticationSpec struct {
}

type AuthorizationSpec struct {
	AlwaysAllow *AlwaysAllowAuthorizationSpec `json:"alwaysAllow,omitempty"`
	RBAC        *RBACAuthorizationSpec        `json:"rbac,omitempty"`
}

func (s *AuthorizationSpec) IsEmpty() bool {
	return s.RBAC == nil && s.AlwaysAllow == nil
}

type RBACAuthorizationSpec struct {
}

type AlwaysAllowAuthorizationSpec struct {
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
	Type               LoadBalancerType `json:"type,omitempty"`
	IdleTimeoutSeconds *int64           `json:"idleTimeoutSeconds,omitempty"`
}

// KubeDNSConfig defines the kube dns configuration
type KubeDNSConfig struct {
	// Image is the name of the docker image to run
	Image string `json:"image,omitempty"`
	// Replicas is the number of pod replicas
	Replicas int `json:"replicas,omitempty"`
	// Domain is the dns domain
	Domain string `json:"domain,omitempty"`
	// ServerIP is the server ip
	ServerIP string `json:"serverIP,omitempty"`
}

// EtcdStorageType defined the etcd storage backend
type EtcdStorageType string

const (
	// EtcdStorageTypeV2 is the old v2 storage
	EtcdStorageTypeV2 EtcdStorageType = "etcd2"
	// EtcdStorageTypeV3 is the new v3 storage
	EtcdStorageTypeV3 EtcdStorageType = "etcd3"
)

// EtcdClusterSpec is the etcd cluster specification
type EtcdClusterSpec struct {
	// Name is the name of the etcd cluster (main, events etc)
	Name string `json:"name,omitempty"`
	// EnableEtcdTLS indicates the etcd service should use TLS between peers and clients
	EnableEtcdTLS bool `json:"enableEtcdTLS,omitempty"`
	// Members stores the configurations for each member of the cluster (including the data volume)
	Members []*EtcdMemberSpec `json:"etcdMembers,omitempty"`
}

// EtcdMemberSpec is a specification for a etcd member
type EtcdMemberSpec struct {
	// Name is the name of the member within the etcd cluster
	Name string `json:"name,omitempty"`
	// InstanceGroup is the instanceGroup this volume is associated
	InstanceGroup *string `json:"instanceGroup,omitempty"`
	// VolumeType is the underlining cloud storage class
	VolumeType *string `json:"volumeType,omitempty"`
	// VolumeSize is the underlining cloud volume size
	VolumeSize *int32 `json:"volumeSize,omitempty"`
	// KmsKeyId is a AWS KMS ID used to encrypt the volume
	KmsKeyId *string `json:"kmsKeyId,omitempty"`
	// EncryptedVolume indicates you want to encrypt the volume
	EncryptedVolume *bool `json:"encryptedVolume,omitempty"`
}

// SubnetType string describes subnet types (public, private, utility)
type SubnetType string

const (
	// SubnetTypePublic means the subnet is public
	SubnetTypePublic SubnetType = "Public"
	// SubnetTypePrivate means the subnet has no public address or is natted
	SubnetTypePrivate SubnetType = "Private"
	// SubnetTypeUtility mean the subnet is used for utility services, such as the bastion
	SubnetTypeUtility SubnetType = "Utility"
)

// ClusterSubnetSpec defines a subnet
type ClusterSubnetSpec struct {
	// Name is the name of the subnet
	Name string `json:"name,omitempty"`
	// CIDR is the network cidr of the subnet
	CIDR string `json:"cidr,omitempty"`
	// Zone is the zone the subnet resides
	Zone string `json:"zone,omitempty"`
	// ProviderID is the cloud provider id for the objects associated with the zone (the subnet on AWS)
	ProviderID string `json:"id,omitempty"`
	// Egress
	Egress string `json:"egress,omitempty"`
	// Type define which one if the internal types (public, utility, private) the network is
	Type SubnetType `json:"type,omitempty"`
}

type EgressProxySpec struct {
	HTTPProxy     HTTPProxy `json:"httpProxy,omitempty"`
	ProxyExcludes string    `json:"excludes,omitempty"`
}

type HTTPProxy struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
	// TODO #3070
	// User     string `json:"user,omitempty"`
	// Password string `json:"password,omitempty"`
}

// FillDefaults populates default values.
// This is different from PerformAssignments, because these values are changeable, and thus we don't need to
// store them (i.e. we don't need to 'lock them')
func (c *Cluster) FillDefaults() error {
	// Topology support
	if c.Spec.Topology == nil {
		c.Spec.Topology = &TopologySpec{Masters: TopologyPublic, Nodes: TopologyPublic}
		c.Spec.Topology.DNS = &DNSSpec{Type: DNSTypePublic}
	}

	if c.Spec.Networking == nil {
		c.Spec.Networking = &NetworkingSpec{}
	}

	// TODO move this into networking.go :(
	if c.Spec.Networking.Classic != nil {
		// OK
	} else if c.Spec.Networking.Kubenet != nil {
		// OK
	} else if c.Spec.Networking.CNI != nil {
		// OK
	} else if c.Spec.Networking.External != nil {
		// OK
	} else if c.Spec.Networking.Kopeio != nil {
		// OK
	} else if c.Spec.Networking.Weave != nil {
		// OK
	} else if c.Spec.Networking.Flannel != nil {
		// OK
	} else if c.Spec.Networking.Calico != nil {
		// OK
	} else if c.Spec.Networking.Canal != nil {
		// OK
	} else if c.Spec.Networking.Kuberouter != nil {
		// OK
	} else {
		// No networking model selected; choose Kubenet
		c.Spec.Networking.Kubenet = &KubenetNetworkingSpec{}
	}

	if c.Spec.Channel == "" {
		c.Spec.Channel = DefaultChannel
	}

	if c.ObjectMeta.Name == "" {
		return fmt.Errorf("cluster Name not set in FillDefaults")
	}

	if c.Spec.MasterInternalName == "" {
		c.Spec.MasterInternalName = "api.internal." + c.ObjectMeta.Name
	}

	if c.Spec.MasterPublicName == "" {
		c.Spec.MasterPublicName = "api." + c.ObjectMeta.Name
	}

	return nil
}

// SharedVPC is a simple helper function which makes the templates for a shared VPC clearer
func (c *Cluster) SharedVPC() bool {
	return c.Spec.NetworkID != ""
}
