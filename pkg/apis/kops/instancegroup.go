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

package kops

import (
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// LabelClusterName is a cluster label cloud tag
	LabelClusterName = "kops.k8s.io/cluster"
	// NodeLabelInstanceGroup is a node label set to the name of the instance group
	NodeLabelInstanceGroup = "kops.k8s.io/instancegroup"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InstanceGroup represents a group of instances with the same configuration.
type InstanceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstanceGroupSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InstanceGroupList is a list of instance groups
type InstanceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []InstanceGroup `json:"items"`
}

// InstanceGroupRole describes the roles of the nodes in this InstanceGroup.
type InstanceGroupRole string

const (
	// InstanceGroupRoleControlPlane is a control-plane role.
	InstanceGroupRoleControlPlane InstanceGroupRole = "ControlPlane"
	// InstanceGroupRoleNode is a node role.
	InstanceGroupRoleNode InstanceGroupRole = "Node"
	// InstanceGroupRoleBastion is a bastion role.
	InstanceGroupRoleBastion InstanceGroupRole = "Bastion"
	// InstanceGroupRoleAPIServer is an API server role.
	InstanceGroupRoleAPIServer InstanceGroupRole = "APIServer"
	// InstanceGroupRoleEtcd is an Etcd role.
	InstanceGroupRoleEtcd InstanceGroupRole = "Etcd"
	// InstanceGroupRoleScheduler is a Scheduler role.
	InstanceGroupRoleScheduler InstanceGroupRole = "Scheduler"
	// InstanceGroupRoleKubeControllerManager is a KubeControllerManager role.
	InstanceGroupRoleKubeControllerManager InstanceGroupRole = "KubeControllerManager"
)

// AllInstanceGroupRoles is a slice of all valid InstanceGroupRole values
var AllInstanceGroupRoles = []InstanceGroupRole{
	InstanceGroupRoleControlPlane,
	InstanceGroupRoleAPIServer,
	InstanceGroupRoleNode,
	InstanceGroupRoleBastion,
	InstanceGroupRoleEtcd,
	InstanceGroupRoleScheduler,
	InstanceGroupRoleKubeControllerManager,
}

// RoleSeparator separates the individual roles in a composite InstanceGroupRole value,
// for example "APIServer,Scheduler".
const RoleSeparator = ","

// Roles returns the set of roles this value carries.
//
// An InstanceGroupRole is a comma-separated list, so that a single InstanceGroup can take on
// several control-plane roles (for example an IG running kube-scheduler alongside the API
// server it talks to). Entries are trimmed and deduplicated. Recognised roles come first, in
// the canonical order of AllInstanceGroupRoles, so that "Scheduler,APIServer" and
// "APIServer,Scheduler" normalize identically; unrecognised entries are kept, in input order,
// after the recognised ones, so validation can report them.
func (r InstanceGroupRole) Roles() []InstanceGroupRole {
	var known, unknown []InstanceGroupRole
	seen := make(map[InstanceGroupRole]bool)

	for _, field := range strings.Split(string(r), RoleSeparator) {
		role := InstanceGroupRole(strings.TrimSpace(field))
		if role == "" || seen[role] {
			continue
		}
		seen[role] = true
		if slices.Contains(AllInstanceGroupRoles, role) {
			known = append(known, role)
		} else {
			unknown = append(unknown, role)
		}
	}

	slices.SortStableFunc(known, func(a, b InstanceGroupRole) int {
		return slices.Index(AllInstanceGroupRoles, a) - slices.Index(AllInstanceGroupRoles, b)
	})

	return append(known, unknown...)
}

// HasRole reports whether this value carries the given role.
func (r InstanceGroupRole) HasRole(role InstanceGroupRole) bool {
	// Fast path for the overwhelmingly common single-role case. It also keeps the behaviour of
	// unrecognised values byte-identical to a direct comparison.
	if !strings.Contains(string(r), RoleSeparator) {
		return r == role
	}
	return slices.Contains(r.Roles(), role)
}

// PrimaryRole returns the single most significant role this value carries, using the canonical
// order of AllInstanceGroupRoles. Use it only where exactly one role can be represented, such
// as naming a cloud resource, selecting an IAM identity, or choosing a state-store path; for
// behavioural decisions prefer the Has* helpers, which consider every role.
func (r InstanceGroupRole) PrimaryRole() InstanceGroupRole {
	roles := r.Roles()
	if len(roles) == 0 {
		return ""
	}
	return roles[0]
}

func (r InstanceGroupRole) HasControlPlane() bool {
	return r.HasRole(InstanceGroupRoleControlPlane)
}

func (r InstanceGroupRole) HasNode() bool {
	return r.HasRole(InstanceGroupRoleNode)
}

func (r InstanceGroupRole) HasBastion() bool {
	return r.HasRole(InstanceGroupRoleBastion)
}

func (r InstanceGroupRole) HasAPIServer() bool {
	return r.HasRole(InstanceGroupRoleAPIServer)
}

func (r InstanceGroupRole) HasEtcd() bool {
	return r.HasRole(InstanceGroupRoleEtcd)
}

func (r InstanceGroupRole) HasScheduler() bool {
	return r.HasRole(InstanceGroupRoleScheduler)
}

func (r InstanceGroupRole) HasKubeControllerManager() bool {
	return r.HasRole(InstanceGroupRoleKubeControllerManager)
}

func (r InstanceGroupRole) IsControlPlaneType() bool {
	return r.HasControlPlane() || r.HasAPIServer() || r.HasEtcd() || r.HasScheduler() || r.HasKubeControllerManager()
}

// WellKnownService names a cluster endpoint that an instance group can serve.
//
// This is instance group membership: which endpoints route traffic to this group. It is related
// to, but distinct from, the WellKnownService values in pkg/wellknownservices, which name the
// addresses a cluster advertises. The API server appears here twice, because a group can serve
// the endpoint clients outside the cluster use without serving the one used inside it, or the
// other way around.
type WellKnownService string

const (
	// WellKnownServiceKubeAPIServerExternal is the API server endpoint used by clients outside
	// the cluster, typically a public load balancer.
	WellKnownServiceKubeAPIServerExternal WellKnownService = "kube-apiserver-external"
	// WellKnownServiceKubeAPIServerInternal is the API server endpoint used by clients inside
	// the cluster, including the nodes themselves.
	WellKnownServiceKubeAPIServerInternal WellKnownService = "kube-apiserver-internal"
	// WellKnownServiceKopsController is the endpoint where kops-controller listens.
	WellKnownServiceKopsController WellKnownService = "kops-controller"
	// WellKnownServiceEtcdMain is the endpoint where the main etcd cluster listens.
	WellKnownServiceEtcdMain WellKnownService = "etcd-main"
)

// AllWellKnownServices is a slice of all valid WellKnownService values
var AllWellKnownServices = []WellKnownService{
	WellKnownServiceKubeAPIServerExternal,
	WellKnownServiceKubeAPIServerInternal,
	WellKnownServiceKopsController,
	WellKnownServiceEtcdMain,
}

// ClusterComponent names a cluster component that kOps schedules onto a node rather than
// running as a static pod, and which therefore has to be placed explicitly.
type ClusterComponent string

const (
	ClusterComponentCloudControllerManager ClusterComponent = "cloud-controller-manager"
	ClusterComponentKopsController         ClusterComponent = "kops-controller"
	ClusterComponentKopsChannel            ClusterComponent = "kops-channel"
	ClusterComponentCertManager            ClusterComponent = "cert-manager"
	ClusterComponentCAPIManager            ClusterComponent = "capi-manager"
)

// AllClusterComponents is a slice of all valid ClusterComponent values
var AllClusterComponents = []ClusterComponent{
	ClusterComponentCloudControllerManager,
	ClusterComponentKopsController,
	ClusterComponentKopsChannel,
	ClusterComponentCertManager,
	ClusterComponentCAPIManager,
}

const (
	// BtfsFilesystem indicates a btfs filesystem
	BtfsFilesystem = "btfs"
	// Ext4Filesystem indicates a ext3 filesystem
	Ext4Filesystem = "ext4"
	// XFSFilesystem indicates a xfs filesystem
	XFSFilesystem = "xfs"
)

// SupportedFilesystems is a list of supported filesystems to format as
var SupportedFilesystems = []string{BtfsFilesystem, Ext4Filesystem, XFSFilesystem}

type InstanceManager string

const (
	InstanceManagerCloudGroup InstanceManager = "CloudGroup"
	InstanceManagerKarpenter  InstanceManager = "Karpenter"
)

// InstanceGroupSpec is the specification for an InstanceGroup
type InstanceGroupSpec struct {
	// Manager determines what is managing the node lifecycle
	Manager InstanceManager `json:"manager,omitempty"`
	// Role determines the role of instances in this instance group.
	// This is a comma-separated list, so one instance group can take on several
	// control-plane roles, for example "APIServer,Scheduler".
	Role InstanceGroupRole `json:"role,omitempty"`
	// ServesWellKnownServices lists the cluster endpoints that should route traffic to this
	// instance group. When unset it is derived from the roles; see
	// InstanceGroup.ServedWellKnownServices.
	ServesWellKnownServices []WellKnownService `json:"servesWellKnownServices,omitempty"`
	// HostedComponents lists the cluster components that kOps schedules onto nodes, rather than
	// running as static pods, which should be placed on this instance group.
	HostedComponents []ClusterComponent `json:"hostedComponents,omitempty"`
	// Image is the instance (ami etc) we should use
	Image string `json:"image,omitempty"`
	// MinSize is the minimum size of the pool
	MinSize *int32 `json:"minSize,omitempty"`
	// MaxSize is the maximum size of the pool
	MaxSize *int32 `json:"maxSize,omitempty"`
	// Autoscale determines if autoscaling will be enabled for this instance group if cluster autoscaler is enabled
	Autoscale *bool `json:"autoscale,omitempty"`
	// AutoscalePriority determines the InstanceGroup priority for scaling when cluster autoscaler uses the priority expander.
	AutoscalePriority int16 `json:"autoscalePriority,omitempty"`
	// MachineType is the instance class
	MachineType string `json:"machineType,omitempty"`
	// RootVolume specifies options for the instances' root volumes.
	RootVolume *InstanceRootVolumeSpec `json:"rootVolume,omitempty"`
	// Volumes is a collection of additional volumes to create for instances within this instance group
	Volumes []VolumeSpec `json:"volumes,omitempty"`
	// VolumeMounts a collection of volume mounts
	VolumeMounts []VolumeMountSpec `json:"volumeMounts,omitempty"`
	// Subnets is the names of the Subnets (as specified in the Cluster) where machines in this instance group should be placed
	Subnets []string `json:"subnets,omitempty"`
	// Zones is the names of the Zones where machines in this instance group should be placed
	// This is needed for regional subnets (e.g. GCE), to restrict placement to particular zones
	Zones []string `json:"zones,omitempty"`
	// Hooks is a list of hooks for this instance group, note: these can override the cluster wide ones if required
	Hooks []HookSpec `json:"hooks,omitempty"`
	// MaxPrice indicates this is a spot-pricing group, with the specified value as our max-price bid
	MaxPrice *string `json:"maxPrice,omitempty"`
	// SpotDurationInMinutes reserves a spot block for the period specified
	SpotDurationInMinutes *int64 `json:"spotDurationInMinutes,omitempty"`
	// CPUCredits is the credit option for CPU Usage on burstable instance types (AWS only)
	CPUCredits *string `json:"cpuCredits,omitempty"`
	// AssociatePublicIP is true if we want instances to have a public IP
	AssociatePublicIP *bool `json:"associatePublicIP,omitempty"`
	// AdditionalSecurityGroups attaches additional security groups (e.g. i-123456)
	AdditionalSecurityGroups []string `json:"additionalSecurityGroups,omitempty"`
	// CloudLabels defines additional tags or labels on cloud provider resources
	CloudLabels map[string]string `json:"cloudLabels,omitempty"`
	// NodeLabels indicates the kubernetes labels for nodes in this instance group
	NodeLabels map[string]string `json:"nodeLabels,omitempty"`
	// FileAssets is a collection of file assets for this instance group
	FileAssets []FileAssetSpec `json:"fileAssets,omitempty"`
	// Describes the tenancy of this instance group. Can be either default or dedicated. Currently only applies to AWS.
	Tenancy string `json:"tenancy,omitempty"`
	// Kubelet overrides kubelet config from the ClusterSpec
	Kubelet *KubeletConfigSpec `json:"kubelet,omitempty"`
	// Taints indicates the kubernetes taints for nodes in this instance group
	Taints []string `json:"taints,omitempty"`
	// MixedInstancesPolicy defined a optional backing of an AWS ASG by a EC2 Fleet (AWS Only)
	MixedInstancesPolicy *MixedInstancesPolicySpec `json:"mixedInstancesPolicy,omitempty"`
	// CapacityRebalance makes ASGs proactively replace spot instances when the ASG receives a rebalance recommendation (AWS Only).
	CapacityRebalance *bool `json:"capacityRebalance,omitempty"`
	// AdditionalUserData is any additional user-data to be passed to the host
	AdditionalUserData []UserData `json:"additionalUserData,omitempty"`
	// SuspendProcesses disables the listed Scaling Policies
	SuspendProcesses []string `json:"suspendProcesses,omitempty"`
	// ExternalLoadBalancers define loadbalancers that should be attached to this instance group
	ExternalLoadBalancers []LoadBalancerSpec `json:"externalLoadBalancers,omitempty"`
	// DetailedInstanceMonitoring defines if detailed-monitoring is enabled (AWS only)
	DetailedInstanceMonitoring *bool `json:"detailedInstanceMonitoring,omitempty"`
	// IAMProfileSpec defines the identity of the cloud group IAM profile (AWS only).
	IAM *IAMProfileSpec `json:"iam,omitempty"`
	// SecurityGroupOverride overrides the default security group created by Kops for this IG (AWS only).
	SecurityGroupOverride *string `json:"securityGroupOverride,omitempty"`
	// InstanceProtection makes new instances in an autoscaling group protected from scale in
	InstanceProtection *bool `json:"instanceProtection,omitempty"`
	// SysctlParameters will configure kernel parameters using sysctl(8). When
	// specified, each parameter must follow the form variable=value, the way
	// it would appear in sysctl.conf.
	SysctlParameters []string `json:"sysctlParameters,omitempty"`
	// RollingUpdate defines the rolling-update behavior
	RollingUpdate *RollingUpdate `json:"rollingUpdate,omitempty"`
	// InstanceInterruptionBehavior defines if a spot instance should be terminated, hibernated,
	// or stopped after interruption
	InstanceInterruptionBehavior *string `json:"instanceInterruptionBehavior,omitempty"`
	// CompressUserData compresses parts of the user data to save space
	CompressUserData *bool `json:"compressUserData,omitempty"`
	// InstanceMetadata defines the EC2 instance metadata service options (AWS Only)
	InstanceMetadata *InstanceMetadataOptions `json:"instanceMetadata,omitempty"`
	// UpdatePolicy determines the policy for applying upgrades automatically.
	// If specified, this value overrides a value specified in the Cluster's "spec.updatePolicy" field.
	// Valid values:
	//   'automatic' (default): apply updates automatically (apply OS security upgrades, avoiding rebooting when possible)
	//   'external': do not apply updates automatically; they are applied manually or by an external system
	UpdatePolicy *string `json:"updatePolicy,omitempty"`
	// WarmPool specifies a pool of pre-warmed instances for later use (AWS only).
	WarmPool *WarmPoolSpec `json:"warmPool,omitempty"`
	// Containerd specifies override configuration for instance group
	Containerd *ContainerdConfig `json:"containerd,omitempty"`
	// Packages specifies additional packages to be installed.
	Packages []string `json:"packages,omitempty"`
	// GuestAccelerators configures additional accelerators
	GuestAccelerators []AcceleratorConfig `json:"guestAccelerators,omitempty"`
	// MaxInstanceLifetime to the maximum amount of time, in seconds, that an instance can be in service.
	// Value expected must be in form of duration ("ms", "s", "m", "h")
	MaxInstanceLifetime *metav1.Duration `json:"maxInstanceLifetime,omitempty"`
	// GCPProvisioningModel: Specifies the provisioning model of the GCP instance.
	// Valid values:
	//   'STANDARD': (default) standard provisioning with user controlled run time, no discounts
	//   'SPOT': heavily discounted, no guaranteed run time.
	GCPProvisioningModel *string `json:"gcpProvisioningModel,omitempty"`
}

const (
	// SpotAllocationStrategyLowestPrices indicates a lowest-price strategy
	SpotAllocationStrategyLowestPrices = "lowest-price"
	// SpotAllocationStrategyDiversified indicates a diversified strategy
	SpotAllocationStrategyDiversified = "diversified"
	// SpotAllocationStrategyCapacityOptimized indicates a capacity optimized strategy
	SpotAllocationStrategyCapacityOptimized = "capacity-optimized"
	// SpotAllocationStrategyCapacityOptimizedPrioritized indicates a capacity optimized prioritized strategy
	SpotAllocationStrategyCapacityOptimizedPrioritized = "capacity-optimized-prioritized"
	// SpotAllocationStrategyPriceCapacityOptimized indicates a price/capacity optimized strategy
	SpotAllocationStrategyPriceCapacityOptimized = "price-capacity-optimized"
)

// SpotAllocationStrategies is a collection of supported strategies
var SpotAllocationStrategies = []string{
	SpotAllocationStrategyLowestPrices,
	SpotAllocationStrategyDiversified,
	SpotAllocationStrategyCapacityOptimized,
	SpotAllocationStrategyCapacityOptimizedPrioritized,
	SpotAllocationStrategyPriceCapacityOptimized,
}

// InstanceRootVolumeSpec specifies options for an instance's root volume.
type InstanceRootVolumeSpec struct {
	// Size is the size of the EBS root volume to use, in GB.
	Size *int32 `json:"size,omitempty"`
	// Type is the type of the EBS root volume to use (for example gp2).
	Type *string `json:"type,omitempty"`
	// IOPS is the provisioned IOPS when the volume type is io1, io2 or gp3 (AWS only).
	IOPS *int32 `json:"iops,omitempty"`
	// Throughput is the volume throughput in MBps when the volume type is gp3 (AWS only).
	Throughput *int32 `json:"throughput,omitempty"`
	// Optimization enables EBS optimization for an instance.
	Optimization *bool `json:"optimization,omitempty"`
	// Encryption enables EBS root volume encryption for an instance.
	Encryption *bool `json:"encryption,omitempty"`
	// EncryptionKey provides the key identifier for root volume encryption.
	EncryptionKey *string `json:"encryptionKey,omitempty"`
}

// InstanceMetadataOptions defines the EC2 instance metadata service options (AWS Only)
type InstanceMetadataOptions struct {
	// HTTPPutResponseHopLimit is the desired HTTP PUT response hop limit for instance metadata requests.
	// The larger the number, the further instance metadata requests can travel. The default value is 1.
	HTTPPutResponseHopLimit *int64 `json:"httpPutResponseHopLimit,omitempty"`
	// HTTPTokens is the state of token usage for the instance metadata requests.
	// If the parameter is not specified in the request, the default state is "required".
	HTTPTokens *string `json:"httpTokens,omitempty"`
}

// MixedInstancesPolicySpec defines the specification for an autoscaling group backed by a ec2 fleet
type MixedInstancesPolicySpec struct {
	// Instances is a list of instance types which we are willing to run in the EC2 fleet
	Instances []string `json:"instances,omitempty"`
	// InstanceRequirements is a list of requirements for any instance type we are willing to run in the EC2 fleet.
	InstanceRequirements *InstanceRequirementsSpec `json:"instanceRequirements,omitempty"`
	// OnDemandAllocationStrategy indicates how to allocate instance types to fulfill On-Demand capacity
	OnDemandAllocationStrategy *string `json:"onDemandAllocationStrategy,omitempty"`
	// OnDemandBase is the minimum amount of the Auto Scaling group's capacity that must be
	// fulfilled by On-Demand Instances. This base portion is provisioned first as your group scales.
	OnDemandBase *int64 `json:"onDemandBase,omitempty"`
	// OnDemandAboveBase controls the percentages of On-Demand Instances and Spot Instances for your
	// additional capacity beyond OnDemandBase. The range is 0–100. The default value is 100. If you
	// leave this parameter set to 100, the percentages are 100% for On-Demand Instances and 0% for
	// Spot Instances.
	OnDemandAboveBase *int64 `json:"onDemandAboveBase,omitempty"`
	// SpotAllocationStrategy diversifies your Spot capacity across multiple instance types to
	// find the best pricing. Higher Spot availability may result from a larger number of
	// instance types to choose from.
	SpotAllocationStrategy *string `json:"spotAllocationStrategy,omitempty"`
	// SpotInstancePools is the number of Spot pools to use to allocate your Spot capacity (defaults to 2)
	// pools are determined from the different instance types in the Overrides array of LaunchTemplate
	SpotInstancePools *int64 `json:"spotInstancePools,omitempty"`
}

// InstanceRequirementsSpec is a list of requirements for any instance type we are willing to run in the EC2 fleet.
type InstanceRequirementsSpec struct {
	CPU    *MinMaxSpec `json:"cpu,omitempty"`
	Memory *MinMaxSpec `json:"memory,omitempty"`
	// ExcludedInstanceTypes is a list of instance types which will not be used by the instance group.
	// You can use strings with one or more wild cards, represented by an asterisk (*), to exclude an
	// instance type, size, or generation.
	ExcludedInstanceTypes []string `json:"excludedInstanceTypes,omitempty"`
}

type MinMaxSpec struct {
	Max *resource.Quantity `json:"max,omitempty"`
	Min *resource.Quantity `json:"min,omitempty"`
}

// UserData defines a user-data section
type UserData struct {
	// Name is the name of the user-data
	Name string `json:"name,omitempty"`
	// Type is the type of user-data
	Type string `json:"type,omitempty"`
	// Content is the user-data content
	Content string `json:"content,omitempty"`
}

// VolumeSpec defined the spec for an additional volume attached to the instance group
type VolumeSpec struct {
	// DeleteOnTermination configures volume retention policy upon instance termination.
	// The volume is deleted by default. Cluster deletion does not remove retained volumes.
	DeleteOnTermination *bool `json:"deleteOnTermination,omitempty"`
	// Device is an optional device name of the block device
	Device string `json:"device,omitempty"`
	// Encrypted indicates you want to encrypt the volume
	Encrypted *bool `json:"encrypted,omitempty"`
	// IOPS is the provisioned IOPS for the volume when the volume type is io1, io2 or gp3 (AWS only).
	IOPS *int64 `json:"iops,omitempty"`
	// Throughput is the volume throughput in MBps when the volume type is gp3 (AWS only).
	Throughput *int64 `json:"throughput,omitempty"`
	// Key is the encryption key identifier for the volume
	Key *string `json:"key,omitempty"`
	// Size is the size of the volume in GB
	Size int64 `json:"size,omitempty"`
	// Type is the type of volume to create and is cloud specific
	Type string `json:"type,omitempty"`
}

// VolumeMountSpec defines the specification for mounting a device
type VolumeMountSpec struct {
	// Device is the device name to provision and mount
	Device string `json:"device,omitempty"`
	// Filesystem is the filesystem to mount
	Filesystem string `json:"filesystem,omitempty"`
	// FormatOptions is a collection of options passed when formatting the device
	FormatOptions []string `json:"formatOptions,omitempty"`
	// MountOptions is a collection of mount options - @TODO need to be added
	MountOptions []string `json:"mountOptions,omitempty"`
	// Path is the location to mount the device
	Path string `json:"path,omitempty"`
}

// IAMProfileSpec is the AWS IAM Profile to attach to instances in this instance
// group. Specify the ARN for the IAM instance profile (AWS only).
type IAMProfileSpec struct {
	// Profile is the AWS IAM Profile to attach to instances in this instance group.
	// Specify the ARN for the IAM instance profile. (AWS only)
	Profile *string `json:"profile,omitempty"`
}

// IsControlPlane checks if instanceGroup is a control-plane node.
func (g *InstanceGroup) IsControlPlane() bool {
	switch {
	case g.Spec.Role.HasControlPlane():
		return true
	default:
		return false
	}
}

// IsControlPlane checks if instanceGroup is a control-plane node.
func (g *InstanceGroup) IsControlPlaneType() bool {
	switch {
	case g.Spec.Role.IsControlPlaneType():
		return true
	default:
		return false
	}
}

// IsRoleOnly reports whether the given role is the only role this instanceGroup carries.
//
// Prefer the Runs* helpers for deciding what a group should run; a group that carries
// additional roles still runs this one. IsRoleOnly is for the narrower question of whether a
// group is dedicated to a single role, for example when deciding whether it needs to reach a
// component over the network rather than on localhost.
func (g *InstanceGroup) IsRoleOnly(role InstanceGroupRole) bool {
	roles := g.Spec.Role.Roles()
	return len(roles) == 1 && roles[0] == role
}

// IsAPIServerOnly checks if the API Server is the only role of this instanceGroup
func (g *InstanceGroup) IsAPIServerOnly() bool {
	return g.IsRoleOnly(InstanceGroupRoleAPIServer)
}

// RunsAPIServer checks if instanceGroup runs an API Server
func (g *InstanceGroup) RunsAPIServer() bool {
	return g.Spec.Role.HasControlPlane() || g.Spec.Role.HasAPIServer()
}

// IsEtcdOnly checks if Etcd is the only role of this instanceGroup
func (g *InstanceGroup) IsEtcdOnly() bool {
	return g.IsRoleOnly(InstanceGroupRoleEtcd)
}

// RunsEtcd checks if instanceGroup runs Etcd
func (g *InstanceGroup) RunsEtcd() bool {
	return g.Spec.Role.HasControlPlane() || g.Spec.Role.HasEtcd()
}

// IsSchedulerOnly checks if Scheduler is the only role of this instanceGroup
func (g *InstanceGroup) IsSchedulerOnly() bool {
	return g.IsRoleOnly(InstanceGroupRoleScheduler)
}

// RunsScheduler checks if instanceGroup runs Scheduler
func (g *InstanceGroup) RunsScheduler() bool {
	return g.Spec.Role.HasControlPlane() || g.Spec.Role.HasScheduler()
}

// IsKubeControllerManagerOnly checks if KubeControllerManager is the only role of this instanceGroup
func (g *InstanceGroup) IsKubeControllerManagerOnly() bool {
	return g.IsRoleOnly(InstanceGroupRoleKubeControllerManager)
}

// RunsKubeControllerManager checks if instanceGroup runs KubeControllerManager
func (g *InstanceGroup) RunsKubeControllerManager() bool {
	return g.Spec.Role.HasControlPlane() || g.Spec.Role.HasKubeControllerManager()
}

// ServedWellKnownServices returns the cluster endpoints that should route traffic to this
// instance group.
//
// When spec.servesWellKnownServices is unset the set is derived from the group's roles, so that
// clusters predating the field keep their behaviour. Because the field is omitempty an
// explicitly empty list cannot be told apart from an unset one; that is not a limitation in
// practice, because the derived value for the one case that needs it -- an API server that
// exists only to serve the kube-scheduler or kube-controller-manager running beside it -- is
// already the empty set.
func (g *InstanceGroup) ServedWellKnownServices() []WellKnownService {
	if g.Spec.ServesWellKnownServices != nil {
		return g.Spec.ServesWellKnownServices
	}

	role := g.Spec.Role

	// An API server co-located with the scheduler or controller-manager is there for that
	// component alone, reachable on localhost, so nothing should be routed to it.
	localOnlyAPIServer := role.HasAPIServer() &&
		(role.HasScheduler() || role.HasKubeControllerManager())

	var services []WellKnownService
	if role.HasControlPlane() || (role.HasAPIServer() && !localOnlyAPIServer) {
		services = append(services,
			WellKnownServiceKubeAPIServerExternal,
			WellKnownServiceKubeAPIServerInternal,
			WellKnownServiceKopsController)
	}
	if role.HasControlPlane() || role.HasEtcd() {
		services = append(services, WellKnownServiceEtcdMain)
	}
	return services
}

// ServesWellKnownService reports whether the given cluster endpoint should route traffic to
// this instance group.
func (g *InstanceGroup) ServesWellKnownService(service WellKnownService) bool {
	return slices.Contains(g.ServedWellKnownServices(), service)
}

// ServesRemoteAPIServer reports whether anything off the instance connects to this group's API
// server. An API server that exists only for the kube-scheduler or kube-controller-manager
// running beside it is reached on localhost, so it needs neither a load balancer nor a firewall
// opening of its own.
func (g *InstanceGroup) ServesRemoteAPIServer() bool {
	return g.ServesWellKnownService(WellKnownServiceKubeAPIServerExternal) ||
		g.ServesWellKnownService(WellKnownServiceKubeAPIServerInternal)
}

// HasGVisor checks if instanceGroup is a worker that has the gVisor (runsc) runtime enabled.
// gVisor is only valid on workers; ValidateInstanceGroup rejects it on other roles.
func (g *InstanceGroup) HasGVisor() bool {
	return g.Spec.Role.HasNode() &&
		g.Spec.Containerd != nil &&
		g.Spec.Containerd.GVisor != nil &&
		g.Spec.Containerd.GVisor.Enabled != nil &&
		*g.Spec.Containerd.GVisor.Enabled
}

// IsBastion checks if instanceGroup is a bastion
func (g *InstanceGroup) IsBastion() bool {
	switch {
	case g.Spec.Role.HasBastion():
		return true
	default:
		return false
	}
}

// IsKarpenterManaged checks if instanceGroup is a worker node group managed by Karpenter.
func (g *InstanceGroup) IsKarpenterManaged() bool {
	return g.Spec.Manager == InstanceManagerKarpenter && g.Spec.Role.HasNode()
}

func (g *InstanceGroup) AddInstanceGroupNodeLabel() {
	if g.Spec.NodeLabels == nil {
		g.Spec.NodeLabels = make(map[string]string)
	}
	g.Spec.NodeLabels[NodeLabelInstanceGroup] = g.Name
}

// ToLowerString returns a lowercase form of the role, safe for use in GCE label keys and
// values, cloud resource names and state-store paths.
//
// Single-role values keep their historic spelling exactly. Composite values join the
// individual roles with an underscore, in canonical order; a hyphen would be ambiguous
// because "control-plane" already contains one.
func (r InstanceGroupRole) ToLowerString() string {
	roles := r.Roles()
	if len(roles) <= 1 {
		switch {
		case r.HasControlPlane():
			return "control-plane"
		default:
			return strings.ToLower(strings.TrimSpace(string(r)))
		}
	}

	parts := make([]string, 0, len(roles))
	for _, role := range roles {
		parts = append(parts, role.ToLowerString())
	}
	return strings.Join(parts, "_")
}

// LoadBalancer defines a load balancer
type LoadBalancerSpec struct {
	// LoadBalancerName to associate with this instance group (AWS ELB)
	LoadBalancerName *string `json:"loadBalancerName,omitempty"`
	// TargetGroupARN to associate with this instance group (AWS ALB/NLB)
	TargetGroupARN *string `json:"targetGroupARN,omitempty"`
}

// AcceleratorConfig defines an accelerator config
type AcceleratorConfig struct {
	AcceleratorCount int64  `json:"acceleratorCount,omitempty"`
	AcceleratorType  string `json:"acceleratorType,omitempty"`
}
