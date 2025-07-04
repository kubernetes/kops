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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kops/pkg/apis/kops"
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
	Addons      []AddonSpec          `json:"addons,omitempty"`
	ConfigStore kops.ConfigStoreSpec `json:"-"`
	// ConfigBase is the path where we store configuration for the cluster
	// This might be different that the location when the cluster spec itself is stored,
	// both because this must be accessible to the cluster,
	// and because it might be on a different cloud or storage system (etcd vs S3)
	// +k8s:conversion-gen=false
	ConfigBase    string                 `json:"configBase,omitempty"`
	CloudProvider kops.CloudProviderSpec `json:"-"`
	// The CloudProvider to use (aws or gce)
	// +k8s:conversion-gen=false
	LegacyCloudProvider string `json:"cloudProvider,omitempty"`
	// GossipConfig for the cluster assuming the use of gossip DNS
	GossipConfig *GossipConfig `json:"gossipConfig,omitempty"`
	// ContainerRuntime was removed.
	ContainerRuntime string `json:"containerRuntime,omitempty"`
	// The version of kubernetes to install (optional, and can be a "spec" like stable)
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
	// Configuration of subnets we are targeting
	// +k8s:conversion-gen=false
	Subnets []ClusterSubnetSpec `json:"subnets,omitempty"`
	// Project is the cloud project we should use, required on GCE
	// +k8s:conversion-gen=false
	Project string `json:"project,omitempty"`
	// MasterPublicName is the external DNS name for the master nodes
	// +k8s:conversion-gen=false
	MasterPublicName string `json:"masterPublicName,omitempty"`
	// MasterInternalName is unused.
	// +k8s:conversion-gen=false
	MasterInternalName string `json:"masterInternalName,omitempty"`
	// NetworkCIDR is the CIDR used for the AWS VPC / GCE Network, or otherwise allocated to k8s
	// This is a real CIDR, not the internal k8s network
	// On AWS, it maps to the VPC CIDR.  It is not required on GCE.
	// +k8s:conversion-gen=false
	NetworkCIDR string `json:"networkCIDR,omitempty"`
	// AdditionalNetworkCIDRs is a list of additional CIDR used for the AWS VPC
	// or otherwise allocated to k8s. This is a real CIDR, not the internal k8s network
	// On AWS, it maps to any additional CIDRs added to a VPC.
	// +k8s:conversion-gen=false
	AdditionalNetworkCIDRs []string `json:"additionalNetworkCIDRs,omitempty"`
	// NetworkID is an identifier of a network, if we want to reuse/share an existing network (e.g. an AWS VPC)
	// +k8s:conversion-gen=false
	NetworkID string `json:"networkID,omitempty"`
	// Topology defines the type of network topology to use on the cluster - default public
	// This is heavily weighted towards AWS for the time being, but should also be agnostic enough
	// to port out to GCE later if needed
	// +k8s:conversion-gen=false
	Topology *TopologySpec `json:"topology,omitempty"`
	// SecretStore is the VFS path to where secrets are stored
	// +k8s:conversion-gen=false
	SecretStore string `json:"secretStore,omitempty"`
	// KeyStore is the VFS path to where SSL keys and certificates are stored
	// +k8s:conversion-gen=false
	KeyStore string `json:"keyStore,omitempty"`
	// ConfigStore is unused.
	// +k8s:conversion-gen=false
	LegacyConfigStore string `json:"configStore,omitempty"`
	// DNSZone is the DNS zone we should use when configuring DNS
	// This is because some clouds let us define a managed zone foo.bar, and then have
	// kubernetes.dev.foo.bar, without needing to define dev.foo.bar as a hosted zone.
	// DNSZone will probably be a suffix of the MasterPublicName.
	// Note that DNSZone can either by the host name of the zone (containing dots),
	// or can be an identifier for the zone.
	DNSZone string `json:"dnsZone,omitempty"`
	// DNSControllerGossipConfig for the cluster assuming the use of gossip DNS
	DNSControllerGossipConfig *DNSControllerGossipConfig `json:"dnsControllerGossipConfig,omitempty"`
	// AdditionalSANs adds additional Subject Alternate Names to apiserver cert that kops generates
	// +k8s:conversion-gen=false
	AdditionalSANs []string `json:"additionalSans,omitempty"`
	// ClusterDNSDomain is the suffix we use for internal DNS names (normally cluster.local)
	ClusterDNSDomain string `json:"clusterDNSDomain,omitempty"`
	// ServiceClusterIPRange is the CIDR, from the internal network, where we allocate IPs for services
	// +k8s:conversion-gen=false
	ServiceClusterIPRange string `json:"serviceClusterIPRange,omitempty"`
	// PodCIDR is the CIDR from which we allocate IPs for pods
	// +k8s:conversion-gen=false
	PodCIDR string `json:"podCIDR,omitempty"`
	// MasterIPRange                 string `json:",omitempty"`
	// NonMasqueradeCIDR is the CIDR for the internal k8s network (for pod IPs)
	// It cannot overlap ServiceClusterIPRange
	// +k8s:conversion-gen=false
	NonMasqueradeCIDR string `json:"nonMasqueradeCIDR,omitempty"`
	// SSHAccess determines the permitted access to SSH
	// Currently only a single CIDR is supported (though a richer grammar could be added in future)
	SSHAccess []string `json:"sshAccess,omitempty"`
	// NodePortAccess is a list of the CIDRs that can access the node ports range (30000-32767).
	NodePortAccess []string `json:"nodePortAccess,omitempty"`
	// HTTPProxy defines connection information to support use of a private cluster behind an forward HTTP Proxy
	// +k8s:conversion-gen=false
	EgressProxy *EgressProxySpec `json:"egressProxy,omitempty"`
	// SSHKeyName specifies a preexisting SSH key to use
	SSHKeyName *string `json:"sshKeyName,omitempty"`
	// KubernetesAPIAccess determines the permitted access to the API endpoints (master HTTPS)
	// Currently only a single CIDR is supported (though a richer grammar could be added in future)
	// +k8s:conversion-gen=false
	KubernetesAPIAccess []string `json:"kubernetesApiAccess,omitempty"`
	// IsolateMasters determines whether we should lock down masters so that they are not on the pod network.
	// true is the kube-up behaviour, but it is very surprising: it means that daemonsets only work on the master
	// if they have hostNetwork=true.
	// false is now the default, and it will:
	//  * give the master a normal PodCIDR
	//  * run kube-proxy on the master
	//  * enable debugging handlers on the master, so kubectl logs works
	// +k8s:conversion-gen=false
	IsolateMasters *bool `json:"isolateMasters,omitempty"`
	// UpdatePolicy determines the policy for applying upgrades automatically.
	// Valid values:
	//   'automatic' (default): apply updates automatically (apply OS security upgrades, avoiding rebooting when possible)
	//   'external': do not apply updates automatically; they are applied manually or by an external system
	UpdatePolicy *string `json:"updatePolicy,omitempty"`
	// ExternalPolicies allows the insertion of pre-existing managed policies on IG Roles
	ExternalPolicies map[string][]string `json:"externalPolicies,omitempty"`
	// Additional policies to add for roles
	AdditionalPolicies map[string]string `json:"additionalPolicies,omitempty"`
	// A collection of files assets for deployed cluster wide
	FileAssets []FileAssetSpec `json:"fileAssets,omitempty"`
	// EtcdClusters stores the configuration for each cluster
	EtcdClusters []EtcdClusterSpec `json:"etcdClusters,omitempty"`
	// Docker was removed.
	Docker *DockerConfig `json:"docker,omitempty"`
	// Component configurations
	Containerd                     *ContainerdConfig             `json:"containerd,omitempty"`
	KubeDNS                        *KubeDNSConfig                `json:"kubeDNS,omitempty"`
	KubeAPIServer                  *KubeAPIServerConfig          `json:"kubeAPIServer,omitempty"`
	KubeControllerManager          *KubeControllerManagerConfig  `json:"kubeControllerManager,omitempty"`
	ExternalCloudControllerManager *CloudControllerManagerConfig `json:"cloudControllerManager,omitempty"`
	KubeScheduler                  *KubeSchedulerConfig          `json:"kubeScheduler,omitempty"`
	KubeProxy                      *KubeProxyConfig              `json:"kubeProxy,omitempty"`
	// Kubelet is the kubelet configuration for nodes not belonging to the control plane.
	// It can be overridden by the kubelet configuration specified in the instance group.
	Kubelet *KubeletConfigSpec `json:"kubelet,omitempty"`
	// MasterKubelet is the kubelet configuration for nodes belonging to the control plane
	// It can be overridden by the kubelet configuration specified in the instance group.
	ControlPlaneKubelet *KubeletConfigSpec  `json:"masterKubelet,omitempty"`
	CloudConfig         *CloudConfiguration `json:"cloudConfig,omitempty"`
	ExternalDNS         *ExternalDNSConfig  `json:"externalDns,omitempty"`
	NTP                 *NTPConfig          `json:"ntp,omitempty"`
	// Packages specifies additional packages to be installed.
	Packages []string `json:"packages,omitempty"`

	// NodeTerminationHandler determines the cluster autoscaler configuration.
	// +k8s:conversion-gen=false
	NodeTerminationHandler *NodeTerminationHandlerSpec `json:"nodeTerminationHandler,omitempty"`
	// NodeProblemDetector determines the node problem detector configuration.
	NodeProblemDetector *NodeProblemDetectorConfig `json:"nodeProblemDetector,omitempty"`
	// MetricsServer determines the metrics server configuration.
	MetricsServer *MetricsServerConfig `json:"metricsServer,omitempty"`
	// CertManager determines the metrics server configuration.
	CertManager *CertManagerConfig `json:"certManager,omitempty"`
	// AWSLoadbalancerControllerConfig determines the AWS LB controller configuration.
	// +k8s:conversion-gen=false
	AWSLoadBalancerController *LoadBalancerControllerSpec `json:"awsLoadBalancerController,omitempty"`

	// Networking configuration
	// +k8s:conversion-gen=false
	LegacyNetworking *NetworkingSpec `json:"networking,omitempty"`
	Networking       NetworkingSpec  `json:"-"`
	// API field controls how the API is exposed outside the cluster
	// +k8s:conversion-gen=false
	LegacyAPI *APISpec `json:"api,omitempty"`
	API       APISpec  `json:"-"`
	// Authentication field controls how the cluster is configured for authentication
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
	// Authorization field controls how the cluster is configured for authorization
	Authorization *AuthorizationSpec `json:"authorization,omitempty"`
	// NodeAuthorization defined the custom node authorization configuration
	NodeAuthorization *NodeAuthorizationSpec `json:"nodeAuthorization,omitempty"`
	// CloudLabels defines additional tags or labels on cloud provider resources
	CloudLabels map[string]string `json:"cloudLabels,omitempty"`
	// Hooks for custom actions e.g. on first installation
	Hooks []HookSpec `json:"hooks,omitempty"`
	// Alternative locations for files and containers
	Assets *AssetsSpec `json:"assets,omitempty"`
	// IAM field adds control over the IAM security policies applied to resources
	IAM *IAMSpec `json:"iam,omitempty"`
	// EncryptionConfig holds the encryption config
	EncryptionConfig *bool `json:"encryptionConfig,omitempty"`
	// DisableSubnetTags controls if subnets are tagged in AWS
	// +k8s:conversion-gen=false
	TagSubnets *bool `json:"DisableSubnetTags,omitempty"`
	// Target allows for us to nest extra config for targets such as terraform
	Target *TargetSpec `json:"target,omitempty"`
	// UseHostCertificates will mount /etc/ssl/certs to inside needed containers.
	// This is needed if some APIs do have self-signed certs
	UseHostCertificates *bool `json:"useHostCertificates,omitempty"`
	// SysctlParameters will configure kernel parameters using sysctl(8). When
	// specified, each parameter must follow the form variable=value, the way
	// it would appear in sysctl.conf.
	SysctlParameters []string `json:"sysctlParameters,omitempty"`
	// RollingUpdate defines the default rolling-update settings for instance groups
	RollingUpdate *RollingUpdate `json:"rollingUpdate,omitempty"`
	// ClusterAutoscaler defines the cluster autoscaler configuration.
	ClusterAutoscaler *ClusterAutoscalerConfig `json:"clusterAutoscaler,omitempty"`
	// WarmPool defines the default warm pool settings for instance groups (AWS only).
	// +k8s:conversion-gen=false
	WarmPool *WarmPoolSpec `json:"warmPool,omitempty"`
	// ServiceAccountIssuerDiscovery configures the OIDC Issuer for ServiceAccounts.
	ServiceAccountIssuerDiscovery *ServiceAccountIssuerDiscoveryConfig `json:"serviceAccountIssuerDiscovery,omitempty"`
	// SnapshotController defines the CSI Snapshot Controller configuration.
	SnapshotController *SnapshotControllerConfig `json:"snapshotController,omitempty"`
	// Karpenter defines the Karpenter configuration.
	Karpenter *KarpenterConfig `json:"karpenter,omitempty"`
	// PodIdentityWebhook determines the EKS Pod Identity Webhook configuration.
	// +k8s:conversion-gen=false
	PodIdentityWebhook *PodIdentityWebhookSpec `json:"podIdentityWebhook,omitempty"`
}

// PodIdentityWebhookSpec configures an EKS Pod Identity Webhook.
type PodIdentityWebhookSpec struct {
	Enabled  bool `json:"enabled,omitempty"`
	Replicas int  `json:"replicas,omitempty"`
}

type KarpenterConfig struct {
	Enabled       bool               `json:"enabled,omitempty"`
	LogEncoding   string             `json:"logEncoding,omitempty"`
	LogLevel      string             `json:"logLevel,omitempty"`
	Image         string             `json:"image,omitempty"`
	MemoryLimit   *resource.Quantity `json:"memoryLimit,omitempty"`
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	CPURequest    *resource.Quantity `json:"cpuRequest,omitempty"`
}

// ServiceAccountIssuerDiscoveryConfig configures an OIDC Issuer.
type ServiceAccountIssuerDiscoveryConfig struct {
	// DiscoveryStore is the VFS path to where OIDC Issuer Discovery metadata is stored.
	DiscoveryStore string `json:"discoveryStore,omitempty"`
	// EnableAWSOIDCProvider will provision an AWS OIDC provider that trusts the ServiceAccount Issuer
	EnableAWSOIDCProvider bool `json:"enableAWSOIDCProvider,omitempty"`
	// AdditionalAudiences adds user defined audiences to the provisioned AWS OIDC provider
	AdditionalAudiences []string `json:"additionalAudiences,omitempty"`
}

// ServiceAccountExternalPermissions grants a ServiceAccount permissions to external resources.
type ServiceAccountExternalPermission struct {
	// Name is the name of the Kubernetes ServiceAccount.
	Name string `json:"name"`
	// Namespace is the namespace of the Kubernetes ServiceAccount.
	Namespace string `json:"namespace"`
	// AWS grants permissions to AWS resources.
	AWS *AWSPermission `json:"aws,omitempty"`
}

// AWSPermission grants permissions to AWS resources.
type AWSPermission struct {
	// PolicyARNs is a list of existing IAM Policies.
	PolicyARNs []string `json:"policyARNs,omitempty"`
	// InlinePolicy is an IAM Policy that will be attached inline to the IAM Role.
	InlinePolicy string `json:"inlinePolicy,omitempty"`
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
	Features []string `json:"features,omitempty"`
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
	// Mode is this file's mode and permission bits
	Mode string `json:"mode,omitempty"`
}

// AssetsSpec defined the privately hosted assets
type AssetsSpec struct {
	// ContainerRegistry is a url for to a docker registry
	ContainerRegistry *string `json:"containerRegistry,omitempty"`
	// FileRepository is the url for a private file serving repository
	FileRepository *string `json:"fileRepository,omitempty"`
	// ContainerProxy is a url for a pull-through proxy of a docker registry
	ContainerProxy *string `json:"containerProxy,omitempty"`
}

// IAMSpec adds control over the IAM security policies applied to resources
type IAMSpec struct {
	Legacy                 bool    `json:"legacy"`
	AllowContainerRegistry bool    `json:"allowContainerRegistry,omitempty"`
	PermissionsBoundary    *string `json:"permissionsBoundary,omitempty"`
	// UseServiceAccountExternalPermissions determines if managed ServiceAccounts will use external permissions directly.
	// If this is set to false, ServiceAccounts will assume external permissions from the instances they run on.
	UseServiceAccountExternalPermissions *bool `json:"useServiceAccountExternalPermissions,omitempty"`
	// ServiceAccountExternalPermissions defines the relationship between Kubernetes ServiceAccounts and permissions with external resources.
	ServiceAccountExternalPermissions []ServiceAccountExternalPermission `json:"serviceAccountExternalPermissions,omitempty"`
}

// HookSpec is a definition hook
type HookSpec struct {
	// Name is an optional name for the hook, otherwise the name is kops-hook-<index>
	Name string `json:"name,omitempty"`
	// Disabled indicates if you want the unit switched off
	Enabled *bool `json:"disabled,omitempty"`
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
	Kopeio *KopeioAuthenticationSpec    `json:"kopeio,omitempty"`
	AWS    *AWSAuthenticationSpec       `json:"aws,omitempty"`
	OIDC   *kops.OIDCAuthenticationSpec `json:"-"`
}

func (s *AuthenticationSpec) IsEmpty() bool {
	return s.Kopeio == nil && s.AWS == nil
}

type KopeioAuthenticationSpec struct{}

type AWSAuthenticationSpec struct {
	// Image is the AWS IAM Authenticator container image to use.
	Image string `json:"image,omitempty"`
	// BackendMode is the AWS IAM Authenticator backend to use. Default MountedFile
	BackendMode string `json:"backendMode,omitempty"`
	// ClusterID identifies the cluster performing authentication to prevent certain replay attacks. Default master public DNS name
	ClusterID string `json:"clusterID,omitempty"`
	// MemoryRequest memory request of AWS IAM Authenticator container. Default 20Mi
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest CPU request of AWS IAM Authenticator container. Default 10m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// MemoryLimit memory limit of AWS IAM Authenticator container. Default 20Mi
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
	// CPULimit CPU limit of AWS IAM Authenticator container. Default 10m
	CPULimit *resource.Quantity `json:"cpuLimit,omitempty"`
	// IdentityMappings maps IAM Identities to Kubernetes users/groups
	IdentityMappings []AWSAuthenticationIdentityMappingSpec `json:"identityMappings,omitempty"`
}

type AWSAuthenticationIdentityMappingSpec struct {
	// Arn of the IAM User or IAM Role to be allowed to authenticate
	ARN string `json:"arn,omitempty"`
	// Username that Kubernetes will see the user as
	Username string `json:"username,omitempty"`
	// Groups to be attached to your users/roles
	Groups []string `json:"groups,omitempty"`
}

type AuthorizationSpec struct {
	AlwaysAllow *AlwaysAllowAuthorizationSpec `json:"alwaysAllow,omitempty"`
	RBAC        *RBACAuthorizationSpec        `json:"rbac,omitempty"`
}

func (s *AuthorizationSpec) IsEmpty() bool {
	return s.RBAC == nil && s.AlwaysAllow == nil
}

type RBACAuthorizationSpec struct{}

type AlwaysAllowAuthorizationSpec struct{}

// APISpec provides configuration details related to kubeapi dns and ELB access
type APISpec struct {
	// DNS will be used to provide config on kube-apiserver ELB DNS
	DNS *DNSAccessSpec `json:"dns,omitempty"`
	// LoadBalancer is the configuration for the kube-apiserver ELB
	LoadBalancer   *LoadBalancerAccessSpec `json:"loadBalancer,omitempty"`
	PublicName     string                  `json:"-"`
	AdditionalSANs []string                `json:"-"`
	Access         []string                `json:"-"`
}

func (s *APISpec) IsEmpty() bool {
	return s.DNS == nil && s.LoadBalancer == nil
}

type DNSAccessSpec struct{}

// LoadBalancerType string describes LoadBalancer types (public, internal)
type LoadBalancerType string

const (
	LoadBalancerTypePublic   LoadBalancerType = "Public"
	LoadBalancerTypeInternal LoadBalancerType = "Internal"
)

// LoadBalancerClass string describes LoadBalancer classes (classic, network)
type LoadBalancerClass string

const (
	LoadBalancerClassClassic LoadBalancerClass = "Classic"
	LoadBalancerClassNetwork LoadBalancerClass = "Network"
)

type AccessLogSpec struct {
	// Interval is publishing interval in minutes. This parameter is only used with classic load balancer.
	Interval int `json:"interval,omitempty"`
	// Bucket is S3 bucket name to store the logs in
	Bucket *string `json:"bucket,omitempty"`
	// BucketPrefix is S3 bucket prefix. Logs are stored in the root if not configured.
	BucketPrefix *string `json:"bucketPrefix,omitempty"`
}

var SupportedLoadBalancerClasses = []string{
	string(LoadBalancerClassClassic),
	string(LoadBalancerClassNetwork),
}

// LoadBalancerSubnetSpec provides configuration for subnets used for a load balancer
type LoadBalancerSubnetSpec struct {
	// Name specifies the name of the cluster subnet
	Name string `json:"name,omitempty"`
	// PrivateIPv4Address specifies the private IPv4 address to use for a NLB
	PrivateIPv4Address *string `json:"privateIPv4Address,omitempty"`
	// AllocationID specifies the Elastic IP Allocation ID for use by a NLB
	AllocationID *string `json:"allocationId,omitempty"`
}

// LoadBalancerAccessSpec provides configuration details related to API LoadBalancer and its access
type LoadBalancerAccessSpec struct {
	// LoadBalancerClass specifies the class of load balancer to create: Classic, Network
	Class LoadBalancerClass `json:"class,omitempty"`
	// Type of load balancer to create may Public or Internal.
	Type LoadBalancerType `json:"type,omitempty"`
	// IdleTimeoutSeconds sets the timeout of the api loadbalancer.
	IdleTimeoutSeconds *int64 `json:"idleTimeoutSeconds,omitempty"`
	// SecurityGroupOverride overrides the default Kops created SG for the load balancer.
	SecurityGroupOverride *string `json:"securityGroupOverride,omitempty"`
	// AdditionalSecurityGroups attaches additional security groups (e.g. sg-123456).
	AdditionalSecurityGroups []string `json:"additionalSecurityGroups,omitempty"`
	// UseForInternalAPI indicates whether the LB should be used by the kubelet
	UseForInternalAPI bool `json:"useForInternalApi,omitempty"`
	// SSLCertificate allows you to specify the ACM cert to be used the LB
	SSLCertificate string `json:"sslCertificate,omitempty"`
	// SSLPolicy allows you to overwrite the LB listener's Security Policy
	SSLPolicy *string `json:"sslPolicy,omitempty"`
	// CrossZoneLoadBalancing allows you to enable the cross zone load balancing
	CrossZoneLoadBalancing *bool `json:"crossZoneLoadBalancing,omitempty"`
	// Subnets allows you to specify the subnets that must be used for the load balancer
	Subnets []LoadBalancerSubnetSpec `json:"subnets,omitempty"`
	// AccessLog is the configuration of access logs
	AccessLog *AccessLogSpec `json:"accessLog,omitempty"`
}

// KubeDNSConfig defines the kube dns configuration
type KubeDNSConfig struct {
	// CacheMaxSize is the maximum entries to keep in dnsmasq
	CacheMaxSize int `json:"cacheMaxSize,omitempty"`
	// CacheMaxConcurrent is the maximum number of concurrent queries for dnsmasq
	CacheMaxConcurrent int `json:"cacheMaxConcurrent,omitempty"`
	// Tolerations	are tolerations to apply to the kube-dns deployment
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Affinity is the kube-dns affinity, uses the same syntax as kubectl's affinity
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// CoreDNSImage is used to override the default image used for CoreDNS
	CoreDNSImage string `json:"coreDNSImage,omitempty"`
	// CPAImage is used to override the default image used for Cluster Proportional Autoscaler
	CPAImage string `json:"cpaImage,omitempty"`
	// Domain is the dns domain
	Domain string `json:"domain,omitempty"`
	// ExternalCoreFile is used to provide a complete CoreDNS CoreFile by the user - ignores other provided flags which modify the CoreFile.
	ExternalCoreFile string `json:"externalCoreFile,omitempty"`
	// Image is unused.
	// +k8s:conversion-gen=false
	Image string `json:"image,omitempty"`
	// Replicas is unused.
	// +k8s:conversion-gen=false
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
	// NodeLocalDNS specifies the configuration for the node-local-dns addon
	NodeLocalDNS *NodeLocalDNSConfig `json:"nodeLocalDNS,omitempty"`
	// PodAnnotations makes possible to add additional annotations to CoreDNS Pods.
	// Default: none
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`
}

// NodeLocalDNSConfig are options of the node-local-dns
type NodeLocalDNSConfig struct {
	// Enabled activates the node-local-dns addon.
	Enabled *bool `json:"enabled,omitempty"`
	// ExternalCoreFile is used to provide a complete NodeLocalDNS CoreFile by the user - ignores other provided flags which modify the CoreFile.
	ExternalCoreFile string `json:"externalCoreFile,omitempty"`
	// AdditionalConfig is used to provide additional config for node local dns by the user - it will include the original CoreFile made by kOps.
	AdditionalConfig string `json:"additionalConfig,omitempty"`
	// Image overrides the default docker image used for node-local-dns addon.
	Image *string `json:"image,omitempty"`
	// Local listen IP address. It can be any IP in the 169.254.20.0/16 space or any other IP address that can be guaranteed to not collide with any existing IP.
	LocalIP string `json:"localIP,omitempty"`
	// If enabled, nodelocal dns will use kubedns as a default upstream
	ForwardToKubeDNS *bool `json:"forwardToKubeDNS,omitempty"`
	// MemoryRequest specifies the memory requests of each node-local-dns container in the daemonset. Default 5Mi.
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest specifies the cpu requests of each node-local-dns container in the daemonset. Default 25m.
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// PodAnnotations makes possible to add additional annotations to node-local-dns.
	// Default: none
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`
}

type ExternalDNSProvider string

const (
	ExternalDNSProviderDNSController ExternalDNSProvider = "dns-controller"
	ExternalDNSProviderExternalDNS   ExternalDNSProvider = "external-dns"
)

// ExternalDNSConfig are options of the dns-controller
type ExternalDNSConfig struct {
	// Disable indicates we do not wish to run the dns-controller addon
	// +k8s:conversion-gen=false
	Disable bool `json:"disable,omitempty"`
	// WatchIngress indicates you want the dns-controller to watch and create dns entries for ingress resources.
	// Default: true if provider is 'external-dns', false otherwise.
	WatchIngress *bool `json:"watchIngress,omitempty"`
	// WatchNamespace is namespace to watch, defaults to all (use to control whom can creates dns entries)
	WatchNamespace string `json:"watchNamespace,omitempty"`
	// Provider determines which implementation of ExternalDNS to use.
	// 'dns-controller' will use kOps DNS Controller.
	// 'external-dns' will use kubernetes-sigs/external-dns.
	Provider ExternalDNSProvider `json:"provider,omitempty"`
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
	// Provider is the provider used to run etcd: Manager, Legacy.
	// Defaults to Manager.
	Provider EtcdProviderType `json:"provider,omitempty"`
	// Members stores the configurations for each member of the cluster (including the data volume)
	Members []EtcdMemberSpec `json:"etcdMembers,omitempty"`
	// EnableEtcdTLS is unused.
	// +k8s:conversion-gen=false
	EnableEtcdTLS bool `json:"enableEtcdTLS,omitempty"`
	// EnableTLSAuth is unused.
	// +k8s:conversion-gen=false
	EnableTLSAuth bool `json:"enableTLSAuth,omitempty"`
	// Version is the version of etcd to run.
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
	// Env allows users to pass in env variables to the etcd-manager container.
	// Variables starting with ETCD_ will be further passed down to the etcd process.
	// This allows etcd setting to be configured/overwriten. No config validation is done.
	// A list of etcd config ENV vars can be found at https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/configuration.md
	Env []EnvVar `json:"env,omitempty"`
	// BackupInterval which is used for backups. The default is 15 minutes.
	BackupInterval *metav1.Duration `json:"backupInterval,omitempty"`
	// BackupRetentionDays which is used for backups. The default is 90 days.
	BackupRetentionDays *uint32 `json:"backupRetentionDays,omitempty"`
	// DiscoveryPollInterval which is used for discovering other cluster members. The default is 60 seconds.
	DiscoveryPollInterval *metav1.Duration `json:"discoveryPollInterval,omitempty"`
	// ListenMetricsURLs is the list of URLs to listen on that will respond to both the /metrics and /health endpoints
	ListenMetricsURLs []string `json:"listenMetricsURLs,omitempty"`
	// LogLevel allows the klog library verbose log level to be set for etcd-manager. The default is 6.
	// https://github.com/google/glog#verbose-logging
	LogLevel *int32 `json:"logLevel,omitempty"`
}

// EtcdMemberSpec is a specification for a etcd member
type EtcdMemberSpec struct {
	// Name is the name of the member within the etcd cluster
	Name string `json:"name,omitempty"`
	// InstanceGroup is the instanceGroup this volume is associated
	InstanceGroup *string `json:"instanceGroup,omitempty"`
	// VolumeType is the underlying cloud storage class
	VolumeType *string `json:"volumeType,omitempty"`
	// If volume type is io1, then we need to specify the number of IOPS.
	VolumeIOPS *int32 `json:"volumeIops,omitempty"`
	// Parameter for disks that support provisioned throughput
	VolumeThroughput *int32 `json:"volumeThroughput,omitempty"`
	// VolumeSize is the underlying cloud volume size
	VolumeSize *int32 `json:"volumeSize,omitempty"`
	// KmsKeyID is a AWS KMS ID used to encrypt the volume
	KmsKeyID *string `json:"kmsKeyId,omitempty"`
	// EncryptedVolume indicates you want to encrypt the volume
	EncryptedVolume *bool `json:"encryptedVolume,omitempty"`
}

// SubnetType string describes subnet types (public, private, utility)
type SubnetType string

const (
	SubnetTypePublic    SubnetType = "Public"
	SubnetTypePrivate   SubnetType = "Private"
	SubnetTypeDualStack SubnetType = "DualStack"
	SubnetTypeUtility   SubnetType = "Utility"
)

type ClusterSubnetSpec struct {
	Name string `json:"name,omitempty"`

	// Zone is the zone the subnet is in, set for subnets that are zonally scoped
	Zone string `json:"zone,omitempty"`
	// Region is the region the subnet is in, set for subnets that are regionally scoped
	Region string `json:"region,omitempty"`

	// CIDR is the IPv4 CIDR block assigned to the subnet.
	CIDR string `json:"cidr,omitempty"`
	// IPv6CIDR is the IPv6 CIDR block assigned to the subnet.
	IPv6CIDR string `json:"ipv6CIDR,omitempty"`

	// ID is the cloud provider ID for the objects associated with the zone (the subnet on AWS).
	ID string `json:"id,omitempty"`

	// Egress defines the method of traffic egress for this subnet
	Egress string `json:"egress,omitempty"`

	Type SubnetType `json:"type,omitempty"`
	// PublicIP to attach to NatGateway
	PublicIP string `json:"publicIP,omitempty"`

	// AdditionalRoutes to attach to the subnet's route table
	AdditionalRoutes []RouteSpec `json:"additionalRoutes,omitempty"`
}

type RouteSpec struct {
	// CIDR destination of the route
	CIDR string `json:"cidr,omitempty"`
	// Target of the route
	Target string `json:"target,omitempty"`
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
	// ProviderExtraConfig contains key/value pairs to add to the main terraform provider block
	ProviderExtraConfig map[string]string `json:"providerExtraConfig,omitempty"`
	// FilesProviderExtraConfig contains key/value pairs to add to the terraform provider block used for managed files
	FilesProviderExtraConfig map[string]string `json:"filesProviderExtraConfig,omitempty"`
}

func (t *TerraformSpec) IsEmpty() bool {
	return len(t.ProviderExtraConfig) == 0 && len(t.FilesProviderExtraConfig) == 0
}

// EnvVar represents an environment variable present in a Container.
type EnvVar struct {
	// Name of the environment variable. Must be a C_IDENTIFIER.
	Name string `json:"name"`

	// Variable references $(VAR_NAME) are expanded
	// using the previous defined environment variables in the container and
	// any service environment variables. If a variable cannot be resolved,
	// the reference in the input string will be unchanged. The $(VAR_NAME)
	// syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped
	// references will never be expanded, regardless of whether the variable
	// exists or not.
	// Defaults to "".
	// +optional
	Value string `json:"value,omitempty"`
}

type GossipConfig struct {
	Protocol  *string                `json:"protocol,omitempty"`
	Listen    *string                `json:"listen,omitempty"`
	Secret    *string                `json:"secret,omitempty"`
	Secondary *GossipConfigSecondary `json:"secondary,omitempty"`
}

type GossipConfigSecondary struct {
	Protocol *string `json:"protocol,omitempty"`
	Listen   *string `json:"listen,omitempty"`
	Secret   *string `json:"secret,omitempty"`
}

type DNSControllerGossipConfig struct {
	Protocol  *string                             `json:"protocol,omitempty"`
	Listen    *string                             `json:"listen,omitempty"`
	Secret    *string                             `json:"secret,omitempty"`
	Secondary *DNSControllerGossipConfigSecondary `json:"secondary,omitempty"`
	Seed      *string                             `json:"seed,omitempty"`
}

type DNSControllerGossipConfigSecondary struct {
	Protocol *string `json:"protocol,omitempty"`
	Listen   *string `json:"listen,omitempty"`
	Secret   *string `json:"secret,omitempty"`
	Seed     *string `json:"seed,omitempty"`
}

type RollingUpdate struct {
	// DrainAndTerminate enables draining and terminating nodes during rolling updates.
	// Defaults to true.
	DrainAndTerminate *bool `json:"drainAndTerminate,omitempty"`
	// MaxUnavailable is the maximum number of nodes that can be unavailable during the update.
	// The value can be an absolute number (for example 5) or a percentage of desired
	// nodes (for example 10%).
	// The absolute number is calculated from a percentage by rounding down.
	// Defaults to 1 if MaxSurge is 0, otherwise defaults to 0.
	// Example: when this is set to 30%, the InstanceGroup can be scaled
	// down to 70% of desired nodes immediately when the rolling update
	// starts. Once new nodes are ready, more old nodes can be drained,
	// ensuring that the total number of nodes available at all times
	// during the update is at least 70% of desired nodes.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
	// MaxSurge is the maximum number of extra nodes that can be created
	// during the update.
	// The value can be an absolute number (for example 5) or a percentage of
	// desired machines (for example 10%).
	// The absolute number is calculated from a percentage by rounding up.
	// Has no effect on instance groups with role "Master".
	// Defaults to 1 on AWS, 0 otherwise.
	// Example: when this is set to 30%, the InstanceGroup can be scaled
	// up immediately when the rolling update starts, such that the total
	// number of old and new nodes do not exceed 130% of desired
	// nodes.
	// +optional
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`
}

type PackagesConfig struct {
	// HashAmd64 overrides the hash for the AMD64 package.
	HashAmd64 *string `json:"hashAmd64,omitempty"`
	// HashArm64 overrides the hash for the ARM64 package.
	HashArm64 *string `json:"hashArm64,omitempty"`
	// UrlAmd64 overrides the URL for the AMD64 package.
	UrlAmd64 *string `json:"urlAmd64,omitempty"`
	// UrlArm64 overrides the URL for the ARM64 package.
	UrlArm64 *string `json:"urlArm64,omitempty"`
}

type WarmPoolSpec struct {
	// MinSize is the minimum size of the pool
	MinSize int64 `json:"minSize,omitempty"`
	// MaxSize is the maximum size of the warm pool. The desired size of the instance group
	// is subtracted from this number to determine the desired size of the warm pool
	// (unless the resulting number is smaller than MinSize).
	// The default is the instance group's MaxSize.
	MaxSize *int64 `json:"maxSize,omitempty"`
	// EnableLifecycleHook determines if an ASG lifecycle hook will be added ensuring that nodeup runs to completion.
	// Note that the metadata API must be protected from arbitrary Pods when this is enabled.
	EnableLifecycleHook bool `json:"enableLifecycleHook,omitempty"`
}
