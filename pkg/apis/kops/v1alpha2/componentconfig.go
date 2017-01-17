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

type KubeletConfigSpec struct {
	APIServers string `json:"apiServers,omitempty" flag:"api-servers"`

	LogLevel *int32 `json:"logLevel,omitempty" flag:"v"`

	// Configuration flags - a subset of https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

	// config is the path to the config file or directory of files
	Config string `json:"config,omitempty" flag:"config"`
	// hostnameOverride is the hostname used to identify the kubelet instead
	// of the actual hostname.
	// Note: We recognize some additional values:
	//  @aws uses the hostname from the AWS metadata service
	HostnameOverride string `json:"hostnameOverride,omitempty" flag:"hostname-override"`
	// allowPrivileged enables containers to request privileged mode.
	// Defaults to false.
	AllowPrivileged *bool `json:"allowPrivileged,omitempty" flag:"allow-privileged"`
	// enableDebuggingHandlers enables server endpoints for log collection
	// and local running of containers and commands
	EnableDebuggingHandlers *bool `json:"enableDebuggingHandlers,omitempty" flag:"enable-debugging-handlers"`
	// clusterDomain is the DNS domain for this cluster. If set, kubelet will
	// configure all containers to search this domain in addition to the
	// host's search domains.
	ClusterDomain string `json:"clusterDomain,omitempty" flag:"cluster-domain"`
	// clusterDNS is the IP address for a cluster DNS server.  If set, kubelet
	// will configure all containers to use this for DNS resolution in
	// addition to the host's DNS servers
	ClusterDNS string `json:"clusterDNS,omitempty" flag:"cluster-dns"`
	// networkPluginName is the name of the network plugin to be invoked for
	// various events in kubelet/pod lifecycle
	NetworkPluginName string `json:"networkPluginName,omitempty" flag:"network-plugin"`
	// cloudProvider is the provider for cloud services.
	CloudProvider string `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	// KubeletCgroups is the absolute name of cgroups to isolate the kubelet in.
	KubeletCgroups string `json:"kubeletCgroups,omitempty" flag:"kubelet-cgroups"`
	// Cgroups that container runtime is expected to be isolated in.
	RuntimeCgroups string `json:"runtimeCgroups,omitempty" flag:"runtime-cgroups"`
	// SystemCgroups is absolute name of cgroups in which to place
	// all non-kernel processes that are not already in a container. Empty
	// for no container. Rolling back the flag requires a reboot.
	SystemCgroups string `json:"systemCgroups,omitempty" flag:"system-cgroups"`
	// cgroupRoot is the root cgroup to use for pods. This is handled by the
	// container runtime on a best effort basis.
	CgroupRoot string `json:"cgroupRoot,omitempty" flag:"cgroup-root"`
	// configureCBR0 enables the kublet to configure cbr0 based on
	// Node.Spec.PodCIDR.
	ConfigureCBR0 *bool `json:"configureCbr0,omitempty" flag:"configure-cbr0"`
	// How should the kubelet configure the container bridge for hairpin packets.
	// Setting this flag allows endpoints in a Service to loadbalance back to
	// themselves if they should try to access their own Service. Values:
	//   "promiscuous-bridge": make the container bridge promiscuous.
	//   "hairpin-veth":       set the hairpin flag on container veth interfaces.
	//   "none":               do nothing.
	// Setting --configure-cbr0 to false implies that to achieve hairpin NAT
	// one must set --hairpin-mode=veth-flag, because bridge assumes the
	// existence of a container bridge named cbr0.
	HairpinMode string `json:"hairpinMode,omitempty" flag:"hairpin-mode"`
	// The node has babysitter process monitoring docker and kubelet.
	BabysitDaemons *bool `json:"babysitDaemons,omitempty" flag:"babysit-daemons"`
	// The CIDR to use for pod IP addresses, only used in standalone mode.
	// In cluster mode, this is obtained from the master.
	PodCIDR string `json:"podCIDR,omitempty" flag:"pod-cidr"`
	// reconcileCIDR is Reconcile node CIDR with the CIDR specified by the
	// API server. No-op if register-node or configure-cbr0 is false.
	ReconcileCIDR *bool `json:"reconcileCIDR,omitempty" flag:"reconcile-cidr"`
	// registerSchedulable tells the kubelet to register the node as
	// schedulable. No-op if register-node is false.
	RegisterSchedulable *bool `json:"registerSchedulable,omitempty" flag:"register-schedulable"`
	// nodeLabels to add when registering the node in the cluster.
	NodeLabels map[string]string `json:"nodeLabels,omitempty" flag:"node-labels"`
	// nonMasqueradeCIDR configures masquerading: traffic to IPs outside this range will use IP masquerade.
	NonMasqueradeCIDR string `json:"nonMasqueradeCIDR,omitempty" flag:"non-masquerade-cidr"`

	// networkPluginMTU is the MTU to be passed to the network plugin,
	// and overrides the default MTU for cases where it cannot be automatically
	// computed (such as IPSEC).
	NetworkPluginMTU *int32 `json:"networkPluginMTU,omitEmpty" flag:"network-plugin-mtu"`
}

type KubeProxyConfig struct {
	Image string `json:"image,omitempty"`
	// TODO: Better type ?
	CPURequest string `json:"cpuRequest,omitempty"` // e.g. "20m"

	LogLevel int32 `json:"logLevel,omitempty" flag:"v"`

	// Configuration flags - a subset of https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

	// master is the address of the Kubernetes API server (overrides any value in kubeconfig)
	Master string `json:"master,omitempty" flag:"master"`
}

type KubeAPIServerConfig struct {
	PathSrvKubernetes string `json:"pathSrvKubernetes,omitempty"`
	PathSrvSshproxy   string `json:"pathSrvSshproxy,omitempty"`
	Image             string `json:"image,omitempty"`

	LogLevel int32 `json:"logLevel,omitempty" flag:"v"`

	CloudProvider         string   `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	SecurePort            int32    `json:"securePort,omitempty" flag:"secure-port"`
	Address               string   `json:"address,omitempty" flag:"address"`
	EtcdServers           []string `json:"etcdServers,omitempty" flag:"etcd-servers"`
	EtcdServersOverrides  []string `json:"etcdServersOverrides,omitempty" flag:"etcd-servers-overrides"`
	AdmissionControl      []string `json:"admissionControl,omitempty" flag:"admission-control"`
	ServiceClusterIPRange string   `json:"serviceClusterIPRange,omitempty" flag:"service-cluster-ip-range"`
	ClientCAFile          string   `json:"clientCAFile,omitempty" flag:"client-ca-file"`
	BasicAuthFile         string   `json:"basicAuthFile,omitempty" flag:"basic-auth-file"`
	TLSCertFile           string   `json:"tlsCertFile,omitempty" flag:"tls-cert-file"`
	TLSPrivateKeyFile     string   `json:"tlsPrivateKeyFile,omitempty" flag:"tls-private-key-file"`
	TokenAuthFile         string   `json:"tokenAuthFile,omitempty" flag:"token-auth-file"`
	AllowPrivileged       *bool    `json:"allowPrivileged,omitempty" flag:"allow-privileged"`
	APIServerCount        *int32   `json:"apiServerCount,omitempty" flag:"apiserver-count"`
	// keys and values in RuntimeConfig are parsed into the `--runtime-config` parameter
	// for KubeAPIServer, concatenated with commas. ex: `--runtime-config=key1=value1,key2=value2`.
	// Use this to enable alpha resources on kube-apiserver
	RuntimeConfig map[string]string `json:"runtimeConfig,omitempty" flag:"runtime-config"`

	AnonymousAuth *bool `json:"anonymousAuth,omitempty" flag:"anonymous-auth"`

	KubeletPreferredAddressTypes []string `json:"kubeletPreferredAddressTypes,omitempty" flag:"kubelet-preferred-address-types"`

	StorageBackend *string `json:"storageBackend,omitempty" flag:"storage-backend"`
}

type KubeControllerManagerConfig struct {
	Master   string `json:"master,omitempty" flag:"master"`
	LogLevel int32  `json:"logLevel,omitempty" flag:"v"`

	ServiceAccountPrivateKeyFile string `json:"serviceAccountPrivateKeyFile,omitempty" flag:"service-account-private-key-file"`

	Image string `json:"image,omitempty"`

	PathSrvKubernetes string `json:"pathSrvKubernetes,omitempty"`

	// Configuration flags - a subset of https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

	// cloudProvider is the provider for cloud services.
	CloudProvider string `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	// clusterName is the instance prefix for the cluster.
	ClusterName string `json:"clusterName,omitempty" flag:"cluster-name"`
	// clusterCIDR is CIDR Range for Pods in cluster.
	ClusterCIDR string `json:"clusterCIDR,omitempty" flag:"cluster-cidr"`
	// allocateNodeCIDRs enables CIDRs for Pods to be allocated and, if
	// ConfigureCloudRoutes is true, to be set on the cloud provider.
	AllocateNodeCIDRs *bool `json:"allocateNodeCIDRs,omitempty" flag:"allocate-node-cidrs"`
	// configureCloudRoutes enables CIDRs allocated with allocateNodeCIDRs
	// to be configured on the cloud provider.
	ConfigureCloudRoutes *bool `json:"configureCloudRoutes,omitempty" flag:"configure-cloud-routes"`
	// rootCAFile is the root certificate authority will be included in service
	// account's token secret. This must be a valid PEM-encoded CA bundle.
	RootCAFile string `json:"rootCAFile,omitempty" flag:"root-ca-file"`
	// leaderElection defines the configuration of leader election client.
	LeaderElection *LeaderElectionConfiguration `json:"leaderElection,omitempty"`
}

type KubeSchedulerConfig struct {
	Master   string `json:"master,omitempty" flag:"master"`
	LogLevel int32  `json:"logLevel,omitempty" flag:"v"`

	Image string `json:"image,omitempty"`

	// Configuration flags - a subset of https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

	// leaderElection defines the configuration of leader election client.
	LeaderElection *LeaderElectionConfiguration `json:"leaderElection,omitempty"`
}

// LeaderElectionConfiguration defines the configuration of leader election
// clients for components that can run with leader election enabled.
type LeaderElectionConfiguration struct {
	// leaderElect enables a leader election client to gain leadership
	// before executing the main loop. Enable this when running replicated
	// components for high availability.
	LeaderElect *bool `json:"leaderElect,omitempty" flag:"leader-elect"`
}
