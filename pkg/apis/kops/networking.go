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

import "k8s.io/apimachinery/pkg/api/resource"

// NetworkingSpec allows selection and configuration of a networking plugin
type NetworkingSpec struct {
	Classic    *ClassicNetworkingSpec    `json:"classic,omitempty"`
	Kubenet    *KubenetNetworkingSpec    `json:"kubenet,omitempty"`
	External   *ExternalNetworkingSpec   `json:"external,omitempty"`
	CNI        *CNINetworkingSpec        `json:"cni,omitempty"`
	Kopeio     *KopeioNetworkingSpec     `json:"kopeio,omitempty"`
	Weave      *WeaveNetworkingSpec      `json:"weave,omitempty"`
	Flannel    *FlannelNetworkingSpec    `json:"flannel,omitempty"`
	Calico     *CalicoNetworkingSpec     `json:"calico,omitempty"`
	Canal      *CanalNetworkingSpec      `json:"canal,omitempty"`
	Kuberouter *KuberouterNetworkingSpec `json:"kuberouter,omitempty"`
	Romana     *RomanaNetworkingSpec     `json:"romana,omitempty"`
	AmazonVPC  *AmazonVPCNetworkingSpec  `json:"amazonvpc,omitempty"`
	Cilium     *CiliumNetworkingSpec     `json:"cilium,omitempty"`
	LyftVPC    *LyftVPCNetworkingSpec    `json:"lyftvpc,omitempty"`
	GCE        *GCENetworkingSpec        `json:"gce,omitempty"`
}

// ClassicNetworkingSpec is the specification of classic networking mode, integrated into kubernetes
type ClassicNetworkingSpec struct {
}

// KubenetNetworkingSpec is the specification for kubenet networking, largely integrated but intended to replace classic
type KubenetNetworkingSpec struct {
}

// ExternalNetworkingSpec is the specification for networking that is implemented by a Daemonset
// It also uses kubenet
type ExternalNetworkingSpec struct {
}

// CNINetworkingSpec is the specification for networking that is implemented by a Daemonset
// Networking is not managed by kops - we can create options here that directly configure e.g. weave
// but this is useful for arbitrary network modes or for modes that don't need additional configuration.
type CNINetworkingSpec struct {
	UsesSecondaryIP bool `json:"usesSecondaryIP,omitempty"`
}

// KopeioNetworkingSpec declares that we want Kopeio networking
type KopeioNetworkingSpec struct {
}

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
	// ChainInsertMode controls whether Felix inserts rules to the top of iptables chains, or
	// appends to the bottom. Leaving the default option is safest to prevent accidentally
	// breaking connectivity. Default: 'insert' (other options: 'append')
	ChainInsertMode string `json:"chainInsertMode,omitempty"`
	// CrossSubnet enables Calico's cross-subnet mode when set to true
	CrossSubnet bool `json:"crossSubnet,omitempty"`
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
	// MajorVersion is the version of Calico to use
	MajorVersion string `json:"majorVersion,omitempty"`
	// IptablesBackend controls which variant of iptables binary Felix uses
	// Default: Auto (other options: Legacy, NFT)
	IptablesBackend string `json:"iptablesBackend,omitempty"`
	// IPIPMode is mode for CALICO_IPV4POOL_IPIP
	IPIPMode string `json:"ipipMode,omitempty"`
	// TyphaPrometheusMetricsEnabled enables Prometheus metrics collection from Typha
	// (default: false)
	TyphaPrometheusMetricsEnabled bool `json:"typhaPrometheusMetricsEnabled,omitempty"`
	// TyphaPrometheusMetricsPort is the TCP port the typha Prometheus metrics server
	// should bind to (default: 9093)
	TyphaPrometheusMetricsPort int32 `json:"typhaPrometheusMetricsPort,omitempty"`
	// TyphaReplicas is the number of replicas of Typha to deploy
	TyphaReplicas int32 `json:"typhaReplicas,omitempty"`
}

// CanalNetworkingSpec declares that we want Canal networking
type CanalNetworkingSpec struct {
	// ChainInsertMode controls whether Felix inserts rules to the top of iptables chains, or
	// appends to the bottom. Leaving the default option is safest to prevent accidentally
	// breaking connectivity. Default: 'insert' (other options: 'append')
	ChainInsertMode string `json:"chainInsertMode,omitempty"`
	// DefaultEndpointToHostAction allows users to configure the default behaviour
	// for traffic between pod to host after calico rules have been processed.
	// Default: ACCEPT (other options: DROP, RETURN)
	DefaultEndpointToHostAction string `json:"defaultEndpointToHostAction,omitempty"`
	// DisableFlannelForwardRules configures Flannel to NOT add the
	// default ACCEPT traffic rules to the iptables FORWARD chain
	DisableFlannelForwardRules bool `json:"disableFlannelForwardRules,omitempty"`
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
type KuberouterNetworkingSpec struct {
}

// RomanaNetworkingSpec declares that we want Romana networking
type RomanaNetworkingSpec struct {
	// DaemonServiceIP is the Kubernetes Service IP for the romana-daemon pod
	DaemonServiceIP string `json:"daemonServiceIP,omitempty"`
	// EtcdServiceIP is the Kubernetes Service IP for the etcd backend used by Romana
	EtcdServiceIP string `json:"etcdServiceIP,omitempty"`
}

// AmazonVPCNetworkingSpec declares that we want Amazon VPC CNI networking
type AmazonVPCNetworkingSpec struct {
	// The container image name to use
	ImageName string `json:"imageName,omitempty"`
	// Env is a list of environment variables to set in the container.
	Env []EnvVar `json:"env,omitempty"`
}

const CiliumIpamEni = "eni"

// CiliumNetworkingSpec declares that we want Cilium networking
type CiliumNetworkingSpec struct {
	// Version is the version of the Cilium agent and the Cilium Operator.
	Version string `json:"version,omitempty"`

	// AccessLog is not implemented and may be removed in the future.
	// Setting this has no effect.
	AccessLog string `json:"accessLog,omitempty"`
	// AgentLabels is not implemented and may be removed in the future.
	// Setting this has no effect.
	AgentLabels []string `json:"agentLabels,omitempty"`
	// AgentPrometheusPort is the port to listen to for Prometheus metrics.
	// Defaults to 9090.
	AgentPrometheusPort int `json:"agentPrometheusPort,omitempty"`
	// AllowLocalhost is not implemented and may be removed in the future.
	// Setting this has no effect.
	AllowLocalhost string `json:"allowLocalhost,omitempty"`
	// AutoIpv6NodeRoutes is not implemented and may be removed in the future.
	// Setting this has no effect.
	AutoIpv6NodeRoutes bool `json:"autoIpv6NodeRoutes,omitempty"`
	// BPFRoot is not implemented and may be removed in the future.
	// Setting this has no effect.
	BPFRoot string `json:"bpfRoot,omitempty"`
	// ContainerRuntime is not implemented and may be removed in the future.
	// Setting this has no effect.
	ContainerRuntime []string `json:"containerRuntime,omitempty"`
	// ContainerRuntimeEndpoint is not implemented and may be removed in the future.
	// Setting this has no effect.
	ContainerRuntimeEndpoint map[string]string `json:"containerRuntimeEndpoint,omitempty"`
	// Debug runs Cilium in debug mode.
	Debug bool `json:"debug,omitempty"`
	// DebugVerbose is not implemented and may be removed in the future.
	// Setting this has no effect.
	DebugVerbose []string `json:"debugVerbose,omitempty"`
	// Device is not implemented and may be removed in the future.
	// Setting this has no effect.
	Device string `json:"device,omitempty"`
	// DisableConntrack is not implemented and may be removed in the future.
	// Setting this has no effect.
	DisableConntrack bool `json:"disableConntrack,omitempty"`
	// DisableIpv4 is deprecated: Use EnableIpv4 instead.
	// Setting this flag has no effect.
	DisableIpv4 bool `json:"disableIpv4,omitempty"`
	// DisableK8sServices is not implemented and may be removed in the future.
	// Setting this has no effect.
	DisableK8sServices bool `json:"disableK8sServices,omitempty"`
	// EnablePolicy specifies the policy enforcement mode.
	// "default": Follows Kubernetes policy enforcement.
	// "always": Cilium restricts all traffic if no policy is in place.
	// "never": Cilium allows all traffic regardless of policies in place.
	// If unspecified, "default" policy mode will be used.
	EnablePolicy string `json:"enablePolicy,omitempty"`
	// EnableTracing is not implemented and may be removed in the future.
	// Setting this has no effect.
	EnableTracing bool `json:"enableTracing,omitempty"`
	// EnablePrometheusMetrics enables the Cilium "/metrics" endpoint for both the agent and the operator.
	EnablePrometheusMetrics bool `json:"enablePrometheusMetrics,omitempty"`
	// EnvoyLog is not implemented and may be removed in the future.
	// Setting this has no effect.
	EnvoyLog string `json:"envoyLog,omitempty"`
	// Ipv4ClusterCIDRMaskSize is not implemented and may be removed in the future.
	// Setting this has no effect.
	Ipv4ClusterCIDRMaskSize int `json:"ipv4ClusterCidrMaskSize,omitempty"`
	// Ipv4Node is not implemented and may be removed in the future.
	// Setting this has no effect.
	Ipv4Node string `json:"ipv4Node,omitempty"`
	// Ipv4Range is not implemented and may be removed in the future.
	// Setting this has no effect.
	Ipv4Range string `json:"ipv4Range,omitempty"`
	// Ipv4ServiceRange is not implemented and may be removed in the future.
	// Setting this has no effect.
	Ipv4ServiceRange string `json:"ipv4ServiceRange,omitempty"`
	// Ipv6ClusterAllocCidr is not implemented and may be removed in the future.
	// Setting this has no effect.
	Ipv6ClusterAllocCidr string `json:"ipv6ClusterAllocCidr,omitempty"`
	// Ipv6Node is not implemented and may be removed in the future.
	// Setting this has no effect.
	Ipv6Node string `json:"ipv6Node,omitempty"`
	// Ipv6Range is not implemented and may be removed in the future.
	// Setting this has no effect.
	Ipv6Range string `json:"ipv6Range,omitempty"`
	// Ipv6ServiceRange is not implemented and may be removed in the future.
	// Setting this has no effect.
	Ipv6ServiceRange string `json:"ipv6ServiceRange,omitempty"`
	// K8sAPIServer is not implemented and may be removed in the future.
	// Setting this has no effect.
	K8sAPIServer string `json:"k8sApiServer,omitempty"`
	// K8sKubeconfigPath is not implemented and may be removed in the future.
	// Setting this has no effect.
	K8sKubeconfigPath string `json:"k8sKubeconfigPath,omitempty"`
	// KeepBPFTemplates is not implemented and may be removed in the future.
	// Setting this has no effect.
	KeepBPFTemplates bool `json:"keepBpfTemplates,omitempty"`
	// KeepConfig is not implemented and may be removed in the future.
	// Setting this has no effect.
	KeepConfig bool `json:"keepConfig,omitempty"`
	// LabelPrefixFile is not implemented and may be removed in the future.
	// Setting this has currently no effect
	LabelPrefixFile string `json:"labelPrefixFile,omitempty"`
	// Labels is not implemented and may be removed in the future.
	// Setting this has no effect.
	Labels []string `json:"labels,omitempty"`
	// LB is not implemented and may be removed in the future.
	// Setting this has no effect.
	LB string `json:"lb,omitempty"`
	// LibDir is not implemented and may be removed in the future.
	// Setting this has no effect.
	LibDir string `json:"libDir,omitempty"`
	// LogDrivers is not implemented and may be removed in the future.
	// Setting this has no effect.
	LogDrivers []string `json:"logDriver,omitempty"`
	// LogOpt is not implemented and may be removed in the future.
	// Setting this has no effect.
	LogOpt map[string]string `json:"logOpt,omitempty"`
	// Logstash is not implemented and may be removed in the future.
	// Setting this has no effect.
	Logstash bool `json:"logstash,omitempty"`
	// LogstashAgent is not implemented and may be removed in the future.
	// Setting this has no effect.
	LogstashAgent string `json:"logstashAgent,omitempty"`
	// LogstashProbeTimer is not implemented and may be removed in the future.
	// Setting this has no effect.
	LogstashProbeTimer uint32 `json:"logstashProbeTimer,omitempty"`
	// DisableMasquerade disables masquerading traffic to external destinations behind the node IP.
	DisableMasquerade bool `json:"disableMasquerade,omitempty"`
	// Nat6Range is not implemented and may be removed in the future.
	// Setting this has no effect.
	Nat46Range string `json:"nat46Range,omitempty"`
	// Pprof is not implemented and may be removed in the future.
	// Setting this has no effect.
	Pprof bool `json:"pprof,omitempty"`
	// PrefilterDevice is not implemented and may be removed in the future.
	// Setting this has no effect.
	PrefilterDevice string `json:"prefilterDevice,omitempty"`
	// PrometheusServeAddr is deprecated. Use EnablePrometheusMetrics and AgentPrometheusPort instead.
	// Setting this has no effect.
	PrometheusServeAddr string `json:"prometheusServeAddr,omitempty"`
	// Restore is not implemented and may be removed in the future.
	// Setting this has no effect.
	Restore bool `json:"restore,omitempty"`
	// SingleClusterRoute is not implemented and may be removed in the future.
	// Setting this has no effect.
	SingleClusterRoute bool `json:"singleClusterRoute,omitempty"`
	// SocketPath is not implemented and may be removed in the future.
	// Setting this has no effect.
	SocketPath string `json:"socketPath,omitempty"`
	// StateDir is not implemented and may be removed in the future.
	// Setting this has no effect.
	StateDir string `json:"stateDir,omitempty"`
	// TracePayloadLen is not implemented and may be removed in the future.
	// Setting this has no effect.
	TracePayloadLen int `json:"tracePayloadlen,omitempty"`
	// Tunnel specifies the Cilium tunelling mode. Possible values are "vxlan", "geneve", or "disabled".
	// Default: vxlan
	Tunnel string `json:"tunnel,omitempty"`
	// EnableIpv6 enables cluster IPv6 traffic. If both EnableIpv6 and EnableIpv4 are set to false
	// then IPv4 will be enabled.
	// Default: false
	EnableIpv6 bool `json:"enableipv6"`
	// EnableIpv4 enables cluster IPv4 traffic. If both EnableIpv6 and EnableIpv4 are set to false
	// then IPv4 will be enabled.
	// Default: false
	EnableIpv4 bool `json:"enableipv4"`
	// MonitorAggregation sets the level of packet monitoring. Possible values are "low", "medium", or "maximum".
	// Default: medium
	MonitorAggregation string `json:"monitorAggregation"`
	// BPFCTGlobalTCPMax is the maximum number of entries in the TCP CT table.
	// Default: 524288
	BPFCTGlobalTCPMax int `json:"bpfCTGlobalTCPMax"`
	// BPFCTGlobalAnyMax is the maximum number of entries in the non-TCP CT table.
	// Default: 262144
	BPFCTGlobalAnyMax int `json:"bpfCTGlobalAnyMax"`
	// PreallocateBPFMaps reduces the per-packet latency at the expense of up-front memory allocation.
	// Default: true
	PreallocateBPFMaps bool `json:"preallocateBPFMaps"`
	// SidecarIstioProxyImage is the regular expression matching compatible Istio sidecar istio-proxy
	// container image names.
	// Default: cilium/istio_proxy
	SidecarIstioProxyImage string `json:"sidecarIstioProxyImage"`
	// ClusterName is the name of the cluster. It is only relevant when building a mesh of clusters.
	ClusterName string `json:"clusterName"`
	// ToFqdnsDNSRejectResponseCode sets the DNS response code for rejecting DNS requests.
	// Possible values are "nameError" or "refused".
	// Default: refused
	ToFqdnsDNSRejectResponseCode string `json:"toFqdnsDnsRejectResponseCode,omitempty"`
	// ToFqdnsEnablePoller replaces the DNS proxy-based implementation of FQDN policies
	// with the less powerful legacy implementation.
	// Default: false
	ToFqdnsEnablePoller bool `json:"toFqdnsEnablePoller"`
	// ContainerRuntimeLabels enables fetching of container-runtime labels from the specified container runtime and associating them with endpoints.
	// Supported values are: "none", "containerd", "crio", "docker", "auto"
	// As of Cilium 1.7.0, Cilium no longer fetches information from the
	// container runtime and this field is ignored.
	// Default: none
	ContainerRuntimeLabels string `json:"containerRuntimeLabels,omitempty"`
	// Ipam specifies the IP address allocation mode to use.
	// Possible values are "crd" and "eni".
	// "eni" will use AWS native networking for pods. Eni requires masquerade to be set to false.
	// "crd" will use CRDs for controlling IP address management.
	// Empty value will use host-scope address management.
	Ipam string `json:"ipam,omitempty"`
	// IPTablesRulesNoinstall disables installing the base IPTables rules used for masquerading and kube-proxy.
	// Default: false
	IPTablesRulesNoinstall bool `json:"IPTablesRulesNoinstall"`
	// AutoDirectNodeRoutes adds automatic L2 routing between nodes.
	// Default: false
	AutoDirectNodeRoutes bool `json:"autoDirectNodeRoutes"`
	// EnableNodePort replaces kube-proxy with Cilium's BPF implementation.
	// Requires spec.kubeProxy.enabled be set to false.
	// Default: false
	EnableNodePort bool `json:"enableNodePort"`
	// EtcdManagd installs an additional etcd cluster that is used for Cilium state change.
	// The cluster is operated by cilium-etcd-operator.
	// Default: false
	EtcdManaged bool `json:"etcdManaged,omitempty"`
	// EnableRemoteNodeIdentity enables the remote-node-identity added in Cilium 1.7.0.
	// Default: false
	EnableRemoteNodeIdentity bool `json:"enableRemoteNodeIdentity"`

	// RemoveCbrBridge is not implemented and may be removed in the future.
	// Setting this has no effect.
	RemoveCbrBridge bool `json:"removeCbrBridge"`
	// RestartPods is not implemented and may be removed in the future.
	// Setting this has no effect.
	RestartPods bool `json:"restartPods"`
	// ReconfigureKubelet is not implemented and may be removed in the future.
	// Setting this has no effect.
	ReconfigureKubelet bool `json:"reconfigureKubelet"`
	// NodeInitBootstrapFile is not implemented and may be removed in the future.
	// Setting this has no effect.
	NodeInitBootstrapFile string `json:"nodeInitBootstrapFile"`
	// CniBinPath is not implemented and may be removed in the future.
	// Setting this has no effect.
	CniBinPath string `json:"cniBinPath"`
}

// LyftVPCNetworkingSpec declares that we want to use the cni-ipvlan-vpc-k8s CNI networking.
type LyftVPCNetworkingSpec struct {
	SubnetTags map[string]string `json:"subnetTags,omitempty"`
}

// GCENetworkingSpec is the specification of GCE's native networking mode, using IP aliases
type GCENetworkingSpec struct {
}
