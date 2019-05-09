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

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

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
	// Additional addons that should be installed on the cluster
	Addons []AddonSpec `json:"addons,omitempty"`
	// ConfigBase is the path where we store configuration for the cluster
	// This might be different that the location when the cluster spec itself is stored,
	// both because this must be accessible to the cluster,
	// and because it might be on a different cloud or storage system (etcd vs S3)
	ConfigBase string `json:"configBase,omitempty"`
	// The CloudProvider to use (aws or gce)
	CloudProvider string `json:"cloudProvider,omitempty"`
	// GossipConfig for the cluster assuming the use of gossip DNS
	GossipConfig *GossipConfig `json:"gossipConfig,omitempty"`
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
	// NetworkCIDR is the CIDR used for the AWS VPC / GCE Network, or otherwise allocated to k8s
	// This is a real CIDR, not the internal k8s network
	// On AWS, it maps to the VPC CIDR.  It is not required on GCE.
	NetworkCIDR string `json:"networkCIDR,omitempty"`
	// AdditionalNetworkCIDRs is a list of additional CIDR used for the AWS VPC
	// or otherwise allocated to k8s. This is a real CIDR, not the internal k8s network
	// On AWS, it maps to any additional CIDRs added to a VPC.
	AdditionalNetworkCIDRs []string `json:"additionalNetworkCIDRs,omitempty"`
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
	// DNSControllerGossipConfig for the cluster assuming the use of gossip DNS
	DNSControllerGossipConfig *DNSControllerGossipConfig `json:"dnsControllerGossipConfig,omitempty"`
	// AdditionalSANs adds additional Subject Alternate Names to apiserver cert that kops generates
	AdditionalSANs []string `json:"additionalSans,omitempty"`
	// ClusterDNSDomain is the suffix we use for internal DNS names (normally cluster.local)
	ClusterDNSDomain string `json:"clusterDNSDomain,omitempty"`
	// ServiceClusterIPRange is the CIDR, from the internal network, where we allocate IPs for services
	ServiceClusterIPRange string `json:"serviceClusterIPRange,omitempty"`
	// PodCIDR is the CIDR from which we allocate IPs for pods
	PodCIDR string `json:"podCIDR,omitempty"`
	//MasterIPRange                 string `json:",omitempty"`
	// NonMasqueradeCIDR is the CIDR for the internal k8s network (on which pods & services live)
	// It cannot overlap ServiceClusterIPRange
	NonMasqueradeCIDR string `json:"nonMasqueradeCIDR,omitempty"`
	// SSHAccess determines the permitted access to SSH
	// Currently only a single CIDR is supported (though a richer grammar could be added in future)
	SSHAccess []string `json:"sshAccess,omitempty"`
	// NodePortAccess is a list of the CIDRs that can access the node ports range (30000-32767).
	NodePortAccess []string `json:"nodePortAccess,omitempty"`
	// HTTPProxy defines connection information to support use of a private cluster behind an forward HTTP Proxy
	EgressProxy *EgressProxySpec `json:"egressProxy,omitempty"`
	// SSHKeyName specifies a preexisting SSH key to use
	SSHKeyName *string `json:"sshKeyName,omitempty"`
	// KubernetesAPIAccess determines the permitted access to the API endpoints (master HTTPS)
	// Currently only a single CIDR is supported (though a richer grammar could be added in future)
	KubernetesAPIAccess []string `json:"kubernetesApiAccess,omitempty"`
	// IsolateMasters determines whether we should lock down masters so that they are not on the pod network.
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
	// A collection of files assets for deployed cluster wide
	FileAssets []FileAssetSpec `json:"fileAssets,omitempty"`
	// EtcdClusters stores the configuration for each cluster
	EtcdClusters []*EtcdClusterSpec `json:"etcdClusters,omitempty"`
	// Component configurations
	Docker                         *DockerConfig                 `json:"docker,omitempty"`
	KubeDNS                        *KubeDNSConfig                `json:"kubeDNS,omitempty"`
	KubeAPIServer                  *KubeAPIServerConfig          `json:"kubeAPIServer,omitempty"`
	KubeControllerManager          *KubeControllerManagerConfig  `json:"kubeControllerManager,omitempty"`
	ExternalCloudControllerManager *CloudControllerManagerConfig `json:"cloudControllerManager,omitempty"`
	KubeScheduler                  *KubeSchedulerConfig          `json:"kubeScheduler,omitempty"`
	KubeProxy                      *KubeProxyConfig              `json:"kubeProxy,omitempty"`
	Kubelet                        *KubeletConfigSpec            `json:"kubelet,omitempty"`
	MasterKubelet                  *KubeletConfigSpec            `json:"masterKubelet,omitempty"`
	CloudConfig                    *CloudConfiguration           `json:"cloudConfig,omitempty"`
	ExternalDNS                    *ExternalDNSConfig            `json:"externalDns,omitempty"`
	// Networking configuration
	Networking *NetworkingSpec `json:"networking,omitempty"`
	// API field controls how the API is exposed outside the cluster
	API *AccessSpec `json:"api,omitempty"`
	// Authentication field controls how the cluster is configured for authentication
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
	// Authorization field controls how the cluster is configured for authorization
	Authorization *AuthorizationSpec `json:"authorization,omitempty"`
	// NodeAuthorization defined the custom node authorization configuration
	NodeAuthorization *NodeAuthorizationSpec `json:"nodeAuthorization,omitempty"`
	// Tags for AWS resources
	CloudLabels map[string]string `json:"cloudLabels,omitempty"`
	// Hooks for custom actions e.g. on first installation
	Hooks []HookSpec `json:"hooks,omitempty"`
	// Alternative locations for files and containers
	Assets *Assets `json:"assets,omitempty"`
	// IAM field adds control over the IAM security policies applied to resources
	IAM *IAMSpec `json:"iam,omitempty"`
	// EncryptionConfig holds the encryption config
	EncryptionConfig *bool `json:"encryptionConfig,omitempty"`
	// DisableSubnetTags controls if subnets are tagged in AWS
	DisableSubnetTags bool `json:"DisableSubnetTags,omitempty"`
	// Target allows for us to nest extra config for targets such as terraform
	Target *TargetSpec `json:"target,omitempty"`
	// UseHostCertificates will mount /etc/ssl/certs to inside needed containers.
	// This is needed if some APIs do have self-signed certs
	UseHostCertificates *bool `json:"useHostCertificates,omitempty"`
}

// NodeAuthorizationSpec is used to node authorization
type NodeAuthorizationSpec struct {
	// NodeAuthorizer defined the configuration for the node authorizer
	NodeAuthorizer *NodeAuthorizerSpec `json:"nodeAuthorizer,omitempty"`
}

// NodeAuthorizerSpec defines the configuration for a node authorizer
type NodeAuthorizerSpec struct {
	// Authorizer is the authorizer to use
	Authorizer string `json:"authorizer,omitempty"`
	// Features is a series of authorizer features to enable or disable
	Features *[]string `json:"features,omitempty"`
	// Image is the location of container
	Image string `json:"image,omitempty"`
	// NodeURL is the node authorization service url
	NodeURL string `json:"nodeURL,omitempty"`
	// Port is the port the service is running on the master
	Port int `json:"port,omitempty"`
	// Interval the time between retires for authorization request
	Interval *metav1.Duration `json:"interval,omitempty"`
	// Timeout the max time for authorization request
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// TokenTTL is the max ttl for an issued token
	TokenTTL *metav1.Duration `json:"tokenTTL,omitempty"`
}

// AddonSpec defines an addon that we want to install in the cluster
type AddonSpec struct {
	// Manifest is a path to the manifest that defines the addon
	Manifest string `json:"manifest,omitempty"`
}

// FileAssetSpec defines the structure for a file asset
type FileAssetSpec struct {
	// Name is a shortened reference to the asset
	Name string `json:"name,omitempty"`
	// Path is the location this file should reside
	Path string `json:"path,omitempty"`
	// Roles is a list of roles the file asset should be applied, defaults to all
	Roles []InstanceGroupRole `json:"roles,omitempty"`
	// Content is the contents of the file
	Content string `json:"content,omitempty"`
	// IsBase64 indicates the contents is base64 encoded
	IsBase64 bool `json:"isBase64,omitempty"`
}

// Assets defined the privately hosted assets
type Assets struct {
	// ContainerRegistry is a url for to a docker registry
	ContainerRegistry *string `json:"containerRegistry,omitempty"`
	// FileRepository is the url for a private file serving repository
	FileRepository *string `json:"fileRepository,omitempty"`
	// ContainerProxy is a url for a pull-through proxy of a docker registry
	ContainerProxy *string `json:"containerProxy,omitempty"`
}

// IAMSpec adds control over the IAM security policies applied to resources
type IAMSpec struct {
	Legacy                 bool `json:"legacy"`
	AllowContainerRegistry bool `json:"allowContainerRegistry,omitempty"`
}

// HookSpec is a definition hook
type HookSpec struct {
	// Name is an optional name for the hook, otherwise the name is kops-hook-<index>
	Name string `json:"name,omitempty"`
	// Disabled indicates if you want the unit switched off
	Disabled bool `json:"disabled,omitempty"`
	// Roles is an optional list of roles the hook should be rolled out to, defaults to all
	Roles []InstanceGroupRole `json:"roles,omitempty"`
	// Requires is a series of systemd units the action requires
	Requires []string `json:"requires,omitempty"`
	// Before is a series of systemd units which this hook must run before
	Before []string `json:"before,omitempty"`
	// ExecContainer is the image itself
	ExecContainer *ExecContainerAction `json:"execContainer,omitempty"`
	// Manifest is a raw systemd unit file
	Manifest string `json:"manifest,omitempty"`
	// UseRawManifest indicates that the contents of Manifest should be used as the contents
	// of the systemd unit, unmodified. Before and Requires are ignored when used together
	// with this value (and validation shouldn't allow them to be set)
	UseRawManifest bool `json:"useRawManifest,omitempty"`
}

// ExecContainerAction defines an hood action
type ExecContainerAction struct {
	// Image is the docker image
	Image string `json:"image,omitempty" `
	// Command is the command supplied to the above image
	Command []string `json:"command,omitempty"`
	// Environment is a map of environment variables added to the hook
	Environment map[string]string `json:"environment,omitempty"`
}

type AuthenticationSpec struct {
	Kopeio *KopeioAuthenticationSpec `json:"kopeio,omitempty"`
	Aws    *AwsAuthenticationSpec    `json:"aws,omitempty"`
}

func (s *AuthenticationSpec) IsEmpty() bool {
	return s.Kopeio == nil && s.Aws == nil
}

type KopeioAuthenticationSpec struct {
}

type AwsAuthenticationSpec struct {
	// Image is the AWS IAM Authenticator docker image to uses
	Image string `json:"image,omitempty"`
	// MemoryRequest memory request of AWS IAM Authenticator container. Default 20Mi
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest CPU request of AWS IAM Authenticator container. Default 10m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// MemoryLimit memory limit of AWS IAM Authenticator container. Default 20Mi
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
	// CPULimit CPU limit of AWS IAM Authenticator container. Default 10m
	CPULimit *resource.Quantity `json:"cpuLimit,omitempty"`
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

// AccessSpec provides configuration details related to kubeapi dns and ELB access
type AccessSpec struct {
	// DNS will be used to provide config on kube-apiserver ELB DNS
	DNS *DNSAccessSpec `json:"dns,omitempty"`
	// LoadBalancer is the configuration for the kube-apiserver ELB
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

// LoadBalancerAccessSpec provides configuration details related to API LoadBalancer and its access
type LoadBalancerAccessSpec struct {
	// Type of load balancer to create may Public or Internal.
	Type LoadBalancerType `json:"type,omitempty"`
	// IdleTimeoutSeconds sets the timeout of the api loadbalancer.
	IdleTimeoutSeconds *int64 `json:"idleTimeoutSeconds,omitempty"`
	// SecurityGroupOverride overrides the default Kops created SG for the load balancer.
	SecurityGroupOverride *string `json:"securityGroupOverride,omitempty"`
	// AdditionalSecurityGroups attaches additional security groups (e.g. sg-123456).
	AdditionalSecurityGroups []string `json:"additionalSecurityGroups,omitempty"`
	// UseForInternalApi indicates whether the LB should be used by the kubelet
	UseForInternalApi bool `json:"useForInternalApi,omitempty"`
	// SSLCertificate allows you to specify the ACM cert to be used the LB
	SSLCertificate string `json:"sslCertificate,omitempty"`
	// CrossZoneLoadBalancing allows you to enable the cross zone load balancing
	CrossZoneLoadBalancing *bool `json:"crossZoneLoadBalancing,omitempty"`
}

// KubeDNSConfig defines the kube dns configuration
type KubeDNSConfig struct {
	// CacheMaxSize is the maximum entries to keep in dnsmasq
	CacheMaxSize int `json:"cacheMaxSize,omitempty"`
	// CacheMaxConcurrent is the maximum number of concurrent queries for dnsmasq
	CacheMaxConcurrent int `json:"cacheMaxConcurrent,omitempty"`
	// CoreDNSImage is used to override the default image used for CoreDNS
	CoreDNSImage string `json:"coreDNSImage,omitempty"`
	// Domain is the dns domain
	Domain string `json:"domain,omitempty"`
	// ExternalCoreFile is used to provide a complete CoreDNS CoreFile by the user - ignores other provided flags which modify the CoreFile.
	ExternalCoreFile string `json:"externalCoreFile,omitempty"`
	// Image is the name of the docker image to run - @deprecated as this is now in the addon
	Image string `json:"image,omitempty"`
	// Replicas is the number of pod replicas - @deprecated as this is now in the addon, and controlled by autoscaler
	Replicas int `json:"replicas,omitempty"`
	// Provider indicates whether CoreDNS or kube-dns will be the default service discovery.
	Provider string `json:"provider,omitempty"`
	// ServerIP is the server ip
	ServerIP string `json:"serverIP,omitempty"`
	// StubDomains redirects a domains to another DNS service
	StubDomains map[string][]string `json:"stubDomains,omitempty"`
	// UpstreamNameservers sets the upstream nameservers for queries not on the cluster domain
	UpstreamNameservers []string `json:"upstreamNameservers,omitempty"`
	// MemoryRequest specifies the memory requests of each dns container in the cluster. Default 70m.
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest specifies the cpu requests of each dns container in the cluster. Default 100m.
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// MemoryLimit specifies the memory limit of each dns container in the cluster. Default 170m.
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
}

// ExternalDNSConfig are options of the dns-controller
type ExternalDNSConfig struct {
	// Disable indicates we do not wish to run the dns-controller addon
	Disable bool `json:"disable,omitempty"`
	// WatchIngress indicates you want the dns-controller to watch and create dns entries for ingress resources
	WatchIngress *bool `json:"watchIngress,omitempty"`
	// WatchNamespace is namespace to watch, defaults to all (use to control whom can creates dns entries)
	WatchNamespace string `json:"watchNamespace,omitempty"`
}

// EtcdProviderType describes etcd cluster provisioning types (Standalone, Manager)
type EtcdProviderType string

const (
	EtcdProviderTypeManager EtcdProviderType = "Manager"
	EtcdProviderTypeLegacy  EtcdProviderType = "Legacy"
)

// EtcdClusterSpec is the etcd cluster specification
type EtcdClusterSpec struct {
	// Name is the name of the etcd cluster (main, events etc)
	Name string `json:"name,omitempty"`
	// Provider is the provider used to run etcd: standalone, manager.
	// We default to manager for kubernetes 1.11 or if the manager is configured; otherwise standalone.
	Provider EtcdProviderType `json:"provider,omitempty"`
	// Members stores the configurations for each member of the cluster (including the data volume)
	Members []*EtcdMemberSpec `json:"etcdMembers,omitempty"`
	// EnableEtcdTLS indicates the etcd service should use TLS between peers and clients
	EnableEtcdTLS bool `json:"enableEtcdTLS,omitempty"`
	// EnableTLSAuth indicates client and peer TLS auth should be enforced
	EnableTLSAuth bool `json:"enableTLSAuth,omitempty"`
	// Version is the version of etcd to run i.e. 2.1.2, 3.0.17 etcd
	Version string `json:"version,omitempty"`
	// LeaderElectionTimeout is the time (in milliseconds) for an etcd leader election timeout
	LeaderElectionTimeout *metav1.Duration `json:"leaderElectionTimeout,omitempty"`
	// HeartbeatInterval is the time (in milliseconds) for an etcd heartbeat interval
	HeartbeatInterval *metav1.Duration `json:"heartbeatInterval,omitempty"`
	// Image is the etcd docker image to use. Setting this will ignore the Version specified.
	Image string `json:"image,omitempty"`
	// Backups describes how we do backups of etcd
	Backups *EtcdBackupSpec `json:"backups,omitempty"`
	// Manager describes the manager configuration
	Manager *EtcdManagerSpec `json:"manager,omitempty"`
	// MemoryRequest specifies the memory requests of each etcd container in the cluster.
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest specifies the cpu requests of each etcd container in the cluster.
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
}

// EtcdBackupSpec describes how we want to do backups of etcd
type EtcdBackupSpec struct {
	// BackupStore is the VFS path where we will read/write backup data
	BackupStore string `json:"backupStore,omitempty"`
	// Image is the etcd backup manager image to use.  Setting this will create a sidecar container in the etcd pod with the specified image.
	Image string `json:"image,omitempty"`
}

// EtcdManagerSpec describes how we configure the etcd manager
type EtcdManagerSpec struct {
	// Image is the etcd manager image to use.
	Image string `json:"image,omitempty"`
}

// EtcdMemberSpec is a specification for a etcd member
type EtcdMemberSpec struct {
	// Name is the name of the member within the etcd cluster
	Name string `json:"name,omitempty"`
	// InstanceGroup is the instanceGroup this volume is associated
	InstanceGroup *string `json:"instanceGroup,omitempty"`
	// VolumeType is the underlying cloud storage class
	VolumeType *string `json:"volumeType,omitempty"`
	// If volume type is io1, then we need to specify the number of Iops.
	VolumeIops *int32 `json:"volumeIops,omitempty"`
	// VolumeSize is the underlying cloud volume size
	VolumeSize *int32 `json:"volumeSize,omitempty"`
	// KmsKeyId is a AWS KMS ID used to encrypt the volume
	KmsKeyId *string `json:"kmsKeyId,omitempty"`
	// EncryptedVolume indicates you want to encrypt the volume
	EncryptedVolume *bool `json:"encryptedVolume,omitempty"`
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

	// Zone is the zone the subnet is in, set for subnets that are zonally scoped
	Zone string `json:"zone,omitempty"`
	// Region is the region the subnet is in, set for subnets that are regionally scoped
	Region string `json:"region,omitempty"`

	CIDR string `json:"cidr,omitempty"`

	// ProviderID is the cloud provider id for the objects associated with the zone (the subnet on AWS)
	ProviderID string `json:"id,omitempty"`

	// Egress defines the method of traffic egress for this subnet
	Egress string `json:"egress,omitempty"`

	Type SubnetType `json:"type,omitempty"`
	// PublicIP to attach to NatGateway
	PublicIP string `json:"publicIP,omitempty"`
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

// TargetSpec allows for specifying target config in an extensible way
type TargetSpec struct {
	Terraform *TerraformSpec `json:"terraform,omitempty"`
}

func (t *TargetSpec) IsEmpty() bool {
	return t.Terraform == nil
}

// TerraformSpec allows us to specify terraform config in an extensible way
type TerraformSpec struct {
	// ProviderExtraConfig contains key/value pairs to add to the rendered terraform "provider" block
	ProviderExtraConfig *map[string]string `json:"providerExtraConfig,omitempty"`
}

func (t *TerraformSpec) IsEmpty() bool {
	return t.ProviderExtraConfig == nil
}

type GossipConfig struct {
	Protocol  *string       `json:"protocol,omitempty"`
	Listen    *string       `json:"listen,omitempty"`
	Secret    *string       `json:"secret,omitempty"`
	Secondary *GossipConfig `json:"secondary,omitempty"`
}

type DNSControllerGossipConfig struct {
	Protocol  *string                    `json:"protocol,omitempty"`
	Listen    *string                    `json:"listen,omitempty"`
	Secret    *string                    `json:"secret,omitempty"`
	Secondary *DNSControllerGossipConfig `json:"secondary,omitempty"`
	Seed      *string                    `json:"seed,omitempty"`
}
