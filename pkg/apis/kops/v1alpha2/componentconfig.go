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
	ClientCAFile string `json:"clientCaFile,omitempty" flag:"client-ca-file"`
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
	// LogLevel is the logging level of the kubelet
	LogLevel *int32 `json:"logLevel,omitempty" flag:"v" flag-empty:"0"`
	// config is the path to the config file or directory of files
	PodManifestPath string `json:"podManifestPath,omitempty" flag:"pod-manifest-path"`
	// HostnameOverride is the hostname used to identify the kubelet instead of the actual hostname.
	HostnameOverride string `json:"hostnameOverride,omitempty" flag:"hostname-override"`
	// PodInfraContainerImage is the image whose network/ipc containers in each pod will use.
	PodInfraContainerImage string `json:"podInfraContainerImage,omitempty" flag:"pod-infra-container-image"`
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
	NetworkPluginName string `json:"networkPluginName,omitempty" flag:"network-plugin"`
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
	//// SerializeImagePulls when enabled, tells the Kubelet to pull images one
	//// at a time. We recommend *not* changing the default value on nodes that
	//// run docker daemon with version  < 1.9 or an Aufs storage backend.
	//// Issue #10959 has more details.
	SerializeImagePulls *bool `json:"serializeImagePulls,omitempty" flag:"serialize-image-pulls"`
	// NodeLabels to add when registering the node in the cluster.
	NodeLabels map[string]string `json:"nodeLabels,omitempty" flag:"node-labels"`
	// NonMasqueradeCIDR configures masquerading: traffic to IPs outside this range will use IP masquerade.
	NonMasqueradeCIDR string `json:"nonMasqueradeCIDR,omitempty" flag:"non-masquerade-cidr"`
	// Enable gathering custom metrics.
	EnableCustomMetrics *bool `json:"enableCustomMetrics,omitempty" flag:"enable-custom-metrics"`
	// NetworkPluginMTU is the MTU to be passed to the network plugin,
	// and overrides the default MTU for cases where it cannot be automatically
	// computed (such as IPSEC).
	NetworkPluginMTU *int32 `json:"networkPluginMTU,omitempty" flag:"network-plugin-mtu"`
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
	// DockerDisableSharedPID uses a shared PID namespace for containers in a pod.
	DockerDisableSharedPID *bool `json:"dockerDisableSharedPID,omitempty" flag:"docker-disable-shared-pid"`
	// RootDir is the directory path for managing kubelet files (volume mounts,etc)
	RootDir string `json:"rootDir,omitempty" flag:"root-dir"`
	// AuthenticationTokenWebhook uses the TokenReview API to determine authentication for bearer tokens.
	AuthenticationTokenWebhook *bool `json:"authenticationTokenWebhook,omitempty" flag:"authentication-token-webhook"`
	// AuthenticationTokenWebhook sets the duration to cache responses from the webhook token authenticator. Default is 2m. (default 2m0s)
	AuthenticationTokenWebhookCacheTTL *metav1.Duration `json:"authenticationTokenWebhookCacheTtl,omitempty" flag:"authentication-token-webhook-cache-ttl"`
	// CPUCFSQuota enables CPU CFS quota enforcement for containers that specify CPU limits
	CPUCFSQuota *bool `json:"cpuCFSQuota,omitempty" flag:"cpu-cfs-quota"`
	// CPUCFSQuotaPeriod sets CPU CFS quota period value, cpu.cfs_period_us, defaults to Linux Kernel default
	CPUCFSQuotaPeriod *metav1.Duration `json:"cpuCFSQuotaPeriod,omitempty" flag:"cpu-cfs-quota-period"`
	// CpuManagerPolicy allows for changing the default policy of None to static
	CpuManagerPolicy string `json:"cpuManagerPolicy,omitempty" flag:"cpu-manager-policy"`
	// RegistryPullQPS if > 0, limit registry pull QPS to this value.  If 0, unlimited. (default 5)
	RegistryPullQPS *int32 `json:"registryPullQPS,omitempty" flag:"registry-qps"`
	//RegistryBurst Maximum size of a bursty pulls, temporarily allows pulls to burst to this number, while still not exceeding registry-qps. Only used if --registry-qps > 0 (default 10)
	RegistryBurst *int32 `json:"registryBurst,omitempty" flag:"registry-burst"`

	// rotateCertificates enables client certificate rotation.
	RotateCertificates *bool `json:"rotateCertificates,omitempty" flag:"rotate-certificates"`
}

// KubeProxyConfig defines the configuration for a proxy
type KubeProxyConfig struct {
	Image string `json:"image,omitempty"`
	// TODO: Better type ?
	// CPURequest, cpu request compute resource for kube proxy e.g. "20m"
	CPURequest string `json:"cpuRequest,omitempty"`
	// CPULimit, cpu limit compute resource for kube proxy e.g. "30m"
	CPULimit string `json:"cpuLimit,omitempty"`
	// MemoryRequest, memory request compute resource for kube proxy e.g. "30Mi"
	MemoryRequest string `json:"memoryRequest,omitempty"`
	// MemoryLimit, memory limit compute resource for kube proxy e.g. "30Mi"
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// LogLevel is the logging level of the proxy
	LogLevel int32 `json:"logLevel,omitempty" flag:"v"`
	// ClusterCIDR is the CIDR range of the pods in the cluster
	ClusterCIDR string `json:"clusterCIDR,omitempty" flag:"cluster-cidr"`
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
	// Which proxy mode to use: (userspace, iptables, ipvs)
	ProxyMode string `json:"proxyMode,omitempty" flag:"proxy-mode"`
	// IPVSExcludeCIDRS is comma-separated list of CIDR's which the ipvs proxier should not touch when cleaning up IPVS rules
	IPVSExcludeCIDRS []string `json:"ipvsExcludeCidrs,omitempty" flag:"ipvs-exclude-cidrs"`
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
	// Image is the docker container used
	Image string `json:"image,omitempty"`
	// DisableBasicAuth removes the --basic-auth-file flag
	DisableBasicAuth bool `json:"disableBasicAuth,omitempty"`
	// LogLevel is the logging level of the api
	LogLevel int32 `json:"logLevel,omitempty" flag:"v" flag-empty:"0"`
	// CloudProvider is the name of the cloudProvider we are using, aws, gce etcd
	CloudProvider string `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	// SecurePort is the port the kube runs on
	SecurePort int32 `json:"securePort,omitempty" flag:"secure-port"`
	// InsecurePort is the port the insecure api runs
	InsecurePort int32 `json:"insecurePort,omitempty" flag:"insecure-port"`
	// Address is the binding address for the kube api: Deprecated - use insecure-bind-address and bind-address
	Address string `json:"address,omitempty" flag:"address"`
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
	EtcdCAFile string `json:"etcdCaFile,omitempty" flag:"etcd-cafile"`
	// EtcdCertFile is the path to a certificate
	EtcdCertFile string `json:"etcdCertFile,omitempty" flag:"etcd-certfile"`
	// EtcdKeyFile is the path to a private key
	EtcdKeyFile string `json:"etcdKeyFile,omitempty" flag:"etcd-keyfile"`
	// TODO: Remove unused BasicAuthFile
	BasicAuthFile string `json:"basicAuthFile,omitempty" flag:"basic-auth-file"`
	// TODO: Remove unused ClientCAFile
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
	OIDCUsernameClaim *string `json:"oidcUsernameClaim,omitempty" flag:"oidc-username-claim"`
	// OIDCUsernamePrefix is the prefix prepended to username claims to prevent
	// clashes with existing names (such as 'system:' users).
	OIDCUsernamePrefix *string `json:"oidcUsernamePrefix,omitempty" flag:"oidc-username-prefix"`
	// OIDCGroupsClaim if provided, the name of a custom OpenID Connect claim for
	// specifying user groups.
	// The claim value is expected to be a string or array of strings.
	OIDCGroupsClaim *string `json:"oidcGroupsClaim,omitempty" flag:"oidc-groups-claim"`
	// OIDCGroupsPrefix is the prefix prepended to group claims to prevent
	// clashes with existing names (such as 'system:' groups)
	OIDCGroupsPrefix *string `json:"oidcGroupsPrefix,omitempty" flag:"oidc-groups-prefix"`
	// OIDCIssuerURL is the URL of the OpenID issuer, only HTTPS scheme will
	// be accepted.
	// If set, it will be used to verify the OIDC JSON Web Token (JWT).
	OIDCIssuerURL *string `json:"oidcIssuerURL,omitempty" flag:"oidc-issuer-url"`
	// OIDCClientID is the client ID for the OpenID Connect client, must be set
	// if oidc-issuer-url is set.
	OIDCClientID *string `json:"oidcClientID,omitempty" flag:"oidc-client-id"`
	// A key=value pair that describes a required claim in the ID Token.
	// If set, the claim is verified to be present in the ID Token with a matching value.
	// Repeat this flag to specify multiple claims.
	OIDCRequiredClaim []string `json:"oidcRequiredClaim,omitempty" flag:"oidc-required-claim,repeat"`
	// OIDCCAFile if set, the OpenID server's certificate will be verified by one
	// of the authorities in the oidc-ca-file
	OIDCCAFile *string `json:"oidcCAFile,omitempty" flag:"oidc-ca-file"`
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
	AuthorizationWebhookCacheAuthorizedTTL *metav1.Duration `json:"authorizationWebhookCacheAuthorizedTtl,omitempty" flag:"authorization-webhook-cache-authorized-ttl"`
	// The duration to cache authorized responses from the webhook token authorizer. Default is 30s. (default 30s)
	AuthorizationWebhookCacheUnauthorizedTTL *metav1.Duration `json:"authorizationWebhookCacheUnauthorizedTtl,omitempty" flag:"authorization-webhook-cache-unauthorized-ttl"`
	// AuthorizationRBACSuperUser is the name of the superuser for default rbac
	AuthorizationRBACSuperUser *string `json:"authorizationRbacSuperUser,omitempty" flag:"authorization-rbac-super-user"`
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
	//Root certificate bundle to use to verify client certificates on incoming requests before trusting usernames in headers specified by --requestheader-username-headers
	RequestheaderClientCAFile string `json:"requestheaderClientCAFile,omitempty" flag:"requestheader-client-ca-file"`
	// List of client certificate common names to allow to provide usernames in headers specified by --requestheader-username-headers. If empty, any client certificate validated by the authorities in --requestheader-client-ca-file is allowed.
	RequestheaderAllowedNames []string `json:"requestheaderAllowedNames,omitempty" flag:"requestheader-allowed-names"`
	// FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
	FeatureGates map[string]string `json:"featureGates,omitempty" flag:"feature-gates"`
	// MaxRequestsInflight The maximum number of non-mutating requests in flight at a given time.
	MaxRequestsInflight int32 `json:"maxRequestsInflight,omitempty" flag:"max-requests-inflight" flag-empty:"0"`
	// MaxMutatingRequestsInflight The maximum number of mutating requests in flight at a given time. Defaults to 200
	MaxMutatingRequestsInflight int32 `json:"maxMutatingRequestsInflight,omitempty" flag:"max-mutating-requests-inflight" flag-empty:"0"`

	// HTTP2MaxStreamsPerConnection sets the limit that the server gives to clients for the maximum number of streams in an HTTP/2 connection. Zero means to use golang's default.
	HTTP2MaxStreamsPerConnection *int32 `json:"http2MaxStreamsPerConnection,omitempty" flag:"http2-max-streams-per-connection"`

	// EtcdQuorumRead configures the etcd-quorum-read flag, which forces consistent reads from etcd
	EtcdQuorumRead *bool `json:"etcdQuorumRead,omitempty" flag:"etcd-quorum-read"`

	// MinRequestTimeout configures the minimum number of seconds a handler must keep a request open before timing it out.
	// Currently only honored by the watch request handler
	MinRequestTimeout *int32 `json:"minRequestTimeout,omitempty" flag:"min-request-timeout"`

	// Memory limit for apiserver in MB (used to configure sizes of caches, etc.)
	TargetRamMb int32 `json:"targetRamMb,omitempty" flag:"target-ram-mb" flag-empty:"0"`

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

	// Identifiers of the API. The service account token authenticator will validate that
	// tokens used against the API are bound to at least one of these audiences. If the
	// --service-account-issuer flag is configured and this flag is not, this field
	// defaults to a single element list containing the issuer URL.
	APIAudiences []string `json:"apiAudiences,omitempty" flag:"api-audiences"`

	// CPURequest, cpu request compute resource for api server. Defaults to "150m"
	CPURequest string `json:"cpuRequest,omitempty"`

	// Amount of time to retain Kubernetes events
	EventTTL *metav1.Duration `json:"eventTTL,omitempty" flag:"event-ttl"`

	// AuditDynamicConfiguration enables dynamic audit configuration via AuditSinks
	AuditDynamicConfiguration *bool `json:"auditDynamicConfiguration,omitempty" flag:"audit-dynamic-configuration"`
}

// KubeControllerManagerConfig is the configuration for the controller
type KubeControllerManagerConfig struct {
	// Master is the url for the kube api master
	Master string `json:"master,omitempty" flag:"master"`
	// LogLevel is the defined logLevel
	LogLevel int32 `json:"logLevel,omitempty" flag:"v" flag-empty:"0"`
	// ServiceAccountPrivateKeyFile the location for a certificate for service account signing
	ServiceAccountPrivateKeyFile string `json:"serviceAccountPrivateKeyFile,omitempty" flag:"service-account-private-key-file"`
	// Image is the docker image to use
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
	// ReconcilerSyncLoopPeriod is the amount of time the reconciler sync states loop
	// wait between successive executions. Is set to 1 min by kops by default
	AttachDetachReconcileSyncPeriod *metav1.Duration `json:"attachDetachReconcileSyncPeriod,omitempty" flag:"attach-detach-reconcile-sync-period"`
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
	// HorizontalPodAutoscalerTolerance is the minimum change (from 1.0) in the
	// desired-to-actual metrics ratio for the horizontal pod autoscaler to
	// consider scaling.
	HorizontalPodAutoscalerTolerance *resource.Quantity `json:"horizontalPodAutoscalerTolerance,omitempty" flag:"horizontal-pod-autoscaler-tolerance"`
	// HorizontalPodAutoscalerUseRestClients determines if the new-style clients
	// should be used if support for custom metrics is enabled.
	HorizontalPodAutoscalerUseRestClients *bool `json:"horizontalPodAutoscalerUseRestClients,omitempty" flag:"horizontal-pod-autoscaler-use-rest-clients"`
	// ExperimentalClusterSigningDuration is the duration that determines
	// the length of duration that the signed certificates will be given. (default 8760h0m0s)
	ExperimentalClusterSigningDuration *metav1.Duration `json:"experimentalClusterSigningDuration,omitempty" flag:"experimental-cluster-signing-duration"`
	// FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
	FeatureGates map[string]string `json:"featureGates,omitempty" flag:"feature-gates"`
	// TLSCipherSuites indicates the allowed TLS cipher suite
	TLSCipherSuites []string `json:"tlsCipherSuites,omitempty" flag:"tls-cipher-suites"`
	// TLSMinVersion indicates the minimum TLS version allowed
	TLSMinVersion string `json:"tlsMinVersion,omitempty" flag:"tls-min-version"`
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
	// This only works on kubernetes >= 1.14
	ConcurrentRcSyncs *int32 `json:"concurrentRcSyncs,omitempty" flag:"concurrent-rc-syncs"`
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
	// ClusterCIDR is CIDR Range for Pods in cluster.
	ClusterCIDR string `json:"clusterCIDR,omitempty" flag:"cluster-cidr"`
	// AllocateNodeCIDRs enables CIDRs for Pods to be allocated and, if
	// ConfigureCloudRoutes is true, to be set on the cloud provider.
	AllocateNodeCIDRs *bool `json:"allocateNodeCIDRs,omitempty" flag:"allocate-node-cidrs"`
	// ConfigureCloudRoutes enables CIDRs allocated with to be configured on the cloud provider.
	ConfigureCloudRoutes *bool `json:"configureCloudRoutes,omitempty" flag:"configure-cloud-routes"`
	// CIDRAllocatorType specifies the type of CIDR allocator to use.
	CIDRAllocatorType *string `json:"cidrAllocatorType,omitempty" flag:"cidr-allocator-type"`
	// LeaderElection defines the configuration of leader election client.
	LeaderElection *LeaderElectionConfiguration `json:"leaderElection,omitempty"`
	// UseServiceAccountCredentials controls whether we use individual service account credentials for each controller.
	UseServiceAccountCredentials *bool `json:"useServiceAccountCredentials,omitempty" flag:"use-service-account-credentials"`
}

// KubeSchedulerConfig is the configuration for the kube-scheduler
type KubeSchedulerConfig struct {
	// Master is a url to the kube master
	Master string `json:"master,omitempty" flag:"master"`
	// LogLevel is the logging level
	LogLevel int32 `json:"logLevel,omitempty" flag:"v"`
	// Image is the docker image to use
	Image string `json:"image,omitempty"`
	// LeaderElection defines the configuration of leader election client.
	LeaderElection *LeaderElectionConfiguration `json:"leaderElection,omitempty"`
	// UsePolicyConfigMap enable setting the scheduler policy from a configmap
	UsePolicyConfigMap *bool `json:"usePolicyConfigMap,omitempty"`
	// FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
	FeatureGates map[string]string `json:"featureGates,omitempty" flag:"feature-gates"`
	// MaxPersistentVolumes changes the maximum number of persistent volumes the scheduler will scheduler onto the same
	// node. Only takes into affect if value is positive. This corresponds to the KUBE_MAX_PD_VOLS environment variable,
	// which has been supported as far back as Kubernetes 1.7. The default depends on the version and the cloud provider
	// as outlined: https://kubernetes.io/docs/concepts/storage/storage-limits/
	MaxPersistentVolumes *int32 `json:"maxPersistentVolumes,omitempty"`
	// Qps sets the maximum qps to send to apiserver after the burst quota is exhausted
	Qps *resource.Quantity `json:"qps,omitempty"`
	// Burst sets the maximum qps to send to apiserver after the burst quota is exhausted
	Burst int32 `json:"burst,omitempty"`
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
	Method            *string `json:"method,omitempty"`
	Provider          *string `json:"provider,omitempty"`
	UseOctavia        *bool   `json:"useOctavia,omitempty"`
	FloatingNetwork   *string `json:"floatingNetwork,omitempty"`
	FloatingNetworkID *string `json:"floatingNetworkID,omitempty"`
	FloatingSubnet    *string `json:"floatingSubnet,omitempty"`
	SubnetID          *string `json:"subnetID,omitempty"`
	ManageSecGroups   *bool   `json:"manageSecurityGroups,omitempty"`
}

type OpenstackBlockStorageConfig struct {
	Version    *string `json:"bs-version,omitempty"`
	IgnoreAZ   *bool   `json:"ignore-volume-az,omitempty"`
	OverrideAZ *string `json:"override-volume-az,omitempty"`
}

// OpenstackMonitor defines the config for a health monitor
type OpenstackMonitor struct {
	Delay      *string `json:"delay,omitempty"`
	Timeout    *string `json:"timeout,omitempty"`
	MaxRetries *int    `json:"maxRetries,omitempty"`
}

// OpenstackRouter defines the config for a router
type OpenstackRouter struct {
	ExternalNetwork *string `json:"externalNetwork,omitempty"`
	DNSServers      *string `json:"dnsServers,omitempty"`
	ExternalSubnet  *string `json:"externalSubnet,omitempty"`
}

// OpenstackConfiguration defines cloud config elements for the openstack cloud provider
type OpenstackConfiguration struct {
	Loadbalancer       *OpenstackLoadbalancerConfig `json:"loadbalancer,omitempty"`
	Monitor            *OpenstackMonitor            `json:"monitor,omitempty"`
	Router             *OpenstackRouter             `json:"router,omitempty"`
	BlockStorage       *OpenstackBlockStorageConfig `json:"blockStorage,omitempty"`
	InsecureSkipVerify *bool                        `json:"insecureSkipVerify,omitempty"`
}

// CloudConfiguration defines the cloud provider configuration
type CloudConfiguration struct {
	// GCE cloud-config options
	Multizone          *bool   `json:"multizone,omitempty"`
	NodeTags           *string `json:"nodeTags,omitempty"`
	NodeInstancePrefix *string `json:"nodeInstancePrefix,omitempty"`
	// AWS cloud-config options
	DisableSecurityGroupIngress *bool   `json:"disableSecurityGroupIngress,omitempty"`
	ElbSecurityGroup            *string `json:"elbSecurityGroup,omitempty"`
	// vSphere cloud-config specs
	VSphereUsername      *string `json:"vSphereUsername,omitempty"`
	VSpherePassword      *string `json:"vSpherePassword,omitempty"`
	VSphereServer        *string `json:"vSphereServer,omitempty"`
	VSphereDatacenter    *string `json:"vSphereDatacenter,omitempty"`
	VSphereResourcePool  *string `json:"vSphereResourcePool,omitempty"`
	VSphereDatastore     *string `json:"vSphereDatastore,omitempty"`
	VSphereCoreDNSServer *string `json:"vSphereCoreDNSServer,omitempty"`
	// Spotinst cloud-config specs
	SpotinstProduct     *string `json:"spotinstProduct,omitempty"`
	SpotinstOrientation *string `json:"spotinstOrientation,omitempty"`
	// Openstack cloud-config options
	Openstack *OpenstackConfiguration `json:"openstack,omitempty"`
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
