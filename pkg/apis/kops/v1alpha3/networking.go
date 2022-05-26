/*
Copyright 2021 The Kubernetes Authors.

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

package v1alpha3

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/kops/pkg/apis/kops"
)

// NetworkingSpec allows selection and configuration of a networking plugin
type NetworkingSpec struct {
	Classic    *kops.ClassicNetworkingSpec `json:"-"`
	Kubenet    *KubenetNetworkingSpec      `json:"kubenet,omitempty"`
	External   *ExternalNetworkingSpec     `json:"external,omitempty"`
	CNI        *CNINetworkingSpec          `json:"cni,omitempty"`
	Kopeio     *KopeioNetworkingSpec       `json:"kopeio,omitempty"`
	Weave      *WeaveNetworkingSpec        `json:"weave,omitempty"`
	Flannel    *FlannelNetworkingSpec      `json:"flannel,omitempty"`
	Calico     *CalicoNetworkingSpec       `json:"calico,omitempty"`
	Canal      *CanalNetworkingSpec        `json:"canal,omitempty"`
	Kuberouter *KuberouterNetworkingSpec   `json:"kuberouter,omitempty"`
	Romana     *kops.RomanaNetworkingSpec  `json:"-"`
	AmazonVPC  *AmazonVPCNetworkingSpec    `json:"amazonvpc,omitempty"`
	Cilium     *CiliumNetworkingSpec       `json:"cilium,omitempty"`
	LyftVPC    *kops.LyftVPCNetworkingSpec `json:"-"`
	GCE        *GCENetworkingSpec          `json:"gce,omitempty"`
}

// KubenetNetworkingSpec is the specification for kubenet networking, largely integrated but intended to replace classic
type KubenetNetworkingSpec struct{}

// ExternalNetworkingSpec is the specification for networking that is implemented by a user-provided Daemonset that uses the Kubenet kubelet networking plugin.
type ExternalNetworkingSpec struct{}

// CNINetworkingSpec is the specification for networking that is implemented by a user-provided Daemonset, which uses the CNI kubelet networking plugin.
type CNINetworkingSpec struct {
	UsesSecondaryIP bool `json:"usesSecondaryIP,omitempty"`
}

// KopeioNetworkingSpec declares that we want Kopeio networking
type KopeioNetworkingSpec struct{}

// WeaveNetworkingSpec declares that we want Weave networking
type WeaveNetworkingSpec struct {
	MTU         *int32 `json:"mtu,omitempty"`
	ConnLimit   *int32 `json:"connLimit,omitempty"`
	NoMasqLocal *int32 `json:"noMasqLocal,omitempty"`

	// MemoryRequest memory request of weave container. Default 200Mi
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest CPU request of weave container. Default 50m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// MemoryLimit memory limit of weave container. Default 200Mi
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
	// CPULimit CPU limit of weave container.
	CPULimit *resource.Quantity `json:"cpuLimit,omitempty"`
	// NetExtraArgs are extra arguments that are passed to weave-kube.
	NetExtraArgs string `json:"netExtraArgs,omitempty"`

	// NPCMemoryRequest memory request of weave npc container. Default 200Mi
	NPCMemoryRequest *resource.Quantity `json:"npcMemoryRequest,omitempty"`
	// NPCCPURequest CPU request of weave npc container. Default 50m
	NPCCPURequest *resource.Quantity `json:"npcCPURequest,omitempty"`
	// NPCMemoryLimit memory limit of weave npc container. Default 200Mi
	NPCMemoryLimit *resource.Quantity `json:"npcMemoryLimit,omitempty"`
	// NPCCPULimit CPU limit of weave npc container
	NPCCPULimit *resource.Quantity `json:"npcCPULimit,omitempty"`
	// NPCExtraArgs are extra arguments that are passed to weave-npc.
	NPCExtraArgs string `json:"npcExtraArgs,omitempty"`

	// Version specifies the Weave container image tag. The default depends on the kOps version.
	Version string `json:"version,omitempty"`
}

// FlannelNetworkingSpec declares that we want Flannel networking
type FlannelNetworkingSpec struct {
	// Backend is the backend overlay type we want to use (vxlan or udp)
	Backend string `json:"backend,omitempty"`
	// IptablesResyncSeconds sets resync period for iptables rules, in seconds
	IptablesResyncSeconds *int32 `json:"iptablesResyncSeconds,omitempty"`
}

// CalicoNetworkingSpec declares that we want Calico networking
type CalicoNetworkingSpec struct {
	// Registry overrides the Calico container image registry.
	Registry string `json:"registry,omitempty"`
	// Version overrides the Calico container image tag.
	Version string `json:"version,omitempty"`

	// AllowIPForwarding enable ip_forwarding setting within the container namespace.
	// (default: false)
	AllowIPForwarding bool `json:"allowIPForwarding,omitempty"`
	// AWSSrcDstCheck enables/disables ENI source/destination checks (AWS only)
	// Options: Disable (default), Enable, or DoNothing
	AWSSrcDstCheck string `json:"awsSrcDstCheck,omitempty"`
	// BPFEnabled enables the eBPF dataplane mode.
	BPFEnabled bool `json:"bpfEnabled,omitempty"`
	// BPFExternalServiceMode controls how traffic from outside the cluster to NodePorts and ClusterIPs is handled.
	// In Tunnel mode, packet is tunneled from the ingress host to the host with the backing pod and back again.
	// In DSR mode, traffic is tunneled to the host with the backing pod and then returned directly;
	// this requires a network that allows direct return.
	// Default: Tunnel (other options: DSR)
	BPFExternalServiceMode string `json:"bpfExternalServiceMode,omitempty"`
	// BPFKubeProxyIptablesCleanupEnabled controls whether Felix will clean up the iptables rules
	// created by the Kubernetes kube-proxy; should only be enabled if kube-proxy is not running.
	BPFKubeProxyIptablesCleanupEnabled bool `json:"bpfKubeProxyIptablesCleanupEnabled,omitempty"`
	// BPFLogLevel controls the log level used by the BPF programs. The logs are emitted
	// to the BPF trace pipe, accessible with the command tc exec BPF debug.
	// Default: Off (other options: Info, Debug)
	BPFLogLevel string `json:"bpfLogLevel,omitempty"`
	// ChainInsertMode controls whether Felix inserts rules to the top of iptables chains, or
	// appends to the bottom. Leaving the default option is safest to prevent accidentally
	// breaking connectivity. Default: 'insert' (other options: 'append')
	ChainInsertMode string `json:"chainInsertMode,omitempty"`
	// CPURequest CPU request of Calico container. Default: 100m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// CrossSubnet is deprecated as of kOps 1.22 and has no effect
	CrossSubnet *bool `json:"-"`
	// EncapsulationMode specifies the network packet encapsulation protocol for Calico to use,
	// employing such encapsulation at the necessary scope per the related CrossSubnet field. In
	// "ipip" mode, Calico will use IP-in-IP encapsulation as needed. In "vxlan" mode, Calico will
	// encapsulate packets as needed using the VXLAN scheme.
	// Options: ipip (default) or vxlan
	EncapsulationMode string `json:"encapsulationMode,omitempty"`
	// IPIPMode determines when to use IP-in-IP encapsulation for the default Calico IPv4 pool.
	// It is conveyed to the "calico-node" daemon container via the CALICO_IPV4POOL_IPIP
	// environment variable. EncapsulationMode must be set to "ipip".
	// Options: "CrossSubnet", "Always", or "Never".
	// Default: "CrossSubnet" if EncapsulationMode is "ipip", "Never" otherwise.
	IPIPMode string `json:"ipipMode,omitempty"`
	// IPv4AutoDetectionMethod configures how Calico chooses the IP address used to route
	// between nodes.  This should be set when the host has multiple interfaces
	// and it is important to select the interface used.
	// Options: "first-found" (default), "can-reach=DESTINATION",
	// "interface=INTERFACE-REGEX", or "skip-interface=INTERFACE-REGEX"
	IPv4AutoDetectionMethod string `json:"ipv4AutoDetectionMethod,omitempty"`
	// IPv6AutoDetectionMethod configures how Calico chooses the IP address used to route
	// between nodes.  This should be set when the host has multiple interfaces
	// and it is important to select the interface used.
	// Options: "first-found" (default), "can-reach=DESTINATION",
	// "interface=INTERFACE-REGEX", or "skip-interface=INTERFACE-REGEX"
	IPv6AutoDetectionMethod string `json:"ipv6AutoDetectionMethod,omitempty"`
	// IptablesBackend controls which variant of iptables binary Felix uses
	// Default: Auto (other options: Legacy, NFT)
	IptablesBackend string `json:"iptablesBackend,omitempty"`
	// LogSeverityScreen lets us set the desired log level. (Default: info)
	LogSeverityScreen string `json:"logSeverityScreen,omitempty"`
	// MTU to be set in the cni-network-config for calico.
	MTU *int32 `json:"mtu,omitempty"`
	// PrometheusMetricsEnabled can be set to enable the experimental Prometheus
	// metrics server (default: false)
	PrometheusMetricsEnabled bool `json:"prometheusMetricsEnabled,omitempty"`
	// PrometheusMetricsPort is the TCP port that the experimental Prometheus
	// metrics server should bind to (default: 9091)
	PrometheusMetricsPort int32 `json:"prometheusMetricsPort,omitempty"`
	// PrometheusGoMetricsEnabled enables Prometheus Go runtime metrics collection
	PrometheusGoMetricsEnabled bool `json:"prometheusGoMetricsEnabled,omitempty"`
	// PrometheusProcessMetricsEnabled enables Prometheus process metrics collection
	PrometheusProcessMetricsEnabled bool `json:"prometheusProcessMetricsEnabled,omitempty"`
	// TyphaPrometheusMetricsEnabled enables Prometheus metrics collection from Typha
	// (default: false)
	TyphaPrometheusMetricsEnabled bool `json:"typhaPrometheusMetricsEnabled,omitempty"`
	// TyphaPrometheusMetricsPort is the TCP port the typha Prometheus metrics server
	// should bind to (default: 9093)
	TyphaPrometheusMetricsPort int32 `json:"typhaPrometheusMetricsPort,omitempty"`
	// TyphaReplicas is the number of replicas of Typha to deploy
	TyphaReplicas int32 `json:"typhaReplicas,omitempty"`
	// VXLANMode determines when to use VXLAN encapsulation for the default Calico IPv4 pool.
	// It is conveyed to the "calico-node" daemon container via the CALICO_IPV4POOL_VXLAN
	// environment variable. EncapsulationMode must be set to "vxlan".
	// Options: "CrossSubnet", "Always", or "Never".
	// Default: "CrossSubnet" if EncapsulationMode is "vxlan", "Never" otherwise.
	VXLANMode string `json:"vxlanMode,omitempty"`
	// WireguardEnabled enables WireGuard encryption for all on-the-wire pod-to-pod traffic
	// (default: false)
	WireguardEnabled bool `json:"wireguardEnabled,omitempty"`
}

// CanalNetworkingSpec declares that we want Canal networking
type CanalNetworkingSpec struct {
	// ChainInsertMode controls whether Felix inserts rules to the top of iptables chains, or
	// appends to the bottom. Leaving the default option is safest to prevent accidentally
	// breaking connectivity. Default: 'insert' (other options: 'append')
	ChainInsertMode string `json:"chainInsertMode,omitempty"`
	// CPURequest CPU request of Canal container. Default: 100m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// DefaultEndpointToHostAction allows users to configure the default behaviour
	// for traffic between pod to host after calico rules have been processed.
	// Default: ACCEPT (other options: DROP, RETURN)
	DefaultEndpointToHostAction string `json:"defaultEndpointToHostAction,omitempty"`
	// FlanneldIptablesForwardRules configures Flannel to add the
	// default ACCEPT traffic rules to the iptables FORWARD chain. (default: true)
	FlanneldIptablesForwardRules *bool `json:"flanneldIptablesForwardRules,omitempty"`
	// IptablesBackend controls which variant of iptables binary Felix uses
	// Default: Auto (other options: Legacy, NFT)
	IptablesBackend string `json:"iptablesBackend,omitempty"`
	// LogSeveritySys the severity to set for logs which are sent to syslog
	// Default: INFO (other options: DEBUG, WARNING, ERROR, CRITICAL, NONE)
	LogSeveritySys string `json:"logSeveritySys,omitempty"`
	// MTU to be set in the cni-network-config (default: 1500)
	MTU *int32 `json:"mtu,omitempty"`
	// PrometheusGoMetricsEnabled enables Prometheus Go runtime metrics collection
	PrometheusGoMetricsEnabled bool `json:"prometheusGoMetricsEnabled,omitempty"`
	// PrometheusMetricsEnabled can be set to enable the experimental Prometheus
	// metrics server (default: false)
	PrometheusMetricsEnabled bool `json:"prometheusMetricsEnabled,omitempty"`
	// PrometheusMetricsPort is the TCP port that the experimental Prometheus
	// metrics server should bind to (default: 9091)
	PrometheusMetricsPort int32 `json:"prometheusMetricsPort,omitempty"`
	// PrometheusProcessMetricsEnabled enables Prometheus process metrics collection
	PrometheusProcessMetricsEnabled bool `json:"prometheusProcessMetricsEnabled,omitempty"`
	// TyphaPrometheusMetricsEnabled enables Prometheus metrics collection from Typha
	// (default: false)
	TyphaPrometheusMetricsEnabled bool `json:"typhaPrometheusMetricsEnabled,omitempty"`
	// TyphaPrometheusMetricsPort is the TCP port the typha Prometheus metrics server
	// should bind to (default: 9093)
	TyphaPrometheusMetricsPort int32 `json:"typhaPrometheusMetricsPort,omitempty"`
	// TyphaReplicas is the number of replicas of Typha to deploy
	TyphaReplicas int32 `json:"typhaReplicas,omitempty"`
}

// KuberouterNetworkingSpec declares that we want Kube-router networking
type KuberouterNetworkingSpec struct{}

// AmazonVPCNetworkingSpec declares that we want Amazon VPC CNI networking
type AmazonVPCNetworkingSpec struct {
	// Image is the container image name to use.
	Image string `json:"image,omitempty"`
	// InitImage is the init container image name to use.
	InitImage string `json:"initImage,omitempty"`
	// Env is a list of environment variables to set in the container.
	Env []EnvVar `json:"env,omitempty"`
}

type CiliumEncryptionType string

// CiliumNetworkingSpec declares that we want Cilium networking
type CiliumNetworkingSpec struct {
	// Version is the version of the Cilium agent and the Cilium Operator.
	Version string `json:"version,omitempty"`

	// MemoryRequest memory request of Cilium agent + operator container. (default: 128Mi)
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest CPU request of Cilium agent + operator container. (default: 25m)
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`

	// AgentPrometheusPort is the port to listen to for Prometheus metrics.
	// Defaults to 9090.
	AgentPrometheusPort int `json:"agentPrometheusPort,omitempty"`
	// Metrics is a list of metrics to add or remove from the default list of metrics the agent exposes.
	Metrics []string `json:"metrics,omitempty"`

	// ChainingMode allows using Cilium in combination with other CNI plugins.
	// With Cilium CNI chaining, the base network connectivity and IP address management is managed
	// by the non-Cilium CNI plugin, but Cilium attaches eBPF programs to the network devices created
	// by the non-Cilium plugin to provide L3/L4 network visibility, policy enforcement and other advanced features.
	// Default: none
	ChainingMode string `json:"chainingMode,omitempty"`
	// Debug runs Cilium in debug mode.
	Debug bool `json:"debug,omitempty"`
	// DisableEndpointCRD disables usage of CiliumEndpoint CRD.
	// Default: false
	DisableEndpointCRD bool `json:"disableEndpointCRD,omitempty"`
	// EnablePolicy specifies the policy enforcement mode.
	// "default": Follows Kubernetes policy enforcement.
	// "always": Cilium restricts all traffic if no policy is in place.
	// "never": Cilium allows all traffic regardless of policies in place.
	// If unspecified, "default" policy mode will be used.
	EnablePolicy string `json:"enablePolicy,omitempty"`
	// EnableL7Proxy enables L7 proxy for L7 policy enforcement.
	// Default: true
	EnableL7Proxy *bool `json:"enableL7Proxy,omitempty"`
	// EnableBPFMasquerade enables masquerading packets from endpoints leaving the host with BPF instead of iptables.
	// Default: false
	EnableBPFMasquerade *bool `json:"enableBPFMasquerade,omitempty"`
	// EnableEndpointHealthChecking enables connectivity health checking between virtual endpoints.
	// Default: true
	EnableEndpointHealthChecking *bool `json:"enableEndpointHealthChecking,omitempty"`
	// EnablePrometheusMetrics enables the Cilium "/metrics" endpoint for both the agent and the operator.
	EnablePrometheusMetrics bool `json:"enablePrometheusMetrics,omitempty"`
	// EnableEncryption enables Cilium Encryption.
	// Default: false
	EnableEncryption bool `json:"enableEncryption,omitempty"`
	// EncryptionType specifies Cilium Encryption method ("ipsec", "wireguard").
	// Default: ipsec
	EncryptionType CiliumEncryptionType `json:"encryptionType,omitempty"`
	// IdentityAllocationMode specifies in which backend identities are stored ("crd", "kvstore").
	// Default: crd
	IdentityAllocationMode string `json:"identityAllocationMode,omitempty"`
	// IdentityChangeGracePeriod specifies the duration to wait before using a changed identity.
	// Default: 5s
	IdentityChangeGracePeriod string `json:"identityChangeGracePeriod,omitempty"`
	// Masquerade enables masquerading IPv4 traffic to external destinations behind the node IP.
	// Default: false if IPAM is "eni" or in IPv6 mode, otherwise true
	Masquerade *bool `json:"masquerade,omitempty"`
	// AgentPodAnnotations makes possible to add additional annotations to the cilium agent.
	// Default: none
	AgentPodAnnotations map[string]string `json:"agentPodAnnotations,omitempty"`
	// Tunnel specifies the Cilium tunnelling mode. Possible values are "vxlan", "geneve", or "disabled".
	// Default: vxlan
	Tunnel string `json:"tunnel,omitempty"`
	// MonitorAggregation sets the level of packet monitoring. Possible values are "low", "medium", or "maximum".
	// Default: medium
	MonitorAggregation string `json:"monitorAggregation,omitempty"`
	// BPFCTGlobalTCPMax is the maximum number of entries in the TCP CT table.
	// Default: 524288
	BPFCTGlobalTCPMax int `json:"bpfCTGlobalTCPMax,omitempty"`
	// BPFCTGlobalAnyMax is the maximum number of entries in the non-TCP CT table.
	// Default: 262144
	BPFCTGlobalAnyMax int `json:"bpfCTGlobalAnyMax,omitempty"`
	// BPFLBAlgorithm is the load balancing algorithm ("random", "maglev").
	// Default: random
	BPFLBAlgorithm string `json:"bpfLBAlgorithm,omitempty"`
	// BPFLBMaglevTableSize is the per service backend table size when going with Maglev (parameter M).
	// Default: 16381
	BPFLBMaglevTableSize string `json:"bpfLBMaglevTableSize,omitempty"`
	// BPFNATGlobalMax is the the maximum number of entries in the BPF NAT table.
	// Default: 524288
	BPFNATGlobalMax int `json:"bpfNATGlobalMax,omitempty"`
	// BPFNeighGlobalMax is the the maximum number of entries in the BPF Neighbor table.
	// Default: 524288
	BPFNeighGlobalMax int `json:"bpfNeighGlobalMax,omitempty"`
	// BPFPolicyMapMax is the maximum number of entries in endpoint policy map.
	// Default: 16384
	BPFPolicyMapMax int `json:"bpfPolicyMapMax,omitempty"`
	// BPFLBMapMax is the maximum number of entries in bpf lb service, backend and affinity maps.
	// Default: 65536
	BPFLBMapMax int `json:"bpfLBMapMax,omitempty"`
	// BPFLBSockHostNSOnly enables skipping socket LB for services when inside a pod namespace,
	// in favor of service LB at the pod interface. Socket LB is still used when in the host namespace.
	// Required by service mesh (e.g., Istio, Linkerd).
	// Default: false
	BPFLBSockHostNSOnly bool `json:"bpfLBSockHostNSOnly,omitempty"`
	// PreallocateBPFMaps reduces the per-packet latency at the expense of up-front memory allocation.
	// Default: true
	PreallocateBPFMaps bool `json:"preallocateBPFMaps,omitempty"`
	// SidecarIstioProxyImage is the regular expression matching compatible Istio sidecar istio-proxy
	// container image names.
	// Default: cilium/istio_proxy
	SidecarIstioProxyImage string `json:"sidecarIstioProxyImage,omitempty"`
	// ClusterName is the name of the cluster. It is only relevant when building a mesh of clusters.
	ClusterName string `json:"clusterName,omitempty"`
	// ToFQDNsDNSRejectResponseCode sets the DNS response code for rejecting DNS requests.
	// Possible values are "nameError" or "refused".
	// Default: refused
	ToFQDNsDNSRejectResponseCode string `json:"toFQDNsDNSRejectResponseCode,omitempty"`
	// ToFQDNsEnablePoller replaces the DNS proxy-based implementation of FQDN policies
	// with the less powerful legacy implementation.
	// Default: false
	ToFQDNsEnablePoller bool `json:"toFQDNsEnablePoller,omitempty"`
	// IPAM specifies the IP address allocation mode to use.
	// Possible values are "crd" and "eni".
	// "eni" will use AWS native networking for pods. Eni requires masquerade to be set to false.
	// "crd" will use CRDs for controlling IP address management.
	// "hostscope" will use hostscope IPAM mode.
	// "kubernetes" will use addersing based on node pod CIDR.
	// Default: "kubernetes".
	IPAM string `json:"ipam,omitempty"`
	// InstallIptablesRules enables installing the base IPTables rules used for masquerading and kube-proxy.
	// Default: true
	InstallIptablesRules *bool `json:"installIptablesRules,omitempty"`
	// AutoDirectNodeRoutes adds automatic L2 routing between nodes.
	// Default: false
	AutoDirectNodeRoutes bool `json:"autoDirectNodeRoutes,omitempty"`
	// EnableHostReachableServices configures Cilium to enable services to be
	// reached from the host namespace in addition to pod namespaces.
	// https://docs.cilium.io/en/v1.9/gettingstarted/host-services/
	// Default: false
	EnableHostReachableServices bool `json:"enableHostReachableServices,omitempty"`
	// EnableNodePort replaces kube-proxy with Cilium's BPF implementation.
	// Requires spec.kubeProxy.enabled be set to false.
	// Default: false
	EnableNodePort bool `json:"enableNodePort,omitempty"`
	// EtcdManagd installs an additional etcd cluster that is used for Cilium state change.
	// The cluster is operated by cilium-etcd-operator.
	// Default: false
	EtcdManaged bool `json:"etcdManaged,omitempty"`
	// EnableRemoteNodeIdentity enables the remote-node-identity.
	// Default: true
	EnableRemoteNodeIdentity *bool `json:"enableRemoteNodeIdentity,omitempty"`
	// Hubble configures the Hubble service on the Cilium agent.
	Hubble *HubbleSpec `json:"hubble,omitempty"`

	// DisableCNPStatusUpdates determines if CNP NodeStatus updates will be sent to the Kubernetes api-server.
	DisableCNPStatusUpdates *bool `json:"disableCNPStatusUpdates,omitempty"`

	// EnableServiceTopology determine if cilium should use topology aware hints.
	EnableServiceTopology bool `json:"enableServiceTopology,omitempty"`
}

// HubbleSpec configures the Hubble service on the Cilium agent.
type HubbleSpec struct {
	// Enabled decides if Hubble is enabled on the agent or not
	Enabled *bool `json:"enabled,omitempty"`

	// Metrics is a list of metrics to collect. If empty or null, metrics are disabled.
	// See https://docs.cilium.io/en/stable/configuration/metrics/#hubble-exported-metrics
	Metrics []string `json:"metrics,omitempty"`
}

// GCENetworkingSpec is the specification of GCE's native networking mode, using IP aliases
type GCENetworkingSpec struct{}
