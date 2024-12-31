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

package nodeup

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/reflectutils"
)

// Config is the configuration for the nodeup binary
type Config struct {
	// Assets are locations where we can find files to be installed
	// TODO: Remove once everything is in containers?
	Assets map[architectures.Architecture][]string `json:",omitempty"`
	// Images are a list of images we should preload
	Images map[architectures.Architecture][]*Image `json:"images,omitempty"`
	// ClusterName is the name of the cluster
	ClusterName string `json:",omitempty"`
	// Channels is a list of channels that we should apply
	Channels []string `json:"channels,omitempty"`
	// ApiserverAdditionalIPs are additional IP address to put in the apiserver server cert.
	ApiserverAdditionalIPs []string `json:",omitempty"`
	// KubernetesVersion is the version of Kubernetes to install.
	KubernetesVersion string
	// Packages specifies additional packages to be installed.
	Packages []string `json:"packages,omitempty"`

	// ConfigStore configures the stores that nodes use to get their configuration when they don't use kops-controller.
	ConfigStore *kops.ConfigStoreSpec `json:"configStore,omitempty"`

	// EtcdClusterNames are the names of the etcd clusters.
	EtcdClusterNames []string `json:",omitempty"`
	// EtcdManifests are the manifests for running etcd.
	EtcdManifests []string `json:"etcdManifests,omitempty"`

	// CAs are the CA certificates to trust.
	CAs map[string]string
	// KeypairIDs are the IDs of keysets used to sign things.
	KeypairIDs map[string]string
	// DefaultMachineType is the first-listed instance machine type, used if querying instance metadata fails.
	DefaultMachineType *string `json:",omitempty"`
	// EnableLifecycleHook defines whether we need to complete a lifecycle hook.
	EnableLifecycleHook bool `json:",omitempty"`
	// StaticManifests describes generic static manifests
	// Using this allows us to keep complex logic out of nodeup
	StaticManifests []*StaticManifest `json:"staticManifests,omitempty"`
	// KubeletConfig defines the kubelet configuration.
	KubeletConfig kops.KubeletConfigSpec
	// KubeProxy defines the kube-proxy configuration.
	KubeProxy *kops.KubeProxyConfig
	// Networking configures networking.
	Networking kops.NetworkingSpec
	// UseCiliumEtcd is true when a Cilium etcd cluster is present.
	UseCiliumEtcd bool `json:",omitempty"`
	// UsesKubenet specifies that the CNI is derived from Kubenet.
	UsesKubenet bool `json:",omitempty"`
	// NTPUnmanaged is true when NTP is not managed by kOps.
	NTPUnmanaged bool `json:",omitempty"`
	// ServiceNodePortRange is the service NodePort range.
	ServiceNodePortRange string `json:",omitempty"`
	// SysctlParameters will configure kernel parameters using sysctl(8).
	SysctlParameters []string `json:",omitempty"`
	// UpdatePolicy determines the policy for applying upgrades automatically.
	UpdatePolicy string
	// VolumeMounts are a collection of volume mounts.
	VolumeMounts []kops.VolumeMountSpec `json:",omitempty"`

	// FileAssets are a collection of file assets for this instance group.
	FileAssets []kops.FileAssetSpec `json:",omitempty"`
	// Hooks are for custom actions, for example on first installation.
	Hooks [][]kops.HookSpec
	// ContainerdConfig holds the configuration for containerd.
	ContainerdConfig *kops.ContainerdConfig `json:"containerdConfig,omitempty"`

	// APIServerConfig is additional configuration for nodes running an APIServer.
	APIServerConfig *APIServerConfig `json:",omitempty"`
	// ControlPlaneConfig is additional configuration for control-plane nodes.
	ControlPlaneConfig *ControlPlaneConfig `json:",omitempty"`
	// GossipConfig is configuration for gossip DNS.
	GossipConfig *kops.GossipConfig `json:",omitempty"`
	// DNSZone is the DNS zone we should use when configuring DNS.
	DNSZone string `json:",omitempty"`
	// NvidiaGPU contains the configuration for nvidia
	NvidiaGPU *kops.NvidiaGPUConfig `json:",omitempty"`

	// AWS-specific
	// DisableSecurityGroupIngress disables the Cloud Controller Manager's creation
	// of an AWS Security Group for each load balancer provisioned for a Service.
	DisableSecurityGroupIngress *bool `json:"disableSecurityGroupIngress,omitempty"`
	// ElbSecurityGroup specifies an existing AWS Security group for the Cloud Controller
	// Manager to assign to each ELB provisioned for a Service, instead of creating
	// one per ELB.
	ElbSecurityGroup *string `json:"elbSecurityGroup,omitempty"`
	// NodeIPFamilies controls the IP families reported for each node.
	NodeIPFamilies []string `json:"nodeIPFamilies,omitempty"`
	// WarmPoolImages are the container images to pre-pull during instance pre-initialization
	WarmPoolImages []string `json:"warmPoolImages,omitempty"`

	// Azure-specific
	// AzureLocation is the location of the resource group that the cluster is deployed in.
	AzureLocation string `json:",omitempty"`
	// AzureSubscriptionID is the ID of the Azure Subscription that the cluster is deployed in.
	AzureSubscriptionID string `json:",omitempty"`
	// AzureTenantID is the ID of the tenant that the cluster is deployed in.
	AzureTenantID string `json:",omitempty"`
	// AzureResourceGroup is the name of the resource group that the cluster is deployed in.
	AzureResourceGroup string `json:",omitempty"`
	// AzureRouteTableName is the name of the route table attached to the subnet that the cluster is deployed in.
	AzureRouteTableName string `json:",omitempty"`

	// GCE-specific
	Multizone          *bool   `json:"multizone,omitempty"`
	NodeTags           *string `json:"nodeTags,omitempty"`
	NodeInstancePrefix *string `json:"nodeInstancePrefix,omitempty"`

	// Openstack-specific
	Openstack *kops.OpenstackSpec `json:",omitempty"`

	// Discovery methods
	UsesLegacyGossip bool `json:"usesLegacyGossip"`
	UsesNoneDNS      bool `json:"usesNoneDNS"`
}

// BootConfig is the configuration for the nodeup binary that might be too big to fit in userdata.
type BootConfig struct {
	// CloudProvider is the cloud provider in use.
	CloudProvider kops.CloudProviderID
	// ConfigBase is the base VFS path for config objects.
	ConfigBase *string `json:",omitempty"`
	// ConfigServer holds the configuration for the configuration server.
	ConfigServer *ConfigServerOptions `json:",omitempty"`
	// APIServerIPs is the API server IP addresses.
	// This field is used for adding an alias for api.internal. in /etc/hosts, when Topology.DNS.Type == DNSTypeNone.
	APIServerIPs []string `json:",omitempty"`
	// ClusterName is the name of the cluster.
	ClusterName string `json:",omitempty"`
	// InstanceGroupName is the name of the instance group.
	InstanceGroupName string `json:",omitempty"`
	// InstanceGroupRole is the instance group role.
	InstanceGroupRole kops.InstanceGroupRole
	// NodeupConfigHash holds a secure hash of the nodeup.Config.
	NodeupConfigHash string
}

type ConfigServerOptions struct {
	// Servers are the addresses of the configuration servers to use (kops-controller)
	Servers []string `json:"servers,omitempty"`
	// CACertificates are the certificates to trust for fi.CertificateIDCA.
	CACertificates string
}

// Image is a container image we should pre-load
type Image struct {
	// This is the name we would pass to "docker run", whereas source could be a URL from which we would download an image.
	Name string `json:"name,omitempty"`
	// Sources is a list of URLs from which we should download the image
	Sources []string `json:"sources,omitempty"`
	// Hash is the hash of the file, to verify image integrity (even over http)
	Hash string `json:"hash,omitempty"`
}

// StaticManifest is a generic static manifest
type StaticManifest struct {
	// Key identifies the static manifest
	Key string `json:"key,omitempty"`
	// Path is the path to the manifest
	Path string `json:"path,omitempty"`
}

// APIServerConfig is additional configuration for nodes running an APIServer.
type APIServerConfig struct {
	// ClusterDNSDomain is the suffix we use for internal DNS names (normally cluster.local).
	ClusterDNSDomain string
	// KubeAPIServer is a copy of the KubeAPIServerConfig from the cluster spec.
	KubeAPIServer *kops.KubeAPIServerConfig
	// API controls how the Kubernetes API is exposed.
	API kops.APISpec
	// Authentication is a copy of the AuthenticationSpec from the cluster spec.
	Authentication *kops.AuthenticationSpec `json:",omitempty"`
	// EncryptionConfigSecretHash is a hash of the encryptionconfig secret.
	// It is empty if EncryptionConfig is not enabled.
	// TODO: give secrets IDs and look them up like we do keypairs.
	EncryptionConfigSecretHash string `json:",omitempty"`
	// ServiceAccountPublicKeys are the service-account public keys to trust.
	ServiceAccountPublicKeys string
}

// ControlPlaneConfig is additional configuration for control-plane nodes.
type ControlPlaneConfig struct {
	// KubeControllerManager is the configuration for the kube-controller-manager.
	KubeControllerManager kops.KubeControllerManagerConfig
	// KubeScheduler is the configuration for the kube-scheduler.
	KubeScheduler kops.KubeSchedulerConfig
}

func NewConfig(cluster *kops.Cluster, instanceGroup *kops.InstanceGroup) (*Config, *BootConfig) {
	role := instanceGroup.Spec.Role

	clusterHooks := filterHooks(cluster.Spec.Hooks, instanceGroup.Spec.Role)
	igHooks := filterHooks(instanceGroup.Spec.Hooks, instanceGroup.Spec.Role)

	config := Config{
		ClusterName:       cluster.ObjectMeta.Name,
		KubernetesVersion: cluster.Spec.KubernetesVersion,
		CAs:               map[string]string{},
		KeypairIDs:        map[string]string{},
		Networking: kops.NetworkingSpec{
			NonMasqueradeCIDR:     cluster.Spec.Networking.NonMasqueradeCIDR,
			ServiceClusterIPRange: cluster.Spec.Networking.ServiceClusterIPRange,
		},
		UsesKubenet:          cluster.Spec.Networking.UsesKubenet(),
		ServiceNodePortRange: cluster.Spec.KubeAPIServer.ServiceNodePortRange,
		VolumeMounts:         instanceGroup.Spec.VolumeMounts,
		FileAssets:           append(filterFileAssets(instanceGroup.Spec.FileAssets, role), filterFileAssets(cluster.Spec.FileAssets, role)...),
		Hooks:                [][]kops.HookSpec{igHooks, clusterHooks},
		UsesLegacyGossip:     cluster.UsesLegacyGossip(),
		UsesNoneDNS:          cluster.UsesNoneDNS(),
	}

	bootConfig := BootConfig{
		CloudProvider:     cluster.GetCloudProvider(),
		ClusterName:       cluster.ObjectMeta.Name,
		InstanceGroupName: instanceGroup.ObjectMeta.Name,
		InstanceGroupRole: role,
	}

	if cluster.Spec.Containerd != nil || instanceGroup.Spec.Containerd != nil {
		config.ContainerdConfig = buildContainerdConfig(cluster, instanceGroup)
	}

	if (cluster.Spec.Containerd != nil && cluster.Spec.Containerd.NvidiaGPU != nil) || (instanceGroup.Spec.Containerd != nil && instanceGroup.Spec.Containerd.NvidiaGPU != nil) {
		config.NvidiaGPU = buildNvidiaConfig(cluster, instanceGroup)
	}

	config.KubeProxy = buildKubeProxy(cluster, instanceGroup)

	if cluster.Spec.NTP != nil && cluster.Spec.NTP.Managed != nil && !*cluster.Spec.NTP.Managed {
		config.NTPUnmanaged = true
	}

	if cluster.Spec.CloudProvider.AWS != nil {
		aws := cluster.Spec.CloudProvider.AWS
		warmPool := aws.WarmPool.ResolveDefaults(instanceGroup)
		if warmPool.IsEnabled() && warmPool.EnableLifecycleHook {
			config.EnableLifecycleHook = true
		}

		if instanceGroup.HasAPIServer() {
			config.DisableSecurityGroupIngress = aws.DisableSecurityGroupIngress
			config.ElbSecurityGroup = aws.ElbSecurityGroup
			config.NodeIPFamilies = aws.NodeIPFamilies
		}
	}

	if cluster.Spec.CloudProvider.Azure != nil {
		config.AzureLocation = cluster.Spec.Networking.Subnets[0].Region
		config.AzureSubscriptionID = cluster.Spec.CloudProvider.Azure.SubscriptionID
		config.AzureTenantID = cluster.Spec.CloudProvider.Azure.TenantID
		config.AzureResourceGroup = cluster.AzureResourceGroupName()
		config.AzureRouteTableName = cluster.AzureRouteTableName()
		config.Networking.NetworkID = cluster.Spec.Networking.NetworkID
	}

	if cluster.Spec.CloudProvider.GCE != nil {
		gce := cluster.Spec.CloudProvider.GCE
		config.Multizone = gce.Multizone
		config.NodeTags = gce.NodeTags
		config.NodeInstancePrefix = gce.NodeInstancePrefix
	}

	config.Openstack = cluster.Spec.CloudProvider.Openstack

	if instanceGroup.Spec.UpdatePolicy != nil {
		config.UpdatePolicy = *instanceGroup.Spec.UpdatePolicy
	} else if cluster.Spec.UpdatePolicy != nil {
		config.UpdatePolicy = *cluster.Spec.UpdatePolicy
	} else {
		config.UpdatePolicy = kops.UpdatePolicyAutomatic
	}

	if cluster.Spec.Networking.AmazonVPC != nil {
		config.Networking.AmazonVPC = &kops.AmazonVPCNetworkingSpec{}
		config.DefaultMachineType = aws.String(strings.Split(instanceGroup.Spec.MachineType, ",")[0])
	}

	if cluster.Spec.Networking.Calico != nil {
		config.Networking.Calico = &kops.CalicoNetworkingSpec{}
	}

	if cluster.Spec.Networking.Cilium != nil {
		config.Networking.Cilium = &kops.CiliumNetworkingSpec{}
		if cluster.Spec.Networking.Cilium.IPAM == kops.CiliumIpamEni {
			config.Networking.Cilium.IPAM = kops.CiliumIpamEni
		}
		if model.UseCiliumEtcd(cluster) {
			config.UseCiliumEtcd = true
		}
	}

	if cluster.Spec.Networking.CNI != nil && cluster.Spec.Networking.CNI.UsesSecondaryIP {
		config.Networking.CNI = &kops.CNINetworkingSpec{UsesSecondaryIP: true}
	}

	if cluster.Spec.Networking.Flannel != nil {
		config.Networking.Flannel = &kops.FlannelNetworkingSpec{}
	}

	if cluster.Spec.Networking.KubeRouter != nil {
		config.Networking.KubeRouter = &kops.KuberouterNetworkingSpec{}
	}

	if cluster.Spec.Networking.Kindnet != nil {
		config.Networking.Kindnet = &kops.KindnetNetworkingSpec{}
	}

	if instanceGroup.Spec.Kubelet != nil {
		config.KubeletConfig = *instanceGroup.Spec.Kubelet
	}

	if instanceGroup.HasAPIServer() {
		config.APIServerConfig = &APIServerConfig{
			ClusterDNSDomain: cluster.Spec.ClusterDNSDomain,
			KubeAPIServer:    cluster.Spec.KubeAPIServer,
			API: kops.APISpec{
				PublicName:     cluster.Spec.API.PublicName,
				AdditionalSANs: cluster.Spec.API.AdditionalSANs,
			},
		}
		if cluster.Spec.Authentication != nil {
			config.APIServerConfig.Authentication = cluster.Spec.Authentication
			if cluster.Spec.Authentication.AWS != nil {
				// The values go into the manifest and aren't needed by nodeup.
				config.APIServerConfig.Authentication.AWS = &kops.AWSAuthenticationSpec{}
			}
		}
		if cluster.Spec.API.DNS != nil {
			config.APIServerConfig.API.DNS = &kops.DNSAccessSpec{}
		}
		if cluster.Spec.API.LoadBalancer != nil && cluster.Spec.API.LoadBalancer.UseForInternalAPI {
			config.APIServerConfig.API.LoadBalancer = &kops.LoadBalancerAccessSpec{UseForInternalAPI: true}
		}
	}

	if instanceGroup.HasAPIServer() || !model.UseKopsControllerForNodeConfig(cluster) {
		config.ConfigStore = &kops.ConfigStoreSpec{
			Keypairs: cluster.Spec.ConfigStore.Keypairs,
			Secrets:  cluster.Spec.ConfigStore.Secrets,
		}
	}

	if instanceGroup.HasAPIServer() || cluster.UsesLegacyGossip() {
		config.Networking.EgressProxy = cluster.Spec.Networking.EgressProxy
	}

	if instanceGroup.IsControlPlane() || cluster.UsesLegacyGossip() {
		config.DNSZone = cluster.Spec.DNSZone
	}

	if cluster.UsesLegacyGossip() {
		config.GossipConfig = cluster.Spec.GossipConfig
	}

	if instanceGroup.IsControlPlane() {
		config.ControlPlaneConfig = &ControlPlaneConfig{
			KubeControllerManager: *cluster.Spec.KubeControllerManager,
			KubeScheduler:         *cluster.Spec.KubeScheduler,
		}
	}

	if len(instanceGroup.Spec.SysctlParameters) > 0 {
		config.SysctlParameters = append(config.SysctlParameters,
			"# Custom sysctl parameters from instance group spec",
			"")
		config.SysctlParameters = append(config.SysctlParameters, instanceGroup.Spec.SysctlParameters...)
	}

	if len(cluster.Spec.SysctlParameters) > 0 {
		config.SysctlParameters = append(config.SysctlParameters,
			"# Custom sysctl parameters from cluster spec",
			"")
		config.SysctlParameters = append(config.SysctlParameters, cluster.Spec.SysctlParameters...)
	}

	return &config, &bootConfig
}

// buildContainerdConfig builds containerd configuration for instance. Instance group configuration will override cluster configuration
func buildContainerdConfig(cluster *kops.Cluster, instanceGroup *kops.InstanceGroup) *kops.ContainerdConfig {
	config := cluster.Spec.Containerd.DeepCopy()
	if instanceGroup.Spec.Containerd != nil {
		reflectutils.JSONMergeStruct(&config, instanceGroup.Spec.Containerd)
	}
	return config
}

// buildNvidiaConfig builds nvidia configuration for instance group
func buildNvidiaConfig(cluster *kops.Cluster, instanceGroup *kops.InstanceGroup) *kops.NvidiaGPUConfig {
	config := &kops.NvidiaGPUConfig{}
	if cluster.Spec.Containerd != nil && cluster.Spec.Containerd.NvidiaGPU != nil {
		config = cluster.Spec.Containerd.NvidiaGPU
	}

	if instanceGroup.Spec.Containerd != nil && instanceGroup.Spec.Containerd.NvidiaGPU != nil {
		reflectutils.JSONMergeStruct(&config, instanceGroup.Spec.Containerd.NvidiaGPU)
	}

	if config.DriverPackage == "" {
		config.DriverPackage = kops.NvidiaDefaultDriverPackage
	}

	return config
}

// buildkubeProxy builds the kube-proxy configuration for an instance group.
func buildKubeProxy(cluster *kops.Cluster, instanceGroup *kops.InstanceGroup) *kops.KubeProxyConfig {
	config := &kops.KubeProxyConfig{}
	if cluster.Spec.KubeProxy != nil {
		config = cluster.Spec.KubeProxy
	}
	if config.Enabled != nil && !*config.Enabled {
		return nil
	}
	if instanceGroup.IsControlPlane() && cluster.Spec.Networking.IsolateControlPlane != nil && *cluster.Spec.Networking.IsolateControlPlane {
		return nil
	}

	return config
}

func filterFileAssets(f []kops.FileAssetSpec, role kops.InstanceGroupRole) []kops.FileAssetSpec {
	var fileAssets []kops.FileAssetSpec
	for _, fileAsset := range f {
		if len(fileAsset.Roles) > 0 && !containsRole(role, fileAsset.Roles) {
			continue
		}
		fileAsset.Roles = nil
		fileAssets = append(fileAssets, fileAsset)
	}
	return fileAssets
}

func filterHooks(h []kops.HookSpec, role kops.InstanceGroupRole) []kops.HookSpec {
	var hooks []kops.HookSpec
	for _, hook := range h {
		if len(hook.Roles) > 0 && !containsRole(role, hook.Roles) {
			continue
		}
		hook.Roles = nil
		hooks = append(hooks, hook)
	}
	return hooks
}

func containsRole(v kops.InstanceGroupRole, list []kops.InstanceGroupRole) bool {
	for _, x := range list {
		if v == x {
			return true
		}
	}

	return false
}
