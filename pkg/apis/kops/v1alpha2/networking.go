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
	MTU          *int32 `json:"mtu,omitempty"`
	ConnLimit    *int32 `json:"connLimit,omitempty"`
	NoMasqLocal  *int32 `json:"noMasqLocal,omitempty"`
	NetExtraArgs string `json:"netExtraArgs,omitempty"`
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
	CrossSubnet bool `json:"crossSubnet,omitempty"` // Enables Calico's cross-subnet mode when set to true
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
	// The container image name to use, which by default is:
	// 602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon-k8s-cni:v1.5.5
	ImageName string `json:"imageName,omitempty"`
}

// CiliumNetworkingSpec declares that we want Cilium networking
type CiliumNetworkingSpec struct {
	Version string `json:"version,omitempty"`

	AccessLog                string            `json:"accessLog,omitempty"`
	AgentLabels              []string          `json:"agentLabels,omitempty"`
	AllowLocalhost           string            `json:"allowLocalhost,omitempty"`
	AutoIpv6NodeRoutes       bool              `json:"autoIpv6NodeRoutes,omitempty"`
	BPFRoot                  string            `json:"bpfRoot,omitempty"`
	ContainerRuntime         []string          `json:"containerRuntime,omitempty"`
	ContainerRuntimeEndpoint map[string]string `json:"containerRuntimeEndpoint,omitempty"`
	Debug                    bool              `json:"debug,omitempty"`
	DebugVerbose             []string          `json:"debugVerbose,omitempty"`
	Device                   string            `json:"device,omitempty"`
	DisableConntrack         bool              `json:"disableConntrack,omitempty"`
	DisableIpv4              bool              `json:"disableIpv4,omitempty"`
	DisableK8sServices       bool              `json:"disableK8sServices,omitempty"`
	EnablePolicy             string            `json:"enablePolicy,omitempty"`
	EnableTracing            bool              `json:"enableTracing,omitempty"`
	EnvoyLog                 string            `json:"envoyLog,omitempty"`
	Ipv4ClusterCIDRMaskSize  int               `json:"ipv4ClusterCidrMaskSize,omitempty"`
	Ipv4Node                 string            `json:"ipv4Node,omitempty"`
	Ipv4Range                string            `json:"ipv4Range,omitempty"`
	Ipv4ServiceRange         string            `json:"ipv4ServiceRange,omitempty"`
	Ipv6ClusterAllocCidr     string            `json:"ipv6ClusterAllocCidr,omitempty"`
	Ipv6Node                 string            `json:"ipv6Node,omitempty"`
	Ipv6Range                string            `json:"ipv6Range,omitempty"`
	Ipv6ServiceRange         string            `json:"ipv6ServiceRange,omitempty"`
	K8sAPIServer             string            `json:"k8sApiServer,omitempty"`
	K8sKubeconfigPath        string            `json:"k8sKubeconfigPath,omitempty"`
	KeepBPFTemplates         bool              `json:"keepBpfTemplates,omitempty"`
	KeepConfig               bool              `json:"keepConfig,omitempty"`
	LabelPrefixFile          string            `json:"labelPrefixFile,omitempty"`
	Labels                   []string          `json:"labels,omitempty"`
	LB                       string            `json:"lb,omitempty"`
	LibDir                   string            `json:"libDir,omitempty"`
	LogDrivers               []string          `json:"logDriver,omitempty"`
	LogOpt                   map[string]string `json:"logOpt,omitempty"`
	Logstash                 bool              `json:"logstash,omitempty"`
	LogstashAgent            string            `json:"logstashAgent,omitempty"`
	LogstashProbeTimer       uint32            `json:"logstashProbeTimer,omitempty"`
	DisableMasquerade        bool              `json:"disableMasquerade,omitempty"`
	Nat46Range               string            `json:"nat46Range,omitempty"`
	Pprof                    bool              `json:"pprof,omitempty"`
	PrefilterDevice          string            `json:"prefilterDevice,omitempty"`
	PrometheusServeAddr      string            `json:"prometheusServeAddr,omitempty"`
	Restore                  bool              `json:"restore,omitempty"`
	SingleClusterRoute       bool              `json:"singleClusterRoute,omitempty"`
	SocketPath               string            `json:"socketPath,omitempty"`
	StateDir                 string            `json:"stateDir,omitempty"`
	TracePayloadLen          int               `json:"tracePayloadlen,omitempty"`
	Tunnel                   string            `json:"tunnel,omitempty"`

	EnableIpv6             bool   `json:"enableipv6"`
	EnableIpv4             bool   `json:"enableipv4"`
	MonitorAggregation     string `json:"monitorAggregation"`
	BPFCTGlobalTCPMax      int    `json:"bpfCTGlobalTCPMax"`
	BPFCTGlobalAnyMax      int    `json:"bpfCTGlobalAnyMax"`
	PreallocateBPFMaps     bool   `json:"preallocateBPFMaps"`
	SidecarIstioProxyImage string `json:"sidecarIstioProxyImage"`
	ClusterName            string `json:"clusterName"`
	ToFqdnsEnablePoller    bool   `json:"toFqdnsEnablePoller"`
	ContainerRuntimeLabels string `json:"containerRuntimeLabels,omitempty"`
	IPTablesRulesNoinstall bool   `json:"IPTablesRulesNoinstall"`
	AutoDirectNodeRoutes   bool   `json:"autoDirectNodeRoutes"`
	EnableNodePort         bool   `json:"enableNodePort"`

	//node init options
	RemoveCbrBridge       bool   `json:"removeCbrBridge"`
	RestartPods           bool   `json:"restartPods"`
	ReconfigureKubelet    bool   `json:"reconfigureKubelet"`
	NodeInitBootstrapFile string `json:"nodeInitBootstrapFile"`
	CniBinPath            string `json:"cniBinPath"`
}

// LyftIpVlanNetworkingSpec declares that we want to use the cni-ipvlan-vpc-k8s CNI networking
type LyftVPCNetworkingSpec struct {
	SubnetTags map[string]string `json:"subnetTags,omitempty"`
}

// GCENetworkingSpec is the specification of GCE's native networking mode, using IP aliases
type GCENetworkingSpec struct {
}
