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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeletConfigSpec defines the kubelet configuration
type KubeletConfigSpec struct {
	// APIServers is not used for clusters version 1.6 and later - flag removed
	APIServers string `json:"apiServers,omitempty" flag:"api-servers"`
	// AnonymousAuth permits you to control auth to the kubelet api
	AnonymousAuth *bool `json:"anonymousAuth,omitempty" flag:"anonymous-auth"`
	// AuthorizationMode is the authorization mode the kubelet is running in
	AuthorizationMode string `json:"authorizationMode,omitempty" flag:"authorization-mode"`
	// BootstrapKubeconfig is the path to a kubeconfig file that will be used to get client certificate for kubelet
	BootstrapKubeconfig string `json:"bootstrapKubeconfig,omitempty" flag:"bootstrap-kubeconfig"`
	// ClientCAFile is the path to a CA certificate
	ClientCAFile string `json:"clientCAFile,omitempty" flag:"client-ca-file"`
	// TODO: Remove unused TLSCertFile
	TLSCertFile string `json:"tlsCertFile,omitempty" flag:"tls-cert-file"`
	// TODO: Remove unused TLSPrivateKeyFile
	TLSPrivateKeyFile string `json:"tlsPrivateKeyFile,omitempty" flag:"tls-private-key-file"`
	// TLSCipherSuites indicates the allowed TLS cipher suite
	TLSCipherSuites []string `json:"tlsCipherSuites,omitempty" flag:"tls-cipher-suites"`
	// TLSMinVersion indicates the minimum TLS version allowed
	TLSMinVersion string `json:"tlsMinVersion,omitempty" flag:"tls-min-version"`
	// KubeconfigPath is the path of kubeconfig for the kubelet
	KubeconfigPath string `json:"kubeconfigPath,omitempty" flag:"kubeconfig"`
	// RequireKubeconfig indicates a kubeconfig is required
	RequireKubeconfig *bool `json:"requireKubeconfig,omitempty" flag:"require-kubeconfig"`
	// LogFormat is the logging format of the kubelet.
	// Supported values: text, json.
	// Default: text
	LogFormat string `json:"logFormat,omitempty" flag:"logging-format" flag-empty:"text"`
	// LogLevel is the logging level of the kubelet
	LogLevel *int32 `json:"logLevel,omitempty" flag:"v" flag-empty:"0"`
	// config is the path to the config file or directory of files
	PodManifestPath string `json:"podManifestPath,omitempty" flag:"pod-manifest-path"`
	// HostnameOverride is the hostname used to identify the kubelet instead of the actual hostname.
	HostnameOverride string `json:"hostnameOverride,omitempty" flag:"hostname-override"`
	// PodInfraContainerImage is the image whose network/ipc containers in each pod will use.
	PodInfraContainerImage string `json:"podInfraContainerImage,omitempty" flag:"pod-infra-container-image"`
	// SeccompDefault enables the use of `RuntimeDefault` as the default seccomp profile for all workloads.
	SeccompDefault *bool `json:"seccompDefault,omitempty" flag:"seccomp-default"`
	// SeccompProfileRoot is the directory path for seccomp profiles.
	SeccompProfileRoot *string `json:"seccompProfileRoot,omitempty" flag:"seccomp-profile-root"`
	// AllowPrivileged enables containers to request privileged mode (defaults to false)
	AllowPrivileged *bool `json:"allowPrivileged,omitempty" flag:"allow-privileged"`
	// EnableDebuggingHandlers enables server endpoints for log collection and local running of containers and commands
	EnableDebuggingHandlers *bool `json:"enableDebuggingHandlers,omitempty" flag:"enable-debugging-handlers"`
	// RegisterNode enables automatic registration with the apiserver.
	RegisterNode *bool `json:"registerNode,omitempty" flag:"register-node"`
	// NodeStatusUpdateFrequency Specifies how often kubelet posts node status to master (default 10s)
	// must work with nodeMonitorGracePeriod in KubeControllerManagerConfig.
	NodeStatusUpdateFrequency *metav1.Duration `json:"nodeStatusUpdateFrequency,omitempty" flag:"node-status-update-frequency"`
	// ClusterDomain is the DNS domain for this cluster
	ClusterDomain string `json:"clusterDomain,omitempty" flag:"cluster-domain"`
	// ClusterDNS is the IP address for a cluster DNS server
	ClusterDNS string `json:"clusterDNS,omitempty" flag:"cluster-dns"`
	// NetworkPluginName is the name of the network plugin to be invoked for various events in kubelet/pod lifecycle
	NetworkPluginName *string `json:"networkPluginName,omitempty" flag:"network-plugin"`
	// CloudProvider is the provider for cloud services.
	CloudProvider string `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	// KubeletCgroups is the absolute name of cgroups to isolate the kubelet in.
	KubeletCgroups string `json:"kubeletCgroups,omitempty" flag:"kubelet-cgroups"`
	// Cgroups that container runtime is expected to be isolated in.
	RuntimeCgroups string `json:"runtimeCgroups,omitempty" flag:"runtime-cgroups"`
	// ReadOnlyPort is the port used by the kubelet api for read-only access (default 10255)
	ReadOnlyPort *int32 `json:"readOnlyPort,omitempty" flag:"read-only-port"`
	// SystemCgroups is absolute name of cgroups in which to place
	// all non-kernel processes that are not already in a container. Empty
	// for no container. Rolling back the flag requires a reboot.
	SystemCgroups string `json:"systemCgroups,omitempty" flag:"system-cgroups"`
	// cgroupRoot is the root cgroup to use for pods. This is handled by the container runtime on a best effort basis.
	CgroupRoot string `json:"cgroupRoot,omitempty" flag:"cgroup-root"`
	// configureCBR0 enables the kubelet to configure cbr0 based on Node.Spec.PodCIDR.
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
	// The node has babysitter process monitoring docker and kubelet. Removed as of 1.7
	BabysitDaemons *bool `json:"babysitDaemons,omitempty" flag:"babysit-daemons"`
	// MaxPods is the number of pods that can run on this Kubelet.
	MaxPods *int32 `json:"maxPods,omitempty" flag:"max-pods"`
	// NvidiaGPUs is the number of NVIDIA GPU devices on this node.
	NvidiaGPUs int32 `json:"nvidiaGPUs,omitempty" flag:"experimental-nvidia-gpus" flag-empty:"0"`
	// PodCIDR is the CIDR to use for pod IP addresses, only used in standalone mode.
	// In cluster mode, this is obtained from the master.
	PodCIDR string `json:"podCIDR,omitempty" flag:"pod-cidr"`
	// ResolverConfig is the resolver configuration file used as the basis for the container DNS resolution configuration."), []
	ResolverConfig *string `json:"resolvConf,omitempty" flag:"resolv-conf" flag-include-empty:"true"`
	// ReconcileCIDR is Reconcile node CIDR with the CIDR specified by the
	// API server. No-op if register-node or configure-cbr0 is false.
	ReconcileCIDR *bool `json:"reconcileCIDR,omitempty" flag:"reconcile-cidr"`
	// registerSchedulable tells the kubelet to register the node as schedulable. No-op if register-node is false.
	RegisterSchedulable *bool `json:"registerSchedulable,omitempty" flag:"register-schedulable"`
	// SerializeImagePulls when enabled, tells the Kubelet to pull images one at a time.
	SerializeImagePulls *bool `json:"serializeImagePulls,omitempty" flag:"serialize-image-pulls"`
	// NodeLabels to add when registering the node in the cluster.
	NodeLabels map[string]string `json:"nodeLabels,omitempty" flag:"node-labels"`
	// NonMasqueradeCIDR configures masquerading: traffic to IPs outside this range will use IP masquerade.
	NonMasqueradeCIDR *string `json:"nonMasqueradeCIDR,omitempty" flag:"non-masquerade-cidr"`
	// Enable gathering custom metrics.
	EnableCustomMetrics *bool `json:"enableCustomMetrics,omitempty" flag:"enable-custom-metrics"`
	// NetworkPluginMTU is the MTU to be passed to the network plugin,
	// and overrides the default MTU for cases where it cannot be automatically
	// computed (such as IPSEC).
	NetworkPluginMTU *int32 `json:"networkPluginMTU,omitempty" flag:"network-plugin-mtu"`
	// imageMinimumGCAge is the minimum age for an unused image before it is garbage collected. Default: "2m"
	ImageMinimumGCAge *string `json:"imageMinimumGCAge,omitempty" flag:"image-minimum-gc-age"`
	// imageMaximumGCAge is the maximum age an image can be unused before it is garbage collected.
	// The default of this field is "0s", which disables this field--meaning images won't be garbage
	// collected based on being unused for too long. Default: "0s" (disabled)
	ImageMaximumGCAge *string `json:"imageMaximumGCAge,omitempty" flag:"image-maximum-gc-age"`
	// ImageGCHighThresholdPercent is the percent of disk usage after which
	// image garbage collection is always run.
	ImageGCHighThresholdPercent *int32 `json:"imageGCHighThresholdPercent,omitempty" flag:"image-gc-high-threshold"`
	// ImageGCLowThresholdPercent is the percent of disk usage before which
	// image garbage collection is never run. Lowest disk usage to garbage
	// collect to.
	ImageGCLowThresholdPercent *int32 `json:"imageGCLowThresholdPercent,omitempty" flag:"image-gc-low-threshold"`
	// ImagePullProgressDeadline is the timeout for image pulls
	// If no pulling progress is made before this deadline, the image pulling will be cancelled. (default 1m0s)
	ImagePullProgressDeadline *metav1.Duration `json:"imagePullProgressDeadline,omitempty" flag:"image-pull-progress-deadline"`
	// Comma-delimited list of hard eviction expressions.  For example, 'memory.available<300Mi'.
	EvictionHard *string `json:"evictionHard,omitempty" flag:"eviction-hard"`
	// Comma-delimited list of soft eviction expressions.  For example, 'memory.available<300Mi'.
	EvictionSoft string `json:"evictionSoft,omitempty" flag:"eviction-soft"`
	// Comma-delimited list of grace periods for each soft eviction signal.  For example, 'memory.available=30s'.
	EvictionSoftGracePeriod string `json:"evictionSoftGracePeriod,omitempty" flag:"eviction-soft-grace-period"`
	// Duration for which the kubelet has to wait before transitioning out of an eviction pressure condition.
	EvictionPressureTransitionPeriod *metav1.Duration `json:"evictionPressureTransitionPeriod,omitempty" flag:"eviction-pressure-transition-period" flag-empty:"0s"`
	// Maximum allowed grace period (in seconds) to use when terminating pods in response to a soft eviction threshold being met.
	EvictionMaxPodGracePeriod int32 `json:"evictionMaxPodGracePeriod,omitempty" flag:"eviction-max-pod-grace-period" flag-empty:"0"`
	// Comma-delimited list of minimum reclaims (e.g. imagefs.available=2Gi) that describes the minimum amount of resource the kubelet will reclaim when performing a pod eviction if that resource is under pressure.
	EvictionMinimumReclaim string `json:"evictionMinimumReclaim,omitempty" flag:"eviction-minimum-reclaim"`
	// The full path of the directory in which to search for additional third party volume plugins (this path must be writeable, dependent on your choice of OS)
	VolumePluginDirectory string `json:"volumePluginDirectory,omitempty" flag:"volume-plugin-dir"`
	// Taints to add when registering a node in the cluster
	Taints []string `json:"taints,omitempty" flag:"register-with-taints"`
	// FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
	FeatureGates map[string]string `json:"featureGates,omitempty" flag:"feature-gates"`
	// Integrate with the kernel memcg notification to determine if memory eviction thresholds are crossed rather than polling.
	KernelMemcgNotification *bool `json:"kernelMemcgNotification,omitempty" flag:"kernel-memcg-notification"`
	// Resource reservation for kubernetes system daemons like the kubelet, container runtime, node problem detector, etc.
	KubeReserved map[string]string `json:"kubeReserved,omitempty" flag:"kube-reserved"`
	// Control group for kube daemons.
	KubeReservedCgroup string `json:"kubeReservedCgroup,omitempty" flag:"kube-reserved-cgroup"`
	// Capture resource reservation for OS system daemons like sshd, udev, etc.
	SystemReserved map[string]string `json:"systemReserved,omitempty" flag:"system-reserved"`
	// Parent control group for OS system daemons.
	SystemReservedCgroup string `json:"systemReservedCgroup,omitempty" flag:"system-reserved-cgroup"`
	// Enforce Allocatable across pods whenever the overall usage across all pods exceeds Allocatable.
	EnforceNodeAllocatable string `json:"enforceNodeAllocatable,omitempty" flag:"enforce-node-allocatable"`
	// RuntimeRequestTimeout is timeout for runtime requests on - pull, logs, exec and attach
	RuntimeRequestTimeout *metav1.Duration `json:"runtimeRequestTimeout,omitempty" flag:"runtime-request-timeout"`
	// VolumeStatsAggPeriod is the interval for kubelet to calculate and cache the volume disk usage for all pods and volumes
	VolumeStatsAggPeriod *metav1.Duration `json:"volumeStatsAggPeriod,omitempty" flag:"volume-stats-agg-period"`
	// Tells the Kubelet to fail to start if swap is enabled on the node.
	FailSwapOn *bool `json:"failSwapOn,omitempty" flag:"fail-swap-on"`
	// ExperimentalAllowedUnsafeSysctls are passed to the kubelet config to whitelist allowable sysctls
	// Was promoted to beta and renamed. https://github.com/kubernetes/kubernetes/pull/63717
	ExperimentalAllowedUnsafeSysctls []string `json:"experimentalAllowedUnsafeSysctls,omitempty" flag:"experimental-allowed-unsafe-sysctls"`
	// AllowedUnsafeSysctls are passed to the kubelet config to whitelist allowable sysctls
	AllowedUnsafeSysctls []string `json:"allowedUnsafeSysctls,omitempty" flag:"allowed-unsafe-sysctls"`
	// StreamingConnectionIdleTimeout is the maximum time a streaming connection can be idle before the connection is automatically closed
	StreamingConnectionIdleTimeout *metav1.Duration `json:"streamingConnectionIdleTimeout,omitempty" flag:"streaming-connection-idle-timeout"`
	// DockerDisableSharedPID was removed.
	DockerDisableSharedPID *bool `json:"-"`
	// RootDir is the directory path for managing kubelet files (volume mounts,etc)
	RootDir string `json:"rootDir,omitempty" flag:"root-dir"`
	// AuthenticationTokenWebhook uses the TokenReview API to determine authentication for bearer tokens.
	AuthenticationTokenWebhook *bool `json:"authenticationTokenWebhook,omitempty" flag:"authentication-token-webhook"`
	// AuthenticationTokenWebhook sets the duration to cache responses from the webhook token authenticator. Default is 2m. (default 2m0s)
	AuthenticationTokenWebhookCacheTTL *metav1.Duration `json:"authenticationTokenWebhookCacheTTL,omitempty" flag:"authentication-token-webhook-cache-ttl"`
	// CPUCFSQuota enables CPU CFS quota enforcement for containers that specify CPU limits
	CPUCFSQuota *bool `json:"cpuCFSQuota,omitempty" flag:"cpu-cfs-quota"`
	// CPUCFSQuotaPeriod sets CPU CFS quota period value, cpu.cfs_period_us, defaults to Linux Kernel default
	CPUCFSQuotaPeriod *metav1.Duration `json:"cpuCFSQuotaPeriod,omitempty" flag:"cpu-cfs-quota-period"`
	// CpuManagerPolicy allows for changing the default policy of None to static
	CpuManagerPolicy string `json:"cpuManagerPolicy,omitempty" flag:"cpu-manager-policy"`
	// RegistryPullQPS if > 0, limit registry pull QPS to this value.  If 0, unlimited. (default 5)
	RegistryPullQPS *int32 `json:"registryPullQPS,omitempty" flag:"registry-qps"`
	// RegistryBurst Maximum size of a bursty pulls, temporarily allows pulls to burst to this number, while still not exceeding registry-qps. Only used if --registry-qps > 0 (default 10)
	RegistryBurst *int32 `json:"registryBurst,omitempty" flag:"registry-burst"`
	// TopologyManagerPolicy determines the allocation policy for the topology manager.
	TopologyManagerPolicy string `json:"topologyManagerPolicy,omitempty" flag:"topology-manager-policy"`
	// rotateCertificates enables client certificate rotation.
	RotateCertificates *bool `json:"rotateCertificates,omitempty" flag:"rotate-certificates"`
	// Default kubelet behaviour for kernel tuning. If set, kubelet errors if any of kernel tunables is different than kubelet defaults.
	// (DEPRECATED: This parameter should be set via the config file specified by the Kubelet's --config flag.
	ProtectKernelDefaults *bool `json:"protectKernelDefaults,omitempty" flag:"protect-kernel-defaults"`
	// CgroupDriver allows the explicit setting of the kubelet cgroup driver. If omitted, defaults to cgroupfs.
	CgroupDriver string `json:"cgroupDriver,omitempty" flag:"cgroup-driver"`
	// HousekeepingInterval allows to specify interval between container housekeepings.
	HousekeepingInterval *metav1.Duration `json:"housekeepingInterval,omitempty" flag:"housekeeping-interval"`
	// EventQPS if > 0, limit event creations per second to this value.  If 0, unlimited.
	EventQPS *int32 `json:"eventQPS,omitempty" flag:"event-qps" flag-empty:"0"`
	// EventBurst temporarily allows event records to burst to this number, while still not exceeding EventQPS. Only used if EventQPS > 0.
	EventBurst *int32 `json:"eventBurst,omitempty" flag:"event-burst"`
	// ContainerLogMaxSize is the maximum size (e.g. 10Mi) of container log file before it is rotated.
	ContainerLogMaxSize string `json:"containerLogMaxSize,omitempty" flag:"container-log-max-size"`
	// ContainerLogMaxFiles is the maximum number of container log files that can be present for a container. The number must be >= 2.
	ContainerLogMaxFiles *int32 `json:"containerLogMaxFiles,omitempty" flag:"container-log-max-files"`
	// EnableCadvisorJsonEndpoints enables cAdvisor json `/spec` and `/stats/*` endpoints. Defaults to False.
	EnableCadvisorJsonEndpoints *bool `json:"enableCadvisorJsonEndpoints,omitempty" flag:"enable-cadvisor-json-endpoints"`
	// PodPidsLimit is the maximum number of pids in any pod.
	PodPidsLimit *int64 `json:"podPidsLimit,omitempty" flag:"pod-max-pids"`
	// ExperimentalAllocatableIgnoreEviction enables ignoring Hard Eviction Thresholds while calculating Node Allocatable
	ExperimentalAllocatableIgnoreEviction *bool `json:"experimentalAllocatableIgnoreEviction,omitempty" flag:"experimental-allocatable-ignore-eviction"`

	// ShutdownGracePeriod specifies the total duration that the node should delay the shutdown by.
	// Default: 30s
	ShutdownGracePeriod *metav1.Duration `json:"shutdownGracePeriod,omitempty"`
	// ShutdownGracePeriodCriticalPods specifies the duration used to terminate critical pods during a node shutdown.
	// Default: 10s
	ShutdownGracePeriodCriticalPods *metav1.Duration `json:"shutdownGracePeriodCriticalPods,omitempty"`
	// MemorySwapBehavior defines how swap is used by container workloads.
	// Supported values: LimitedSwap, "UnlimitedSwap.
	MemorySwapBehavior string `json:"memorySwapBehavior,omitempty"`
}

// KubeProxyConfig defines the configuration for a proxy
type KubeProxyConfig struct {
	Image string `json:"image,omitempty"`
	// CPURequest, cpu request compute resource for kube proxy e.g. "20m"
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// CPULimit, cpu limit compute resource for kube proxy e.g. "30m"
	CPULimit *resource.Quantity `json:"cpuLimit,omitempty"`
	// MemoryRequest, memory request compute resource for kube proxy e.g. "30Mi"
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// MemoryLimit, memory limit compute resource for kube proxy e.g. "30Mi"
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
	// LogLevel is the logging level of the proxy
	LogLevel int32 `json:"logLevel,omitempty" flag:"v"`
	// ClusterCIDR is the CIDR range of the pods in the cluster
	ClusterCIDR *string `json:"clusterCIDR,omitempty" flag:"cluster-cidr"`
	// HostnameOverride, if non-empty, will be used as the identity instead of the actual hostname.
	HostnameOverride string `json:"hostnameOverride,omitempty" flag:"hostname-override"`
	// BindAddress is IP address for the proxy server to serve on
	BindAddress string `json:"bindAddress,omitempty" flag:"bind-address"`
	// Master is the address of the Kubernetes API server (overrides any value in kubeconfig)
	Master string `json:"master,omitempty" flag:"master"`
	// MetricsBindAddress is the IP address for the metrics server to serve on
	MetricsBindAddress *string `json:"metricsBindAddress,omitempty" flag:"metrics-bind-address"`
	// Enabled allows enabling or disabling kube-proxy
	Enabled *bool `json:"enabled,omitempty"`
	// Which proxy mode to use: (userspace, iptables(default), ipvs)
	ProxyMode string `json:"proxyMode,omitempty" flag:"proxy-mode"`
	// IPVSExcludeCIDRs is comma-separated list of CIDR's which the ipvs proxier should not touch when cleaning up IPVS rules
	IPVSExcludeCIDRs []string `json:"ipvsExcludeCIDRs,omitempty" flag:"ipvs-exclude-cidrs"`
	// IPVSMinSyncPeriod is the minimum interval of how often the ipvs rules can be refreshed as endpoints and services change (e.g. '5s', '1m', '2h22m')
	IPVSMinSyncPeriod *metav1.Duration `json:"ipvsMinSyncPeriod,omitempty" flag:"ipvs-min-sync-period"`
	// IPVSScheduler is the ipvs scheduler type when proxy mode is ipvs
	IPVSScheduler *string `json:"ipvsScheduler,omitempty" flag:"ipvs-scheduler"`
	// IPVSSyncPeriod duration is the maximum interval of how often ipvs rules are refreshed
	IPVSSyncPeriod *metav1.Duration `json:"ipvsSyncPeriod,omitempty" flag:"ipvs-sync-period"`
	// FeatureGates is a series of key pairs used to switch on features for the proxy
	FeatureGates map[string]string `json:"featureGates,omitempty" flag:"feature-gates"`
	// Maximum number of NAT connections to track per CPU core (default: 131072)
	ConntrackMaxPerCore *int32 `json:"conntrackMaxPerCore,omitempty" flag:"conntrack-max-per-core"`
	// Minimum number of conntrack entries to allocate, regardless of conntrack-max-per-core
	ConntrackMin *int32 `json:"conntrackMin,omitempty" flag:"conntrack-min"`
}

// KubeAPIServerConfig defines the configuration for the kube api
type KubeAPIServerConfig struct {
	// Image is the container image used.
	Image string `json:"image,omitempty"`
	// DisableBasicAuth removes the --basic-auth-file flag
	DisableBasicAuth *bool `json:"disableBasicAuth,omitempty"`
	// LogFormat is the logging format of the api.
	// Supported values: text, json.
	// Default: text
	LogFormat string `json:"logFormat,omitempty" flag:"logging-format" flag-empty:"text"`
	// LogLevel is the logging level of the api
	LogLevel int32 `json:"logLevel,omitempty" flag:"v" flag-empty:"0"`
	// CloudProvider is the name of the cloudProvider we are using, aws, gce etcd
	CloudProvider string `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	// SecurePort is the port the kube runs on
	SecurePort int32 `json:"securePort,omitempty" flag:"secure-port"`
	// InsecurePort is the port the insecure api runs
	InsecurePort *int32 `json:"insecurePort,omitempty" flag:"insecure-port"`
	// Address is the binding address for the kube api: Deprecated - use insecure-bind-address and bind-address
	Address string `json:"address,omitempty" flag:"address"`
	// AdvertiseAddress is the IP address on which to advertise the apiserver to members of the cluster.
	AdvertiseAddress string `json:"advertiseAddress,omitempty" flag:"advertise-address"`
	// BindAddress is the binding address for the secure kubernetes API
	BindAddress string `json:"bindAddress,omitempty" flag:"bind-address"`
	// InsecureBindAddress is the binding address for the InsecurePort for the insecure kubernetes API
	InsecureBindAddress string `json:"insecureBindAddress,omitempty" flag:"insecure-bind-address"`
	// EnableBootstrapAuthToken enables 'bootstrap.kubernetes.io/token' in the 'kube-system' namespace to be used for TLS bootstrapping authentication
	EnableBootstrapAuthToken *bool `json:"enableBootstrapTokenAuth,omitempty" flag:"enable-bootstrap-token-auth"`
	// EnableAggregatorRouting enables aggregator routing requests to endpoints IP rather than cluster IP
	EnableAggregatorRouting *bool `json:"enableAggregatorRouting,omitempty" flag:"enable-aggregator-routing"`
	// AdmissionControl is a list of admission controllers to use: Deprecated - use enable-admission-plugins instead
	AdmissionControl []string `json:"admissionControl,omitempty" flag:"admission-control"`
	// AppendAdmissionPlugins appends list of enabled admission plugins
	AppendAdmissionPlugins []string `json:"appendAdmissionPlugins,omitempty"`
	// EnableAdmissionPlugins is a list of enabled admission plugins
	EnableAdmissionPlugins []string `json:"enableAdmissionPlugins,omitempty" flag:"enable-admission-plugins"`
	// DisableAdmissionPlugins is a list of disabled admission plugins
	DisableAdmissionPlugins []string `json:"disableAdmissionPlugins,omitempty" flag:"disable-admission-plugins"`
	// AdmissionControlConfigFile is the location of the admission-control-config-file
	AdmissionControlConfigFile string `json:"admissionControlConfigFile,omitempty" flag:"admission-control-config-file"`
	// ServiceClusterIPRange is the service address range
	ServiceClusterIPRange string `json:"serviceClusterIPRange,omitempty" flag:"service-cluster-ip-range"`
	// Passed as --service-node-port-range to kube-apiserver. Expects 'startPort-endPort' format e.g. 30000-33000
	ServiceNodePortRange string `json:"serviceNodePortRange,omitempty" flag:"service-node-port-range"`
	// EtcdServers is a list of the etcd service to connect
	EtcdServers []string `json:"etcdServers,omitempty" flag:"etcd-servers"`
	// EtcdServersOverrides is per-resource etcd servers overrides, comma separated. The individual override format: group/resource#servers, where servers are http://ip:port, semicolon separated
	EtcdServersOverrides []string `json:"etcdServersOverrides,omitempty" flag:"etcd-servers-overrides"`
	// EtcdCAFile is the path to a ca certificate
	EtcdCAFile string `json:"etcdCAFile,omitempty" flag:"etcd-cafile"`
	// EtcdCertFile is the path to a certificate
	EtcdCertFile string `json:"etcdCertFile,omitempty" flag:"etcd-certfile"`
	// EtcdKeyFile is the path to a private key
	EtcdKeyFile string `json:"etcdKeyFile,omitempty" flag:"etcd-keyfile"`
	// TODO: Remove unused BasicAuthFile
	BasicAuthFile string `json:"basicAuthFile,omitempty" flag:"basic-auth-file"`
	// ClientCAFile is the file used by apisever that contains the client CA
	ClientCAFile string `json:"clientCAFile,omitempty" flag:"client-ca-file"`
	// TODO: Remove unused TLSCertFile
	TLSCertFile string `json:"tlsCertFile,omitempty" flag:"tls-cert-file"`
	// TODO: Remove unused TLSPrivateKeyFile
	TLSPrivateKeyFile string `json:"tlsPrivateKeyFile,omitempty" flag:"tls-private-key-file"`
	// TLSCipherSuites indicates the allowed TLS cipher suite
	TLSCipherSuites []string `json:"tlsCipherSuites,omitempty" flag:"tls-cipher-suites"`
	// TLSMinVersion indicates the minimum TLS version allowed
	TLSMinVersion string `json:"tlsMinVersion,omitempty" flag:"tls-min-version"`
	// TODO: Remove unused TokenAuthFile
	TokenAuthFile string `json:"tokenAuthFile,omitempty" flag:"token-auth-file"`
	// AllowPrivileged indicates if we can run privileged containers
	AllowPrivileged *bool `json:"allowPrivileged,omitempty" flag:"allow-privileged"`
	// APIServerCount is the number of api servers
	APIServerCount *int32 `json:"apiServerCount,omitempty" flag:"apiserver-count"`
	// RuntimeConfig is a series of keys/values are parsed into the `--runtime-config` parameters
	RuntimeConfig map[string]string `json:"runtimeConfig,omitempty" flag:"runtime-config"`
	// KubeletClientCertificate is the path of a certificate for secure communication between api and kubelet
	KubeletClientCertificate string `json:"kubeletClientCertificate,omitempty" flag:"kubelet-client-certificate"`
	// KubeletCertificateAuthority is the path of a certificate authority for secure communication between api and kubelet.
	KubeletCertificateAuthority string `json:"kubeletCertificateAuthority,omitempty" flag:"kubelet-certificate-authority"`
	// KubeletClientKey is the path of a private to secure communication between api and kubelet
	KubeletClientKey string `json:"kubeletClientKey,omitempty" flag:"kubelet-client-key"`
	// AnonymousAuth indicates if anonymous authentication is permitted
	AnonymousAuth *bool `json:"anonymousAuth,omitempty" flag:"anonymous-auth"`
	// KubeletPreferredAddressTypes is a list of the preferred NodeAddressTypes to use for kubelet connections
	KubeletPreferredAddressTypes []string `json:"kubeletPreferredAddressTypes,omitempty" flag:"kubelet-preferred-address-types"`
	// StorageBackend is the backend storage
	StorageBackend *string `json:"storageBackend,omitempty" flag:"storage-backend"`
	// OIDCUsernameClaim is the OpenID claim to use as the user name.
	// Note that claims other than the default ('sub') is not guaranteed to be
	// unique and immutable.
	OIDCUsernameClaim *string `json:"-" flag:"oidc-username-claim"`
	// OIDCUsernamePrefix is the prefix prepended to username claims to prevent
	// clashes with existing names (such as 'system:' users).
	OIDCUsernamePrefix *string `json:"-" flag:"oidc-username-prefix"`
	// OIDCGroupsClaim if provided, the name of a custom OpenID Connect claim for
	// specifying user groups.
	// The claim value is expected to be a string or array of strings.
	OIDCGroupsClaim *string `json:"-" flag:"oidc-groups-claim"`
	// OIDCGroupsPrefix is the prefix prepended to group claims to prevent
	// clashes with existing names (such as 'system:' groups)
	OIDCGroupsPrefix *string `json:"-" flag:"oidc-groups-prefix"`
	// OIDCIssuerURL is the URL of the OpenID issuer, only HTTPS scheme will
	// be accepted.
	// If set, it will be used to verify the OIDC JSON Web Token (JWT).
	OIDCIssuerURL *string `json:"-" flag:"oidc-issuer-url"`
	// OIDCClientID is the client ID for the OpenID Connect client, must be set
	// if oidc-issuer-url is set.
	OIDCClientID *string `json:"-" flag:"oidc-client-id"`
	// A key=value pair that describes a required claim in the ID Token.
	// If set, the claim is verified to be present in the ID Token with a matching value.
	// Repeat this flag to specify multiple claims.
	OIDCRequiredClaim []string `json:"-" flag:"oidc-required-claim,repeat"`
	// OIDCCAFile if set, the OpenID server's certificate will be verified by one
	// of the authorities in the oidc-ca-file
	OIDCCAFile *string `json:"oidcCAFile,omitempty" flag:"oidc-ca-file"`
	// AuthenticationConfigFile is the location of the authentication-config
	// this option is mutually exclusive with all OIDC options
	AuthenticationConfigFile string `json:"authenticationConfigFile,omitempty" flag:"authentication-config"`
	// The apiserver's client certificate used for outbound requests.
	ProxyClientCertFile *string `json:"proxyClientCertFile,omitempty" flag:"proxy-client-cert-file"`
	// The apiserver's client key used for outbound requests.
	ProxyClientKeyFile *string `json:"proxyClientKeyFile,omitempty" flag:"proxy-client-key-file"`
	// AuditLogFormat flag specifies the format type for audit log files.
	AuditLogFormat *string `json:"auditLogFormat,omitempty" flag:"audit-log-format"`
	// If set, all requests coming to the apiserver will be logged to this file.
	AuditLogPath *string `json:"auditLogPath,omitempty" flag:"audit-log-path"`
	// The maximum number of days to retain old audit log files based on the timestamp encoded in their filename.
	AuditLogMaxAge *int32 `json:"auditLogMaxAge,omitempty" flag:"audit-log-maxage"`
	// The maximum number of old audit log files to retain.
	AuditLogMaxBackups *int32 `json:"auditLogMaxBackups,omitempty" flag:"audit-log-maxbackup"`
	// The maximum size in megabytes of the audit log file before it gets rotated. Defaults to 100MB.
	AuditLogMaxSize *int32 `json:"auditLogMaxSize,omitempty" flag:"audit-log-maxsize"`
	// AuditPolicyFile is the full path to a advanced audit configuration file e.g. /srv/kubernetes/audit.conf
	AuditPolicyFile string `json:"auditPolicyFile,omitempty" flag:"audit-policy-file"`
	// AuditWebhookBatchBufferSize is The size of the buffer to store events before batching and writing. Only used in batch mode. (default 10000)
	AuditWebhookBatchBufferSize *int32 `json:"auditWebhookBatchBufferSize,omitempty" flag:"audit-webhook-batch-buffer-size"`
	// AuditWebhookBatchMaxSize is The maximum size of a batch. Only used in batch mode. (default 400)
	AuditWebhookBatchMaxSize *int32 `json:"auditWebhookBatchMaxSize,omitempty" flag:"audit-webhook-batch-max-size"`
	// AuditWebhookBatchMaxWait is The amount of time to wait before force writing the batch that hadn't reached the max size. Only used in batch mode. (default 30s)
	AuditWebhookBatchMaxWait *metav1.Duration `json:"auditWebhookBatchMaxWait,omitempty" flag:"audit-webhook-batch-max-wait"`
	// AuditWebhookBatchThrottleBurst is Maximum number of requests sent at the same moment if ThrottleQPS was not utilized before. Only used in batch mode. (default 15)
	AuditWebhookBatchThrottleBurst *int32 `json:"auditWebhookBatchThrottleBurst,omitempty" flag:"audit-webhook-batch-throttle-burst"`
	// AuditWebhookBatchThrottleEnable is Whether batching throttling is enabled. Only used in batch mode. (default true)
	AuditWebhookBatchThrottleEnable *bool `json:"auditWebhookBatchThrottleEnable,omitempty" flag:"audit-webhook-batch-throttle-enable"`
	// AuditWebhookBatchThrottleQps is Maximum average number of batches per second. Only used in batch mode. (default 10)
	AuditWebhookBatchThrottleQps *resource.Quantity `json:"auditWebhookBatchThrottleQps,omitempty" flag:"audit-webhook-batch-throttle-qps"`
	// AuditWebhookConfigFile is Path to a kubeconfig formatted file that defines the audit webhook configuration. Requires the 'AdvancedAuditing' feature gate.
	AuditWebhookConfigFile string `json:"auditWebhookConfigFile,omitempty" flag:"audit-webhook-config-file"`
	// AuditWebhookInitialBackoff is The amount of time to wait before retrying the first failed request. (default 10s)
	AuditWebhookInitialBackoff *metav1.Duration `json:"auditWebhookInitialBackoff,omitempty" flag:"audit-webhook-initial-backoff"`
	// AuditWebhookMode is Strategy for sending audit events. Blocking indicates sending events should block server responses. Batch causes the backend to buffer and write events asynchronously. Known modes are batch,blocking. (default "batch")
	AuditWebhookMode string `json:"auditWebhookMode,omitempty" flag:"audit-webhook-mode"`
	// File with webhook configuration for token authentication in kubeconfig format. The API server will query the remote service to determine authentication for bearer tokens.
	AuthenticationTokenWebhookConfigFile *string `json:"authenticationTokenWebhookConfigFile,omitempty" flag:"authentication-token-webhook-config-file"`
	// The duration to cache responses from the webhook token authenticator. Default is 2m. (default 2m0s)
	AuthenticationTokenWebhookCacheTTL *metav1.Duration `json:"authenticationTokenWebhookCacheTtl,omitempty" flag:"authentication-token-webhook-cache-ttl"`
	// AuthorizationMode is the authorization mode the kubeapi is running in
	AuthorizationMode *string `json:"authorizationMode,omitempty" flag:"authorization-mode"`
	// File with webhook configuration for authorization in kubeconfig format. The API server will query the remote service to determine whether to authorize the request.
	AuthorizationWebhookConfigFile *string `json:"authorizationWebhookConfigFile,omitempty" flag:"authorization-webhook-config-file"`
	// The duration to cache authorized responses from the webhook token authorizer. Default is 5m. (default 5m0s)
	AuthorizationWebhookCacheAuthorizedTTL *metav1.Duration `json:"authorizationWebhookCacheAuthorizedTTL,omitempty" flag:"authorization-webhook-cache-authorized-ttl"`
	// The duration to cache authorized responses from the webhook token authorizer. Default is 30s. (default 30s)
	AuthorizationWebhookCacheUnauthorizedTTL *metav1.Duration `json:"authorizationWebhookCacheUnauthorizedTTL,omitempty" flag:"authorization-webhook-cache-unauthorized-ttl"`
	// AuthorizationRBACSuperUser is the name of the superuser for default rbac
	AuthorizationRBACSuperUser *string `json:"authorizationRBACSuperUser,omitempty" flag:"authorization-rbac-super-user"`
	// EncryptionProviderConfig enables encryption at rest for secrets.
	EncryptionProviderConfig *string `json:"encryptionProviderConfig,omitempty" flag:"encryption-provider-config"`
	// ExperimentalEncryptionProviderConfig enables encryption at rest for secrets.
	ExperimentalEncryptionProviderConfig *string `json:"experimentalEncryptionProviderConfig,omitempty" flag:"experimental-encryption-provider-config"`

	// List of request headers to inspect for usernames. X-Remote-User is common.
	RequestheaderUsernameHeaders []string `json:"requestheaderUsernameHeaders,omitempty" flag:"requestheader-username-headers"`
	// List of request headers to inspect for groups. X-Remote-Group is suggested.
	RequestheaderGroupHeaders []string `json:"requestheaderGroupHeaders,omitempty" flag:"requestheader-group-headers"`
	// List of request header prefixes to inspect. X-Remote-Extra- is suggested.
	RequestheaderExtraHeaderPrefixes []string `json:"requestheaderExtraHeaderPrefixes,omitempty" flag:"requestheader-extra-headers-prefix"`
	// Root certificate bundle to use to verify client certificates on incoming requests before trusting usernames in headers specified by --requestheader-username-headers
	RequestheaderClientCAFile string `json:"requestheaderClientCAFile,omitempty" flag:"requestheader-client-ca-file"`
	// List of client certificate common names to allow to provide usernames in headers specified by --requestheader-username-headers. If empty, any client certificate validated by the authorities in --requestheader-client-ca-file is allowed.
	RequestheaderAllowedNames []string `json:"requestheaderAllowedNames,omitempty" flag:"requestheader-allowed-names"`
	// FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
	FeatureGates map[string]string `json:"featureGates,omitempty" flag:"feature-gates"`
	// GoawayChance is the probability that send a GOAWAY to HTTP/2 clients. Default to 0, means never send GOAWAY. Max is 0.02 to prevent break the apiserver.
	GoawayChance string `json:"goawayChance,omitempty" flag:"goaway-chance"`
	// MaxRequestsInflight The maximum number of non-mutating requests in flight at a given time.
	MaxRequestsInflight int32 `json:"maxRequestsInflight,omitempty" flag:"max-requests-inflight" flag-empty:"0"`
	// MaxMutatingRequestsInflight The maximum number of mutating requests in flight at a given time. Defaults to 200
	MaxMutatingRequestsInflight int32 `json:"maxMutatingRequestsInflight,omitempty" flag:"max-mutating-requests-inflight" flag-empty:"0"`

	// HTTP2MaxStreamsPerConnection sets the limit that the server gives to clients for the maximum number of streams in an HTTP/2 connection. Zero means to use golang's default.
	HTTP2MaxStreamsPerConnection *int32 `json:"http2MaxStreamsPerConnection,omitempty" flag:"http2-max-streams-per-connection"`

	// EtcdQuorumRead configures the etcd-quorum-read flag, which forces consistent reads from etcd
	EtcdQuorumRead *bool `json:"etcdQuorumRead,omitempty" flag:"etcd-quorum-read"`

	// RequestTimeout configures the duration a handler must keep a request open before timing it out. (default 1m0s)
	RequestTimeout *metav1.Duration `json:"requestTimeout,omitempty" flag:"request-timeout"`

	// MinRequestTimeout configures the minimum number of seconds a handler must keep a request open before timing it out.
	// Currently only honored by the watch request handler
	MinRequestTimeout *int32 `json:"minRequestTimeout,omitempty" flag:"min-request-timeout"`

	// Used to disable watch caching in the apiserver, defaults to enabling caching by omission
	WatchCache *bool `json:"watchCache,omitempty" flag:"watch-cache"`

	// Set the watch-cache-sizes parameter for the apiserver
	// The only meaningful value is setting to 0, which disable caches for specific object types.
	// Setting any values other than 0 for a resource will yield no effect since the caches are dynamic
	WatchCacheSizes []string `json:"watchCacheSizes,omitempty" flag:"watch-cache-sizes" flag-empty:"0"`

	// File containing PEM-encoded x509 RSA or ECDSA private or public keys, used to verify ServiceAccount tokens.
	// The specified file can contain multiple keys, and the flag can be specified multiple times with different files.
	// If unspecified, --tls-private-key-file is used.
	ServiceAccountKeyFile []string `json:"serviceAccountKeyFile,omitempty" flag:"service-account-key-file,repeat"`

	// Path to the file that contains the current private key of the service account token issuer.
	// The issuer will sign issued ID tokens with this private key. (Requires the 'TokenRequest' feature gate.)
	ServiceAccountSigningKeyFile *string `json:"serviceAccountSigningKeyFile,omitempty" flag:"service-account-signing-key-file"`

	// Identifier of the service account token issuer. The issuer will assert this identifier
	// in "iss" claim of issued tokens. This value is a string or URI.
	ServiceAccountIssuer *string `json:"serviceAccountIssuer,omitempty" flag:"service-account-issuer"`

	// AdditionalServiceAccountIssuers can contain additional service account token issuers.
	AdditionalServiceAccountIssuers []string `json:"additionalServiceAccountIssuers,omitempty"`

	// ServiceAccountJWKSURI overrides the path for the jwks document; this is useful when we are republishing the service account discovery information elsewhere.
	ServiceAccountJWKSURI *string `json:"serviceAccountJWKSURI,omitempty" flag:"service-account-jwks-uri"`

	// Identifiers of the API. The service account token authenticator will validate that
	// tokens used against the API are bound to at least one of these audiences. If the
	// --service-account-issuer flag is configured and this flag is not, this field
	// defaults to a single element list containing the issuer URL.
	APIAudiences []string `json:"apiAudiences,omitempty" flag:"api-audiences"`

	// CPURequest, cpu request compute resource for api server. Defaults to "150m"
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// CPULimit, cpu limit compute resource for api server e.g. "500m"
	CPULimit *resource.Quantity `json:"cpuLimit,omitempty"`
	// MemoryRequest, memory request compute resource for api server e.g. "30Mi"
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// MemoryLimit, memory limit compute resource for api server e.g. "30Mi"
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`

	// Amount of time to retain Kubernetes events
	EventTTL *metav1.Duration `json:"eventTTL,omitempty" flag:"event-ttl"`

	// AuditDynamicConfiguration enables dynamic audit configuration via AuditSinks
	AuditDynamicConfiguration *bool `json:"auditDynamicConfiguration,omitempty" flag:"audit-dynamic-configuration"`

	// EnableProfiling enables profiling via web interface host:port/debug/pprof/
	EnableProfiling *bool `json:"enableProfiling,omitempty" flag:"profiling"`
	// EnableContentionProfiling enables block profiling, if profiling is enabled
	EnableContentionProfiling *bool `json:"enableContentionProfiling,omitempty" flag:"contention-profiling"`

	// CorsAllowedOrigins is a list of origins for CORS. An allowed origin can be a regular
	// expression to support subdomain matching. If this list is empty CORS will not be enabled.
	CorsAllowedOrigins []string `json:"corsAllowedOrigins,omitempty" flag:"cors-allowed-origins"`

	// DefaultNotReadyTolerationSeconds indicates the tolerationSeconds of the toleration for notReady:NoExecute that is added by default to every pod that does not already have such a toleration.
	DefaultNotReadyTolerationSeconds *int64 `json:"defaultNotReadyTolerationSeconds,omitempty" flag:"default-not-ready-toleration-seconds"`
	// DefaultUnreachableTolerationSeconds indicates the tolerationSeconds of the toleration for unreachable:NoExecute that is added by default to every pod that does not already have such a toleration.
	DefaultUnreachableTolerationSeconds *int64 `json:"defaultUnreachableTolerationSeconds,omitempty" flag:"default-unreachable-toleration-seconds"`

	// Env allows users to pass in env variables to the apiserver container.
	// This can be useful to control some environment runtime settings, such as GOMEMLIMIT and GOCG to tweak the memory settings of the apiserver
	// This also allows the flexibility for adding any other variables for future use cases
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// KubeControllerManagerConfig is the configuration for the controller
type KubeControllerManagerConfig struct {
	// Master is the url for the kube api master
	Master string `json:"master,omitempty" flag:"master"`
	// LogFormat is the logging format of the controler manager.
	// Supported values: text, json.
	// Default: text
	LogFormat string `json:"logFormat,omitempty" flag:"logging-format" flag-empty:"text"`
	// LogLevel is the defined logLevel
	LogLevel int32 `json:"logLevel,omitempty" flag:"v" flag-empty:"0"`
	// ServiceAccountPrivateKeyFile is the location of the private key for service account token signing.
	ServiceAccountPrivateKeyFile string `json:"serviceAccountPrivateKeyFile,omitempty" flag:"service-account-private-key-file"`
	// Image is the container image to use.
	Image string `json:"image,omitempty"`
	// CloudProvider is the provider for cloud services.
	CloudProvider string `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	// ClusterName is the instance prefix for the cluster.
	ClusterName string `json:"clusterName,omitempty" flag:"cluster-name"`
	// ClusterCIDR is CIDR Range for Pods in cluster.
	ClusterCIDR string `json:"clusterCIDR,omitempty" flag:"cluster-cidr"`
	// AllocateNodeCIDRs enables CIDRs for Pods to be allocated and, if ConfigureCloudRoutes is true, to be set on the cloud provider.
	AllocateNodeCIDRs *bool `json:"allocateNodeCIDRs,omitempty" flag:"allocate-node-cidrs"`
	// NodeCIDRMaskSize set the size for the mask of the nodes.
	NodeCIDRMaskSize *int32 `json:"nodeCIDRMaskSize,omitempty" flag:"node-cidr-mask-size"`
	// ConfigureCloudRoutes enables CIDRs allocated with to be configured on the cloud provider.
	ConfigureCloudRoutes *bool `json:"configureCloudRoutes,omitempty" flag:"configure-cloud-routes"`
	// Controllers is a list of controllers to enable on the controller-manager
	Controllers []string `json:"controllers,omitempty" flag:"controllers"`
	// CIDRAllocatorType specifies the type of CIDR allocator to use.
	CIDRAllocatorType *string `json:"cidrAllocatorType,omitempty" flag:"cidr-allocator-type"`
	// rootCAFile is the root certificate authority will be included in service account's token secret. This must be a valid PEM-encoded CA bundle.
	RootCAFile string `json:"rootCAFile,omitempty" flag:"root-ca-file"`
	// LeaderElection defines the configuration of leader election client.
	LeaderElection *LeaderElectionConfiguration `json:"leaderElection,omitempty"`
	// AttachDetachReconcileSyncPeriod is the amount of time the reconciler sync states loop
	// wait between successive executions. Is set to 1 min by kops by default
	AttachDetachReconcileSyncPeriod *metav1.Duration `json:"attachDetachReconcileSyncPeriod,omitempty" flag:"attach-detach-reconcile-sync-period"`
	// DisableAttachDetachReconcileSync disables the reconcile sync loop in the attach-detach controller.
	// This can cause volumes to become mismatched with pods
	DisableAttachDetachReconcileSync *bool `json:"disableAttachDetachReconcileSync,omitempty" flag:"disable-attach-detach-reconcile-sync"`
	// TerminatedPodGCThreshold is the number of terminated pods that can exist
	// before the terminated pod garbage collector starts deleting terminated pods.
	// If <= 0, the terminated pod garbage collector is disabled.
	TerminatedPodGCThreshold *int32 `json:"terminatedPodGCThreshold,omitempty" flag:"terminated-pod-gc-threshold"`
	// NodeMonitorPeriod is the period for syncing NodeStatus in NodeController. (default 5s)
	NodeMonitorPeriod *metav1.Duration `json:"nodeMonitorPeriod,omitempty" flag:"node-monitor-period"`
	// NodeMonitorGracePeriod is the amount of time which we allow running Node to be unresponsive before marking it unhealthy. (default 40s)
	// Must be N-1 times more than kubelet's nodeStatusUpdateFrequency, where N means number of retries allowed for kubelet to post node status.
	NodeMonitorGracePeriod *metav1.Duration `json:"nodeMonitorGracePeriod,omitempty" flag:"node-monitor-grace-period"`
	// PodEvictionTimeout is the grace period for deleting pods on failed nodes. (default 5m0s)
	PodEvictionTimeout *metav1.Duration `json:"podEvictionTimeout,omitempty" flag:"pod-eviction-timeout"`
	// UseServiceAccountCredentials controls whether we use individual service account credentials for each controller.
	UseServiceAccountCredentials *bool `json:"useServiceAccountCredentials,omitempty" flag:"use-service-account-credentials"`
	// HorizontalPodAutoscalerSyncPeriod is the amount of time between syncs
	// During each period, the controller manager queries the resource utilization
	// against the metrics specified in each HorizontalPodAutoscaler definition.
	HorizontalPodAutoscalerSyncPeriod *metav1.Duration `json:"horizontalPodAutoscalerSyncPeriod,omitempty" flag:"horizontal-pod-autoscaler-sync-period"`
	// HorizontalPodAutoscalerDownscaleDelay is a duration that specifies
	// how long the autoscaler has to wait before another downscale
	// operation can be performed after the current one has completed.
	HorizontalPodAutoscalerDownscaleDelay *metav1.Duration `json:"horizontalPodAutoscalerDownscaleDelay,omitempty" flag:"horizontal-pod-autoscaler-downscale-delay"`
	// HorizontalPodAutoscalerDownscaleStabilization is the period for which
	// autoscaler will look backwards and not scale down below any
	// recommendation it made during that period.
	HorizontalPodAutoscalerDownscaleStabilization *metav1.Duration `json:"horizontalPodAutoscalerDownscaleStabilization,omitempty" flag:"horizontal-pod-autoscaler-downscale-stabilization"`
	// HorizontalPodAutoscalerUpscaleDelay is a duration that specifies how
	// long the autoscaler has to wait before another upscale operation can
	// be performed after the current one has completed.
	HorizontalPodAutoscalerUpscaleDelay *metav1.Duration `json:"horizontalPodAutoscalerUpscaleDelay,omitempty" flag:"horizontal-pod-autoscaler-upscale-delay"`
	// HorizontalPodAutoscalerInitialReadinessDelay is the period after pod start
	// during which readiness changes will be treated as initial readiness. (default 30s)
	HorizontalPodAutoscalerInitialReadinessDelay *metav1.Duration `json:"horizontalPodAutoscalerInitialReadinessDelay,omitempty" flag:"horizontal-pod-autoscaler-initial-readiness-delay"`
	// HorizontalPodAutoscalerCPUInitializationPeriod is the period after pod start
	// when CPU samples might be skipped. (default 5m)
	HorizontalPodAutoscalerCPUInitializationPeriod *metav1.Duration `json:"horizontalPodAutoscalerCpuInitializationPeriod,omitempty" flag:"horizontal-pod-autoscaler-cpu-initialization-period"`
	// HorizontalPodAutoscalerTolerance is the minimum change (from 1.0) in the
	// desired-to-actual metrics ratio for the horizontal pod autoscaler to
	// consider scaling.
	HorizontalPodAutoscalerTolerance *resource.Quantity `json:"horizontalPodAutoscalerTolerance,omitempty" flag:"horizontal-pod-autoscaler-tolerance"`
	// HorizontalPodAutoscalerUseRestClients determines if the new-style clients
	// should be used if support for custom metrics is enabled.
	HorizontalPodAutoscalerUseRestClients *bool `json:"horizontalPodAutoscalerUseRestClients,omitempty" flag:"horizontal-pod-autoscaler-use-rest-clients"`
	// ExperimentalClusterSigningDuration is the max length of duration that the signed certificates will be given. (default 365*24h)
	// Deprecated - use cluster-signing-duration instead
	ExperimentalClusterSigningDuration *metav1.Duration `json:"experimentalClusterSigningDuration,omitempty" flag:"experimental-cluster-signing-duration"`
	// ClusterSigningDuration is the max length of duration that the signed certificates will be given. (default 365*24h)
	ClusterSigningDuration *metav1.Duration `json:"ClusterSigningDuration,omitempty" flag:"cluster-signing-duration"`
	// FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
	FeatureGates map[string]string `json:"featureGates,omitempty" flag:"feature-gates"`
	// TLSCertFile is the file containing the TLS server certificate.
	TLSCertFile *string `json:"tlsCertFile,omitempty" flag:"tls-cert-file"`
	// TLSCipherSuites indicates the allowed TLS cipher suite
	TLSCipherSuites []string `json:"tlsCipherSuites,omitempty" flag:"tls-cipher-suites"`
	// TLSMinVersion indicates the minimum TLS version allowed
	TLSMinVersion string `json:"tlsMinVersion,omitempty" flag:"tls-min-version"`
	// TLSPrivateKeyFile is the file containing the private key for the TLS server certificate.
	TLSPrivateKeyFile string `json:"tlsPrivateKeyFile,omitempty" flag:"tls-private-key-file"`
	// MinResyncPeriod indicates the resync period in reflectors.
	// The resync period will be random between MinResyncPeriod and 2*MinResyncPeriod. (default 12h0m0s)
	MinResyncPeriod string `json:"minResyncPeriod,omitempty" flag:"min-resync-period"`
	// KubeAPIQPS QPS to use while talking with kubernetes apiserver. (default 20)
	KubeAPIQPS *resource.Quantity `json:"kubeAPIQPS,omitempty" flag:"kube-api-qps"`
	// KubeAPIBurst Burst to use while talking with kubernetes apiserver. (default 30)
	KubeAPIBurst *int32 `json:"kubeAPIBurst,omitempty" flag:"kube-api-burst"`
	// The number of deployment objects that are allowed to sync concurrently.
	ConcurrentDeploymentSyncs *int32 `json:"concurrentDeploymentSyncs,omitempty" flag:"concurrent-deployment-syncs"`
	// The number of endpoint objects that are allowed to sync concurrently.
	ConcurrentEndpointSyncs *int32 `json:"concurrentEndpointSyncs,omitempty" flag:"concurrent-endpoint-syncs"`
	// The number of namespace objects that are allowed to sync concurrently.
	ConcurrentNamespaceSyncs *int32 `json:"concurrentNamespaceSyncs,omitempty" flag:"concurrent-namespace-syncs"`
	// The number of replicaset objects that are allowed to sync concurrently.
	ConcurrentReplicasetSyncs *int32 `json:"concurrentReplicasetSyncs,omitempty" flag:"concurrent-replicaset-syncs"`
	// The number of service objects that are allowed to sync concurrently.
	ConcurrentServiceSyncs *int32 `json:"concurrentServiceSyncs,omitempty" flag:"concurrent-service-syncs"`
	// The number of resourcequota objects that are allowed to sync concurrently.
	ConcurrentResourceQuotaSyncs *int32 `json:"concurrentResourceQuotaSyncs,omitempty" flag:"concurrent-resource-quota-syncs"`
	// The number of serviceaccount objects that are allowed to sync concurrently to create tokens.
	ConcurrentServiceaccountTokenSyncs *int32 `json:"concurrentServiceaccountTokenSyncs,omitempty" flag:"concurrent-serviceaccount-token-syncs"`
	// The number of replicationcontroller objects that are allowed to sync concurrently.
	ConcurrentRCSyncs *int32 `json:"concurrentRCSyncs,omitempty" flag:"concurrent-rc-syncs"`
	// The number of horizontal pod autoscaler objects that are allowed to sync concurrently (default 5).
	ConcurrentHorizontalPodAustoscalerSyncs *int32 `json:"concurrentHorizontalPodAustoscalerSyncs,omitempty" flag:"concurrent-horizontal-pod-autoscaler-syncs"`
	// The number of job objects that are allowed to sync concurrently (default 5).
	ConcurrentJobSyncs *int32 `json:"concurrentJobSyncs,omitempty" flag:"concurrent-job-syncs"`
	// AuthenticationKubeconfig is the path to an Authentication Kubeconfig
	AuthenticationKubeconfig string `json:"authenticationKubeconfig,omitempty" flag:"authentication-kubeconfig"`
	// AuthorizationKubeconfig is the path to an Authorization Kubeconfig
	AuthorizationKubeconfig string `json:"authorizationKubeconfig,omitempty" flag:"authorization-kubeconfig"`
	// AuthorizationAlwaysAllowPaths is the list of HTTP paths to skip during authorization
	AuthorizationAlwaysAllowPaths []string `json:"authorizationAlwaysAllowPaths,omitempty" flag:"authorization-always-allow-paths"`
	// ExternalCloudVolumePlugin is a fallback mechanism that allows a legacy, in-tree cloudprovider to be used for volume plugins
	// even when an external cloud controller manager is being used.  This can be used instead of installing CSI.  The value should
	// be the same as is used for the --cloud-provider flag, i.e. "aws".
	ExternalCloudVolumePlugin string `json:"externalCloudVolumePlugin,omitempty" flag:"external-cloud-volume-plugin"`
	// The length of endpoint updates batching period. Processing of pod changes will be delayed by this duration
	// to join them with potential upcoming updates and reduce the overall number of endpoints updates.
	// Larger number = higher endpoint programming latency, but lower number of endpoints revision generated
	EndpointUpdatesBatchPeriod *metav1.Duration `json:"endpointUpdatesBatchPeriod,omitempty" flag:"endpoint-updates-batch-period"`
	// The length of endpoint slice updates batching period. Processing of pod changes will be delayed by this duration
	// to join them with potential upcoming updates and reduce the overall number of endpoints updates.
	// Larger number = higher endpoint programming latency, but lower number of endpoints revision generated.
	EndpointSliceUpdatesBatchPeriod *metav1.Duration `json:"endpointSliceUpdatesBatchPeriod,omitempty" flag:"endpointslice-updates-batch-period"`

	// EnableProfiling enables profiling via web interface host:port/debug/pprof/
	EnableProfiling *bool `json:"enableProfiling,omitempty" flag:"profiling"`
	// EnableContentionProfiling enables block profiling, if profiling is enabled
	EnableContentionProfiling *bool `json:"enableContentionProfiling,omitempty" flag:"contention-profiling"`
	// EnableLeaderMigration enables controller leader migration.
	EnableLeaderMigration *bool `json:"enableLeaderMigration,omitempty" flag:"enable-leader-migration"`

	// CPURequest, cpu request compute resource for kube-controler-manager. Defaults to "100m"
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// CPULimit, cpu limit compute resource for kube-controler-manager e.g. "500m"
	CPULimit *resource.Quantity `json:"cpuLimit,omitempty"`
	// MemoryRequest, memory request compute resource for kube-controler-manager e.g. "30Mi"
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// MemoryLimit, memory limit compute resource for kube-controler-manager e.g. "30Mi"
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
}

// CloudControllerManagerConfig is the configuration of the cloud controller
type CloudControllerManagerConfig struct {
	// Master is the url for the kube api master.
	Master string `json:"master,omitempty" flag:"master"`
	// LogLevel is the verbosity of the logs.
	LogLevel int32 `json:"logLevel,omitempty" flag:"v" flag-empty:"0"`
	// Image is the OCI image of the cloud controller manager.
	Image string `json:"image,omitempty"`
	// CloudProvider is the provider for cloud services.
	CloudProvider string `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	// ClusterName is the instance prefix for the cluster.
	ClusterName string `json:"clusterName,omitempty" flag:"cluster-name"`
	// Allow the cluster to run without the cluster-id on cloud instances
	AllowUntaggedCloud *bool `json:"allowUntaggedCloud,omitempty" flag:"allow-untagged-cloud"`
	// ClusterCIDR is CIDR Range for Pods in cluster.
	ClusterCIDR string `json:"clusterCIDR,omitempty" flag:"cluster-cidr"`
	// AllocateNodeCIDRs enables CIDRs for Pods to be allocated and, if
	// ConfigureCloudRoutes is true, to be set on the cloud provider.
	AllocateNodeCIDRs *bool `json:"allocateNodeCIDRs,omitempty" flag:"allocate-node-cidrs"`
	// ConfigureCloudRoutes enables CIDRs allocated with to be configured on the cloud provider.
	ConfigureCloudRoutes *bool `json:"configureCloudRoutes,omitempty" flag:"configure-cloud-routes"`
	// Controllers is a list of controllers to enable on the controller-manager
	Controllers []string `json:"controllers,omitempty" flag:"controllers"`
	// CIDRAllocatorType specifies the type of CIDR allocator to use.
	CIDRAllocatorType *string `json:"cidrAllocatorType,omitempty" flag:"cidr-allocator-type"`
	// LeaderElection defines the configuration of leader election client.
	LeaderElection *LeaderElectionConfiguration `json:"leaderElection,omitempty"`
	// UseServiceAccountCredentials controls whether we use individual service account credentials for each controller.
	UseServiceAccountCredentials *bool `json:"useServiceAccountCredentials,omitempty" flag:"use-service-account-credentials"`
	// EnableLeaderMigration enables controller leader migration.
	EnableLeaderMigration *bool `json:"enableLeaderMigration,omitempty" flag:"enable-leader-migration"`
	// CPURequest of NodeTerminationHandler container.
	// Default: 200m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// NodeStatusUpdateFrequency is the duration between node status updates. (default: 5m)
	NodeStatusUpdateFrequency *metav1.Duration `json:"nodeStatusUpdateFrequency,omitempty" flag:"node-status-update-frequency"`
	// ConcurrentNodeSyncs is the number of workers concurrently synchronizing nodes. (default: 1)
	ConcurrentNodeSyncs *int32 `json:"concurrentNodeSyncs,omitempty" flag:"concurrent-node-syncs"`
}

// KubeSchedulerConfig is the configuration for the kube-scheduler
type KubeSchedulerConfig struct {
	// Master is a url to the kube master
	Master string `json:"master,omitempty" flag:"master"`
	// LogFormat is the logging format of the scheduler.
	// Supported values: text, json.
	// Default: text
	LogFormat string `json:"logFormat,omitempty" flag:"logging-format" flag-empty:"text"`
	// LogLevel is the logging level
	LogLevel int32 `json:"logLevel,omitempty" flag:"v"`
	// Image is the container image to use.
	Image string `json:"image,omitempty"`
	// LeaderElection defines the configuration of leader election client.
	LeaderElection *LeaderElectionConfiguration `json:"leaderElection,omitempty"`
	// UsePolicyConfigMap enable setting the scheduler policy from a configmap
	// Deprecated - use KubeSchedulerConfiguration instead
	UsePolicyConfigMap *bool `json:"usePolicyConfigMap,omitempty"`
	// FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
	FeatureGates map[string]string `json:"featureGates,omitempty" flag:"feature-gates"`
	// MaxPersistentVolumes changes the maximum number of persistent volumes the scheduler will scheduler onto the same
	// node. Only takes effect if value is positive. This corresponds to the KUBE_MAX_PD_VOLS environment variable.
	// The default depends on the version and the cloud provider
	// as outlined: https://kubernetes.io/docs/concepts/storage/storage-limits/
	MaxPersistentVolumes *int32 `json:"maxPersistentVolumes,omitempty"`
	// Qps sets the maximum qps to send to apiserver after the burst quota is exhausted
	Qps *resource.Quantity `json:"qps,omitempty" configfile:"ClientConnection.QPS" config:"clientConnection.qps,omitempty"`
	// Burst sets the maximum qps to send to apiserver after the burst quota is exhausted
	Burst int32 `json:"burst,omitempty" configfile:"ClientConnection.Burst" config:"clientConnection.burst,omitempty"`
	// KubeAPIQPS QPS to use while talking with kubernetes apiserver. (default 20)
	KubeAPIQPS *resource.Quantity `json:"kubeAPIQPS,omitempty" flag:"kube-api-qps"`
	// KubeAPIBurst Burst to use while talking with kubernetes apiserver. (default 30)
	KubeAPIBurst *int32 `json:"kubeAPIBurst,omitempty" flag:"kube-api-burst"`
	// AuthenticationKubeconfig is the path to an Authentication Kubeconfig
	AuthenticationKubeconfig string `json:"authenticationKubeconfig,omitempty" flag:"authentication-kubeconfig"`
	// AuthorizationKubeconfig is the path to an Authorization Kubeconfig
	AuthorizationKubeconfig string `json:"authorizationKubeconfig,omitempty" flag:"authorization-kubeconfig"`
	// AuthorizationAlwaysAllowPaths is the list of HTTP paths to skip during authorization
	AuthorizationAlwaysAllowPaths []string `json:"authorizationAlwaysAllowPaths,omitempty" flag:"authorization-always-allow-paths"`

	// EnableProfiling enables profiling via web interface host:port/debug/pprof/
	EnableProfiling *bool `json:"enableProfiling,omitempty" flag:"profiling"`
	// EnableContentionProfiling enables block profiling, if profiling is enabled
	EnableContentionProfiling *bool `json:"enableContentionProfiling,omitempty" flag:"contention-profiling"`
	// TLSCertFile is the file containing the TLS server certificate.
	TLSCertFile *string `json:"tlsCertFile,omitempty" flag:"tls-cert-file"`
	// TLSPrivateKeyFile is the file containing the private key for the TLS server certificate.
	TLSPrivateKeyFile string `json:"tlsPrivateKeyFile,omitempty" flag:"tls-private-key-file"`

	// CPURequest, cpu request compute resource for scheduler. Defaults to "100m"
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// CPULimit, cpu limit compute resource for scheduler e.g. "500m"
	CPULimit *resource.Quantity `json:"cpuLimit,omitempty"`
	// MemoryRequest, memory request compute resource for scheduler e.g. "30Mi"
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// MemoryLimit, memory limit compute resource for scheduler e.g. "30Mi"
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
}

// LeaderElectionConfiguration defines the configuration of leader election
// clients for components that can run with leader election enabled.
type LeaderElectionConfiguration struct {
	// leaderElect enables a leader election client to gain leadership
	// before executing the main loop. Enable this when running replicated
	// components for high availability.
	LeaderElect *bool `json:"leaderElect,omitempty" flag:"leader-elect"`
	// leaderElectLeaseDuration is the length in time non-leader candidates
	// will wait after observing a leadership renewal until attempting to acquire
	// leadership of a led but unrenewed leader slot. This is effectively the
	// maximum duration that a leader can be stopped before it is replaced by another candidate
	LeaderElectLeaseDuration *metav1.Duration `json:"leaderElectLeaseDuration,omitempty" flag:"leader-elect-lease-duration"`
	// LeaderElectRenewDeadlineDuration is the interval between attempts by the acting master to
	// renew a leadership slot before it stops leading. This must be less than or equal to the lease duration.
	LeaderElectRenewDeadlineDuration *metav1.Duration `json:"leaderElectRenewDeadlineDuration,omitempty" flag:"leader-elect-renew-deadline"`
	// LeaderElectResourceLock is the type of resource object that is used for locking during
	// leader election. Supported options are endpoints (default) and `configmaps`.
	LeaderElectResourceLock *string `json:"leaderElectResourceLock,omitempty" flag:"leader-elect-resource-lock"`
	// LeaderElectResourceName is the name of resource object that is used for locking during leader election.
	LeaderElectResourceName *string `json:"leaderElectResourceName,omitempty" flag:"leader-elect-resource-name"`
	// LeaderElectResourceNamespace is the namespace of resource object that is used for locking during leader election.
	LeaderElectResourceNamespace *string `json:"leaderElectResourceNamespace,omitempty" flag:"leader-elect-resource-namespace"`
	// LeaderElectRetryPeriod is The duration the clients should wait between attempting acquisition
	// and renewal of a leadership. This is only applicable if leader election is enabled.
	LeaderElectRetryPeriod *metav1.Duration `json:"leaderElectRetryPeriod,omitempty" flag:"leader-elect-retry-period"`
}

// OpenstackLoadbalancerConfig defines the config for a neutron loadbalancer
type OpenstackLoadbalancerConfig struct {
	Method                *string `json:"method,omitempty"`
	Provider              *string `json:"provider,omitempty"`
	UseOctavia            *bool   `json:"useOctavia,omitempty"`
	FloatingNetwork       *string `json:"floatingNetwork,omitempty"`
	FloatingNetworkID     *string `json:"floatingNetworkID,omitempty"`
	FloatingSubnet        *string `json:"floatingSubnet,omitempty"`
	SubnetID              *string `json:"subnetID,omitempty"`
	ManageSecGroups       *bool   `json:"manageSecurityGroups,omitempty"`
	EnableIngressHostname *bool   `json:"enableIngressHostname,omitempty"`
	IngressHostnameSuffix *string `json:"ingressHostnameSuffix,omitempty"`
	FlavorID              *string `json:"flavorID,omitempty"`
}

type OpenstackBlockStorageConfig struct {
	Version                  *string `json:"bs-version,omitempty"`
	IgnoreAZ                 *bool   `json:"ignore-volume-az,omitempty"`
	OverrideAZ               *string `json:"override-volume-az,omitempty"`
	IgnoreVolumeMicroVersion *bool   `json:"ignore-volume-microversion,omitempty"`
	MetricsEnabled           *bool   `json:"metricsEnabled,omitempty"`
	// CreateStorageClass provisions a default class for the Cinder plugin
	CreateStorageClass *bool  `json:"createStorageClass,omitempty"`
	CSIPluginImage     string `json:"csiPluginImage,omitempty"`
	CSITopologySupport *bool  `json:"csiTopologySupport,omitempty"`
	// ClusterName sets the --cluster flag for the cinder-csi-plugin to the provided name
	ClusterName string `json:"clusterName,omitempty"`
}

// OpenstackMonitor defines the config for a health monitor
type OpenstackMonitor struct {
	Delay      *string `json:"delay,omitempty"`
	Timeout    *string `json:"timeout,omitempty"`
	MaxRetries *int    `json:"maxRetries,omitempty"`
}

// OpenstackRouter defines the config for a router
type OpenstackRouter struct {
	ExternalNetwork       *string   `json:"externalNetwork,omitempty"`
	DNSServers            *string   `json:"dnsServers,omitempty"`
	ExternalSubnet        *string   `json:"externalSubnet,omitempty"`
	AvailabilityZoneHints []*string `json:"availabilityZoneHints,omitempty"`
}

// OpenstackNetwork defines the config for a network
type OpenstackNetwork struct {
	AvailabilityZoneHints []*string `json:"availabilityZoneHints,omitempty"`
	IPv6SupportDisabled   *bool     `json:"ipv6SupportDisabled,omitempty"`
	PublicNetworkNames    []*string `json:"publicNetworkNames,omitempty"`
	InternalNetworkNames  []*string `json:"internalNetworkNames,omitempty"`
	AddressSortOrder      *string   `json:"addressSortOrder,omitempty"`
}

// OpenstackMetadata defines config for metadata service related settings
type OpenstackMetadata struct {
	// ConfigDrive specifies to use config drive for retrieving user data instead of the metadata service when launching instances
	ConfigDrive *bool `json:"configDrive,omitempty"`
}

// OpenstackSpec defines cloud config elements for the openstack cloud provider
type OpenstackSpec struct {
	Loadbalancer       *OpenstackLoadbalancerConfig `json:"loadbalancer,omitempty"`
	Monitor            *OpenstackMonitor            `json:"monitor,omitempty"`
	Router             *OpenstackRouter             `json:"router,omitempty"`
	BlockStorage       *OpenstackBlockStorageConfig `json:"blockStorage,omitempty"`
	InsecureSkipVerify *bool                        `json:"insecureSkipVerify,omitempty"`
	Network            *OpenstackNetwork            `json:"network,omitempty"`
	Metadata           *OpenstackMetadata           `json:"metadata,omitempty"`
}

// AzureSpec defines Azure specific cluster configuration.
type AzureSpec struct {
	// SubscriptionID specifies the subscription used for the cluster installation.
	SubscriptionID string `json:"subscriptionID,omitempty"`
	// StorageAccountID specifies the storage account used for the cluster installation.
	StorageAccountID string `json:"storageAccountID,omitempty"`
	// TenantID is the ID of the tenant that the cluster is deployed in.
	TenantID string `json:"tenantID"`
	// ResourceGroupName specifies the name of the resource group
	// where the cluster is built.
	// If this is empty, kops will create a new resource group
	// whose name is same as the cluster name. If this is not
	// empty, kops will not create a new resource group, and
	// it will just reuse the existing resource group of the name.
	// This follows the model that kops takes for AWS VPC.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`
	// RouteTableName is the name of the route table attached to the subnet that the cluster is deployed in.
	RouteTableName string `json:"routeTableName,omitempty"`
	// AdminUser specifies the admin user of VMs.
	AdminUser string `json:"adminUser,omitempty"`
}

// CloudConfiguration defines the cloud provider configuration
type CloudConfiguration struct {
	// Cross-cloud provider options

	// ManageStorageClasses specifies whether kOps should create and maintain a set of
	// StorageClasses, one of which it nominates as the default class for the cluster.
	ManageStorageClasses *bool `json:"manageStorageClasses,omitempty"`
}

// EBSCSIDriverSpec is the config for the AWS EBS CSI driver
type EBSCSIDriverSpec struct {
	// Enabled enables the AWS EBS CSI driver. Can only be set to true.
	// Default: true
	Enabled *bool `json:"-"`

	// Managed controls if aws-ebs-csi-driver is manged and deployed by kOps.
	// The deployment of aws-ebs-csi-driver is skipped if this is set to false.
	Managed *bool `json:"managed,omitempty"`

	// Version is the container image tag used.
	// Default: The latest stable release which is compatible with your Kubernetes version
	Version *string `json:"version,omitempty"`

	// KubeAPIQPS QPS to use while talking with Kubernetes API server. (default 20)
	KubeAPIQPS *resource.Quantity `json:"kubeAPIQPS,omitempty"`
	// KubeAPIBurst Burst to use while talking with Kubernetes API server. (default 100)
	KubeAPIBurst *int32 `json:"kubeAPIBurst,omitempty"`

	// HostNetwork can be used for large clusters for faster access to node info via instance metadata.
	// Default: false
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// VolumeAttachLimit is the maximum number of volumes attachable per node.
	// If specified, the limit applies to all nodes.
	// If not specified, the value is approximated from the instance type.
	// Default: -
	VolumeAttachLimit *int `json:"volumeAttachLimit,omitempty"`

	// PodAnnotations are the annotations added to AWS EBS CSI node and controller Pods.
	// Default: none
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`
}

// PDCSIDriver is the config for the GCP PD CSI driver
type PDCSIDriver struct {
	// Enabled enables the GCP PD CSI driver
	Enabled *bool `json:"enabled,omitempty"`
}

// SnapshotControllerConfig is the config for the CSI Snapshot Controller
type SnapshotControllerConfig struct {
	// Enabled enables the CSI Snapshot Controller
	Enabled *bool `json:"enabled,omitempty"`
	// InstallDefaultClass will install the default VolumeSnapshotClass
	InstallDefaultClass bool `json:"installDefaultClass,omitempty"`
}

// NodeTerminationHandlerSpec determines the node termination handler configuration.
type NodeTerminationHandlerSpec struct {
	// DeleteSQSMsgIfNodeNotFound makes node termination handler delete the SQS Message from the SQS Queue if the targeted node is not found.
	// Only used in Queue Processor mode.
	// Default: false
	DeleteSQSMsgIfNodeNotFound *bool `json:"deleteSQSMsgIfNodeNotFound,omitempty"`
	// Enabled enables the node termination handler.
	// Default: true
	Enabled *bool `json:"enabled,omitempty"`
	// EnableSpotInterruptionDraining makes node termination handler drain nodes when spot interruption termination notice is received.
	// Cannot be disabled in queue-processor mode.
	// Default: true
	EnableSpotInterruptionDraining *bool `json:"enableSpotInterruptionDraining,omitempty"`
	// EnableScheduledEventDraining makes node termination handler drain nodes before the maintenance window starts for an EC2 instance scheduled event.
	// Cannot be disabled in queue-processor mode.
	// Default: true
	EnableScheduledEventDraining *bool `json:"enableScheduledEventDraining,omitempty"`
	// EnableRebalanceMonitoring makes node termination handler cordon nodes when the rebalance recommendation notice is received.
	// In queue-processor mode, cannot be enabled without rebalance draining.
	// Default: false
	EnableRebalanceMonitoring *bool `json:"enableRebalanceMonitoring,omitempty"`
	// EnableRebalanceDraining makes node termination handler drain nodes when the rebalance recommendation notice is received.
	// Default: false
	EnableRebalanceDraining *bool `json:"enableRebalanceDraining,omitempty"`

	// EnablePrometheusMetrics enables the "/metrics" endpoint.
	// Default: false
	EnablePrometheusMetrics *bool `json:"prometheusEnable,omitempty"`

	// EnableSQSTerminationDraining enables queue-processor mode which drains nodes when an SQS termination event is received.
	// Default: true
	EnableSQSTerminationDraining *bool `json:"enableSQSTerminationDraining,omitempty"`

	// ExcludeFromLoadBalancers makes node termination handler will mark for exclusion from load balancers before node are cordoned.
	// Default: true
	ExcludeFromLoadBalancers *bool `json:"excludeFromLoadBalancers,omitempty"`

	// ManagedASGTag is the tag used to determine which nodes NTH can take action on
	// This field has kept its name even though it now maps to the --managed-tag flag due to keeping the API stable.
	// Node termination handler does no longer check the ASG for this tag, but the actual EC2 instances.
	ManagedASGTag *string `json:"managedASGTag,omitempty"`

	// PodTerminationGracePeriod is the time in seconds given to each pod to terminate gracefully.
	// If negative, the default value specified in the pod will be used, which defaults to 30 seconds if not specified for the pod.
	// Default: -1
	PodTerminationGracePeriod *int32 `json:"podTerminationGracePeriod,omitempty"`

	// TaintNode makes node termination handler taint nodes when an interruption event occurs.
	// Default: false
	TaintNode *bool `json:"taintNode,omitempty"`

	// MemoryLimit of NodeTerminationHandler container.
	// Default: none
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
	// MemoryRequest of NodeTerminationHandler container.
	// Default: 64Mi
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest of NodeTerminationHandler container.
	// Default: 50m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// Version is the container image tag used.
	Version *string `json:"version,omitempty"`

	// Replaces the default webhook message template.
	WebhookTemplate *string `json:"webhookTemplate,omitempty"`
	// If specified, posts event data to URL upon instance interruption action.
	WebhookURL *string `json:"webhookURL,omitempty"`
}

func (n *NodeTerminationHandlerSpec) IsQueueMode() bool {
	return n != nil && n.Enabled != nil && *n.Enabled && (n.EnableSQSTerminationDraining == nil || *n.EnableSQSTerminationDraining)
}

// NodeProblemDetector determines the node problem detector configuration.
type NodeProblemDetectorConfig struct {
	// Enabled enables the NodeProblemDetector.
	// Default: false
	Enabled *bool `json:"enabled,omitempty"`
	// Image is the NodeProblemDetector container image used.
	Image *string `json:"image,omitempty"`

	// MemoryRequest of NodeProblemDetector container.
	// Default: 80Mi
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest of NodeProblemDetector container.
	// Default: 10m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// MemoryLimit of NodeProblemDetector container.
	// Default: 80Mi
	MemoryLimit *resource.Quantity `json:"memoryLimit,omitempty"`
	// CPULimit of NodeProblemDetector container.
	CPULimit *resource.Quantity `json:"cpuLimit,omitempty"`
}

// ClusterAutoscalerConfig determines the cluster autoscaler configuration.
type ClusterAutoscalerConfig struct {
	// Enabled enables the cluster autoscaler.
	// Default: false
	Enabled *bool `json:"enabled,omitempty"`
	// Expander determines the strategy for which instance group gets expanded.
	// Supported values: least-waste, most-pods, random, price, priority.
	// The price expander is only supported on GCE.
	// By default, kOps will generate the priority expander ConfigMap based on the `autoscale` and `autoscalePriority` fields in the InstanceGroup specs.
	// Default: least-waste
	Expander string `json:"expander,omitempty"`
	// BalanceSimilarNodeGroups makes the cluster autoscaler treat similar node groups as one.
	// Default: false
	BalanceSimilarNodeGroups *bool `json:"balanceSimilarNodeGroups,omitempty"`
	// EmitPerNodegroupMetrics If true, publishes the node groups min and max metrics count set on the cluster autoscaler.
	// Default: false
	EmitPerNodegroupMetrics *bool `json:"emitPerNodegroupMetrics,omitempty"`
	// AWSUseStaticInstanceList makes cluster autoscaler to use statically defined set of AWS EC2 Instance List.
	// Default: false
	AWSUseStaticInstanceList *bool `json:"awsUseStaticInstanceList,omitempty"`
	// IgnoreDaemonSetsUtilization causes the cluster autoscaler to ignore DaemonSet-managed pods when calculating resource utilization for scaling down.
	// Default: false
	IgnoreDaemonSetsUtilization *bool `json:"ignoreDaemonSetsUtilization,omitempty"`
	// ScaleDownUtilizationThreshold determines the utilization threshold for node scale-down.
	// Default: 0.5
	ScaleDownUtilizationThreshold *string `json:"scaleDownUtilizationThreshold,omitempty"`
	// SkipNodesWithCustomControllerPods makes the cluster autoscaler skip scale-down of nodes with pods owned by custom controllers.
	// Default: true
	SkipNodesWithCustomControllerPods *bool `json:"skipNodesWithCustomControllerPods,omitempty"`
	// SkipNodesWithSystemPods makes the cluster autoscaler skip scale-down of nodes with non-DaemonSet pods in the kube-system namespace.
	// Default: true
	SkipNodesWithSystemPods *bool `json:"skipNodesWithSystemPods,omitempty"`
	// SkipNodesWithLocalStorage makes the cluster autoscaler skip scale-down of nodes with local storage.
	// Default: true
	SkipNodesWithLocalStorage *bool `json:"skipNodesWithLocalStorage,omitempty"`
	// NewPodScaleUpDelay causes the cluster autoscaler to ignore unschedulable pods until they are a certain "age", regardless of the scan-interval
	// Default: 0s
	NewPodScaleUpDelay *string `json:"newPodScaleUpDelay,omitempty"`
	// ScaleDownDelayAfterAdd determines the time after scale up that scale down evaluation resumes
	// Default: 10m0s
	ScaleDownDelayAfterAdd *string `json:"scaleDownDelayAfterAdd,omitempty"`
	// scaleDownUnneededTime determines the time a node should be unneeded before it is eligible for scale down
	// Default: 10m0s
	ScaleDownUnneededTime *string `json:"scaleDownUnneededTime,omitempty"`
	// ScaleDownUnreadyTime determines the time an unready node should be unneeded before it is eligible for scale down
	// Default: 20m0s
	ScaleDownUnreadyTime *string `json:"scaleDownUnreadyTime,omitempty"`
	// CordonNodeBeforeTerminating should CA cordon nodes before terminating during downscale process
	// Default: false
	CordonNodeBeforeTerminating *bool `json:"cordonNodeBeforeTerminating,omitempty"`
	// Image is the container image used.
	// Default: the latest supported image for the specified kubernetes version.
	Image *string `json:"image,omitempty"`
	// MemoryRequest of cluster autoscaler container.
	// Default: 300Mi
	MemoryRequest *resource.Quantity `json:"memoryRequest,omitempty"`
	// CPURequest of cluster autoscaler container.
	// Default: 100m
	CPURequest *resource.Quantity `json:"cpuRequest,omitempty"`
	// MaxNodeProvisionTime determines how long CAS will wait for a node to join the cluster.
	MaxNodeProvisionTime string `json:"maxNodeProvisionTime,omitempty"`
	// PodAnnotations are the annotations added to cluster autoscaler pods when they are created.
	// Default: none
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`
	// CreatePriorityExpenderConfig makes kOps create the priority-expander ConfigMap
	// Default: true
	CreatePriorityExpenderConfig *bool `json:"createPriorityExpanderConfig,omitempty"`
	// CustomPriorityExpanderConfig overides the priority-expander ConfigMap with the provided configuration. Any InstanceGroup configuration will be ignored if this is set.
	// This could be useful in order to use regex on priorities configuration
	CustomPriorityExpanderConfig map[string][]string `json:"customPriorityExpanderConfig,omitempty"`
}

// MetricsServerConfig determines the metrics server configuration.
type MetricsServerConfig struct {
	// Enabled enables the metrics server.
	// Default: false
	Enabled *bool `json:"enabled,omitempty"`
	// Image is the container image used.
	// Default: the latest supported image for the specified kubernetes version.
	Image *string `json:"image,omitempty"`
	// Insecure determines if API server will validate metrics server TLS cert.
	// Default: true
	Insecure *bool `json:"insecure,omitempty"`
}

// CertManagerConfig determines the cert manager configuration.
type CertManagerConfig struct {
	// Enabled enables the cert manager.
	// Default: false
	Enabled *bool `json:"enabled,omitempty"`

	// Managed controls if cert-manager is manged and deployed by kOps.
	// The deployment of cert-manager is skipped if this is set to false.
	Managed *bool `json:"managed,omitempty"`

	// Image is the container image used.
	// Default: the latest supported image for the specified kubernetes version.
	Image *string `json:"image,omitempty"`

	// defaultIssuer sets a default clusterIssuer
	// Default: none
	DefaultIssuer *string `json:"defaultIssuer,omitempty"`

	// nameservers is a list of nameserver IP addresses to use instead of the pod defaults.
	// Default: none
	Nameservers []string `json:"nameservers,omitempty"`

	// HostedZoneIDs is a list of route53 hostedzone IDs that cert-manager will be allowed to do dns-01 validation for
	HostedZoneIDs []string `json:"hostedZoneIDs,omitempty"`

	// FeatureGates is a list of experimental features that can be enabled or disabled.
	FeatureGates map[string]bool `json:"featureGates,omitempty"`
}

// LoadBalancerControllerSpec determines the AWS LB controller configuration.
type LoadBalancerControllerSpec struct {
	// Enabled enables the loadbalancer controller.
	// Default: false
	Enabled *bool `json:"enabled,omitempty"`
	// Version is the container image tag used.
	Version *string `json:"version,omitempty"`
	// EnableWAF specifies whether the controller can use WAFs (Classic Regional).
	// Default: false
	EnableWAF bool `json:"enableWAF,omitempty"`
	// EnableWAFv2 specifies whether the controller can use WAFs (V2).
	// Default: false
	EnableWAFv2 bool `json:"enableWAFv2,omitempty"`
	// EnableShield specifies whether the controller can enable Shield Advanced.
	// Default: false
	EnableShield bool `json:"enableShield,omitempty"`
}

// HasAdmissionController checks if a specific admission controller is enabled
func (c *KubeAPIServerConfig) HasAdmissionController(name string) bool {
	for _, x := range c.AdmissionControl {
		if x == name {
			return true
		}
	}

	for _, x := range c.DisableAdmissionPlugins {
		if x == name {
			return false
		}
	}
	for _, x := range c.EnableAdmissionPlugins {
		if x == name {
			return true
		}
	}

	return false
}
