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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type KubeletConfigSpec struct {
	// not used for clusters version 1.6 and later
	APIServers string `json:"apiServers,omitempty" flag:"api-servers"`

	// kubeconfigPath is the path to the kubeconfig file with authorization
	// information and API server location
	// kops will only use this for clusters version 1.6 and later
	KubeconfigPath    string `json:"kubeconfigPath,omitempty" flag:"kubeconfig"`
	RequireKubeconfig *bool  `json:"requireKubeconfig,omitempty" flag:"require-kubeconfig"`

	LogLevel *int32 `json:"logLevel,omitempty" flag:"v"`

	// Configuration flags - a subset of https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

	// config is the path to the config file or directory of files
	PodManifestPath string `json:"podManifestPath,omitempty" flag:"pod-manifest-path"`
	//// syncFrequency is the max period between synchronizing running
	//// containers and config
	//SyncFrequency unversioned.Duration `json:"syncFrequency"`
	//// fileCheckFrequency is the duration between checking config files for
	//// new data
	//FileCheckFrequency unversioned.Duration `json:"fileCheckFrequency"`
	//// httpCheckFrequency is the duration between checking http for new data
	//HTTPCheckFrequency unversioned.Duration `json:"httpCheckFrequency"`
	//// manifestURL is the URL for accessing the container manifest
	//ManifestURL string `json:"manifestURL"`
	//// manifestURLHeader is the HTTP header to use when accessing the manifest
	//// URL, with the key separated from the value with a ':', as in 'key:value'
	//ManifestURLHeader string `json:"manifestURLHeader"`
	//// enableServer enables the Kubelet's server
	//EnableServer bool `json:"enableServer"`
	//// address is the IP address for the Kubelet to serve on (set to 0.0.0.0
	//// for all interfaces)
	//Address string `json:"address"`
	//// port is the port for the Kubelet to serve on.
	//Port uint `json:"port"`
	//// readOnlyPort is the read-only port for the Kubelet to serve on with
	//// no authentication/authorization (set to 0 to disable)
	//ReadOnlyPort uint `json:"readOnlyPort"`
	//// tLSCertFile is the file containing x509 Certificate for HTTPS.  (CA cert,
	//// if any, concatenated after server cert). If tlsCertFile and
	//// tlsPrivateKeyFile are not provided, a self-signed certificate
	//// and key are generated for the public address and saved to the directory
	//// passed to certDir.
	//TLSCertFile string `json:"tlsCertFile"`
	//// tLSPrivateKeyFile is the ile containing x509 private key matching
	//// tlsCertFile.
	//TLSPrivateKeyFile string `json:"tlsPrivateKeyFile"`
	//// certDirectory is the directory where the TLS certs are located (by
	//// default /var/run/kubernetes). If tlsCertFile and tlsPrivateKeyFile
	//// are provided, this flag will be ignored.
	//CertDirectory string `json:"certDirectory"`
	// hostnameOverride is the hostname used to identify the kubelet instead
	// of the actual hostname.
	// Note: We recognize some additional values:
	//  @aws uses the hostname from the AWS metadata service
	HostnameOverride string `json:"hostnameOverride,omitempty" flag:"hostname-override"`
	//// podInfraContainerImage is the image whose network/ipc namespaces
	//// containers in each pod will use.
	//PodInfraContainerImage string `json:"podInfraContainerImage"`
	//// dockerEndpoint is the path to the docker endpoint to communicate with.
	//DockerEndpoint string `json:"dockerEndpoint"`
	//// rootDirectory is the directory path to place kubelet files (volume
	//// mounts,etc).
	//RootDirectory string `json:"rootDirectory"`
	//// seccompProfileRoot is the directory path for seccomp profiles.
	//SeccompProfileRoot string `json:"seccompProfileRoot"`
	// allowPrivileged enables containers to request privileged mode.
	// Defaults to false.
	AllowPrivileged *bool `json:"allowPrivileged,omitempty" flag:"allow-privileged"`
	//// hostNetworkSources is a comma-separated list of sources from which the
	//// Kubelet allows pods to use of host network. Defaults to "*".
	//HostNetworkSources []string `json:"hostNetworkSources"`
	//// hostPIDSources is a comma-separated list of sources from which the
	//// Kubelet allows pods to use the host pid namespace. Defaults to "*".
	//HostPIDSources []string `json:"hostPIDSources"`
	//// hostIPCSources is a comma-separated list of sources from which the
	//// Kubelet allows pods to use the host ipc namespace. Defaults to "*".
	//HostIPCSources []string `json:"hostIPCSources"`
	//// registryPullQPS is the limit of registry pulls per second. If 0,
	//// unlimited. Set to 0 for no limit. Defaults to 5.0.
	//RegistryPullQPS float64 `json:"registryPullQPS"`
	//// registryBurst is the maximum size of a bursty pulls, temporarily allows
	//// pulls to burst to this number, while still not exceeding registryQps.
	//// Only used if registryQps > 0.
	//RegistryBurst int32 `json:"registryBurst"`
	//// eventRecordQPS is the maximum event creations per second. If 0, there
	//// is no limit enforced.
	//EventRecordQPS float32 `json:"eventRecordQPS"`
	//// eventBurst is the maximum size of a bursty event records, temporarily
	//// allows event records to burst to this number, while still not exceeding
	//// event-qps. Only used if eventQps > 0
	//EventBurst int32 `json:"eventBurst"`
	// enableDebuggingHandlers enables server endpoints for log collection
	// and local running of containers and commands
	EnableDebuggingHandlers *bool `json:"enableDebuggingHandlers,omitempty" flag:"enable-debugging-handlers"`
	//// minimumGCAge is the minimum age for a finished container before it is
	//// garbage collected.
	//MinimumGCAge unversioned.Duration `json:"minimumGCAge"`
	//// maxPerPodContainerCount is the maximum number of old instances to
	//// retain per container. Each container takes up some disk space.
	//MaxPerPodContainerCount int32 `json:"maxPerPodContainerCount"`
	//// maxContainerCount is the maximum number of old instances of containers
	//// to retain globally. Each container takes up some disk space.
	//MaxContainerCount int32 `json:"maxContainerCount"`
	//// cAdvisorPort is the port of the localhost cAdvisor endpoint
	//CAdvisorPort uint `json:"cAdvisorPort"`
	//// healthzPort is the port of the localhost healthz endpoint
	//HealthzPort int32 `json:"healthzPort"`
	//// healthzBindAddress is the IP address for the healthz server to serve
	//// on.
	//HealthzBindAddress string `json:"healthzBindAddress"`
	//// oomScoreAdj is The oom-score-adj value for kubelet process. Values
	//// must be within the range [-1000, 1000].
	//OOMScoreAdj int32 `json:"oomScoreAdj"`
	//// registerNode enables automatic registration with the apiserver.
	//RegisterNode bool `json:"registerNode"`
	// clusterDomain is the DNS domain for this cluster. If set, kubelet will
	// configure all containers to search this domain in addition to the
	// host's search domains.
	ClusterDomain string `json:"clusterDomain,omitempty" flag:"cluster-domain"`
	//// masterServiceNamespace is The namespace from which the kubernetes
	//// master services should be injected into pods.
	//MasterServiceNamespace string `json:"masterServiceNamespace"`
	// clusterDNS is the IP address for a cluster DNS server.  If set, kubelet
	// will configure all containers to use this for DNS resolution in
	// addition to the host's DNS servers
	ClusterDNS string `json:"clusterDNS,omitempty" flag:"cluster-dns"`
	//// streamingConnectionIdleTimeout is the maximum time a streaming connection
	//// can be idle before the connection is automatically closed.
	//StreamingConnectionIdleTimeout unversioned.Duration `json:"streamingConnectionIdleTimeout"`
	//// nodeStatusUpdateFrequency is the frequency that kubelet posts node
	//// status to master. Note: be cautious when changing the constant, it
	//// must work with nodeMonitorGracePeriod in nodecontroller.
	//NodeStatusUpdateFrequency unversioned.Duration `json:"nodeStatusUpdateFrequency"`
	//// minimumGCAge is the minimum age for a unused image before it is
	//// garbage collected.
	//ImageMinimumGCAge unversioned.Duration `json:"imageMinimumGCAge"`
	//// lowDiskSpaceThresholdMB is the absolute free disk space, in MB, to
	//// maintain. When disk space falls below this threshold, new pods would
	//// be rejected.
	//LowDiskSpaceThresholdMB int32 `json:"lowDiskSpaceThresholdMB"`
	//// How frequently to calculate and cache volume disk usage for all pods
	//VolumeStatsAggPeriod unversioned.Duration `json:"volumeStatsAggPeriod"`
	// networkPluginName is the name of the network plugin to be invoked for
	// various events in kubelet/pod lifecycle
	NetworkPluginName string `json:"networkPluginName,omitempty" flag:"network-plugin"`
	//// networkPluginDir is the full path of the directory in which to search
	//// for network plugins
	//NetworkPluginDir string `json:"networkPluginDir"`
	//// volumePluginDir is the full path of the directory in which to search
	//// for additional third party volume plugins
	//VolumePluginDir string `json:"volumePluginDir"`
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
	//// containerRuntime is the container runtime to use.
	//ContainerRuntime string `json:"containerRuntime"`
	//// rktPath is the path of rkt binary. Leave empty to use the first rkt in
	//// $PATH.
	//RktPath string `json:"rktPath,omitempty"`
	//// rktApiEndpoint is the endpoint of the rkt API service to communicate with.
	//RktAPIEndpoint string `json:"rktAPIEndpoint,omitempty"`
	//// rktStage1Image is the image to use as stage1. Local paths and
	//// http/https URLs are supported.
	//RktStage1Image string `json:"rktStage1Image,omitempty"`
	//// lockFilePath is the path that kubelet will use to as a lock file.
	//// It uses this file as a lock to synchronize with other kubelet processes
	//// that may be running.
	//LockFilePath string `json:"lockFilePath"`
	//// ExitOnLockContention is a flag that signifies to the kubelet that it is running
	//// in "bootstrap" mode. This requires that 'LockFilePath' has been set.
	//// This will cause the kubelet to listen to inotify events on the lock file,
	//// releasing it and exiting when another process tries to open that file.
	//ExitOnLockContention bool `json:"exitOnLockContention"`
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

	// maxPods is the number of pods that can run on this Kubelet.
	MaxPods *int32 `json:"maxPods,omitempty" flag:"max-pods"`

	// nvidiaGPUs is the number of NVIDIA GPU devices on this node.
	NvidiaGPUs int32 `json:"nvidiaGPUs,omitempty" flag:"experimental-nvidia-gpus"`

	//// dockerExecHandlerName is the handler to use when executing a command
	//// in a container. Valid values are 'native' and 'nsenter'. Defaults to
	//// 'native'.
	//DockerExecHandlerName string `json:"dockerExecHandlerName"`
	// The CIDR to use for pod IP addresses, only used in standalone mode.
	// In cluster mode, this is obtained from the master.
	PodCIDR string `json:"podCIDR,omitempty" flag:"pod-cidr"`
	//// ResolverConfig is the resolver configuration file used as the basis
	//// for the container DNS resolution configuration."), []
	//ResolverConfig string `json:"resolvConf"`
	//// cpuCFSQuota is Enable CPU CFS quota enforcement for containers that
	//// specify CPU limits
	//CPUCFSQuota bool `json:"cpuCFSQuota"`
	//// containerized should be set to true if kubelet is running in a container.
	//Containerized bool `json:"containerized"`
	//// maxOpenFiles is Number of files that can be opened by Kubelet process.
	//MaxOpenFiles uint64 `json:"maxOpenFiles"`
	// reconcileCIDR is Reconcile node CIDR with the CIDR specified by the
	// API server. No-op if register-node or configure-cbr0 is false.
	ReconcileCIDR *bool `json:"reconcileCIDR,omitempty" flag:"reconcile-cidr"`
	// registerSchedulable tells the kubelet to register the node as
	// schedulable. No-op if register-node is false.
	RegisterSchedulable *bool `json:"registerSchedulable,omitempty" flag:"register-schedulable"`
	//// contentType is contentType of requests sent to apiserver.
	//ContentType string `json:"contentType"`
	//// kubeAPIQPS is the QPS to use while talking with kubernetes apiserver
	//KubeAPIQPS float32 `json:"kubeAPIQPS"`
	//// kubeAPIBurst is the burst to allow while talking with kubernetes
	//// apiserver
	//KubeAPIBurst int32 `json:"kubeAPIBurst"`
	//// serializeImagePulls when enabled, tells the Kubelet to pull images one
	//// at a time. We recommend *not* changing the default value on nodes that
	//// run docker daemon with version  < 1.9 or an Aufs storage backend.
	//// Issue #10959 has more details.
	//SerializeImagePulls bool `json:"serializeImagePulls"`
	//// experimentalFlannelOverlay enables experimental support for starting the
	//// kubelet with the default overlay network (flannel). Assumes flanneld
	//// is already running in client mode.
	//ExperimentalFlannelOverlay bool `json:"experimentalFlannelOverlay"`
	//// outOfDiskTransitionFrequency is duration for which the kubelet has to
	//// wait before transitioning out of out-of-disk node condition status.
	//OutOfDiskTransitionFrequency unversioned.Duration `json:"outOfDiskTransitionFrequency,omitempty"`
	//// nodeIP is IP address of the node. If set, kubelet will use this IP
	//// address for the node.
	//NodeIP string `json:"nodeIP,omitempty"`
	// nodeLabels to add when registering the node in the cluster.
	NodeLabels map[string]string `json:"nodeLabels,omitempty" flag:"node-labels"`
	// nonMasqueradeCIDR configures masquerading: traffic to IPs outside this range will use IP masquerade.
	NonMasqueradeCIDR string `json:"nonMasqueradeCIDR,omitempty" flag:"non-masquerade-cidr"`

	// enable gathering custom metrics.
	EnableCustomMetrics *bool `json:"enableCustomMetrics,omitempty" flag:"enable-custom-metrics"`

	//// Maximum number of pods per core. Cannot exceed MaxPods
	//PodsPerCore int32 `json:"podsPerCore"`
	//// enableControllerAttachDetach enables the Attach/Detach controller to
	//// manage attachment/detachment of volumes scheduled to this node, and
	//// disables kubelet from executing any attach/detach operations
	//EnableControllerAttachDetach bool `json:"enableControllerAttachDetach"`

	// networkPluginMTU is the MTU to be passed to the network plugin,
	// and overrides the default MTU for cases where it cannot be automatically
	// computed (such as IPSEC).
	NetworkPluginMTU *int32 `json:"networkPluginMTU,omitempty" flag:"network-plugin-mtu"`

	// imageGCHighThresholdPercent is the percent of disk usage after which
	// image garbage collection is always run.
	ImageGCHighThresholdPercent *int32 `json:"imageGCHighThresholdPercent,omitempty" flag:"image-gc-high-threshold"`
	// imageGCLowThresholdPercent is the percent of disk usage before which
	// image garbage collection is never run. Lowest disk usage to garbage
	// collect to.
	ImageGCLowThresholdPercent *int32 `json:"imageGCLowThresholdPercent,omitempty" flag:"image-gc-low-threshold"`

	// Comma-delimited list of hard eviction expressions.  For example, 'memory.available<300Mi'.
	EvictionHard *string `json:"evictionHard,omitempty" flag:"eviction-hard"`
	// Comma-delimited list of soft eviction expressions.  For example, 'memory.available<300Mi'.
	EvictionSoft string `json:"evictionSoft,omitempty" flag:"eviction-soft"`
	// Comma-delimited list of grace periods for each soft eviction signal.  For example, 'memory.available=30s'.
	EvictionSoftGracePeriod string `json:"evictionSoftGracePeriod,omitempty" flag:"eviction-soft-grace-period"`
	// Duration for which the kubelet has to wait before transitioning out of an eviction pressure condition.
	EvictionPressureTransitionPeriod *metav1.Duration `json:"evictionPressureTransitionPeriod,omitempty" flag:"eviction-pressure-transition-period"`
	// Maximum allowed grace period (in seconds) to use when terminating pods in response to a soft eviction threshold being met.
	EvictionMaxPodGracePeriod int32 `json:"evictionMaxPodGracePeriod,omitempty" flag:"eviction-max-pod-grace-period"`
	// Comma-delimited list of minimum reclaims (e.g. imagefs.available=2Gi) that describes the minimum amount of resource the kubelet will reclaim when performing a pod eviction if that resource is under pressure.
	EvictionMinimumReclaim string `json:"evictionMinimumReclaim,omitempty" flag:"eviction-minimum-reclaim"`

	// The full path of the directory in which to search for additional third party volume plugins
	VolumePluginDirectory string `json:"volumePluginDirectory,omitempty" flag:"volume-plugin-dir"`
}

type KubeProxyConfig struct {
	Image string `json:"image,omitempty"`
	// TODO: Better type ?
	CPURequest string `json:"cpuRequest,omitempty"` // e.g. "20m"

	LogLevel int32 `json:"logLevel,omitempty" flag:"v"`

	// Configuration flags - a subset of https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

	//// bindAddress is the IP address for the proxy server to serve on (set to 0.0.0.0
	//// for all interfaces)
	//BindAddress string `json:"bindAddress"`
	//// clusterCIDR is the CIDR range of the pods in the cluster. It is used to
	//// bridge traffic coming from outside of the cluster. If not provided,
	//// no off-cluster bridging will be performed.
	//ClusterCIDR string `json:"clusterCIDR"`
	//// healthzBindAddress is the IP address for the health check server to serve on,
	//// defaulting to 127.0.0.1 (set to 0.0.0.0 for all interfaces)
	//HealthzBindAddress string `json:"healthzBindAddress"`
	//// healthzPort is the port to bind the health check server. Use 0 to disable.
	//HealthzPort int32 `json:"healthzPort"`
	//// hostnameOverride, if non-empty, will be used as the identity instead of the actual hostname.
	//HostnameOverride string `json:"hostnameOverride"`
	//// iptablesMasqueradeBit is the bit of the iptables fwmark space to use for SNAT if using
	//// the pure iptables proxy mode. Values must be within the range [0, 31].
	//IPTablesMasqueradeBit *int32 `json:"iptablesMasqueradeBit"`
	//// iptablesSyncPeriod is the period that iptables rules are refreshed (e.g. '5s', '1m',
	//// '2h22m').  Must be greater than 0.
	//IPTablesSyncPeriod unversioned.Duration `json:"iptablesSyncPeriodSeconds"`
	//// kubeconfigPath is the path to the kubeconfig file with authorization information (the
	//// master location is set by the master flag).
	//KubeconfigPath string `json:"kubeconfigPath"`
	//// masqueradeAll tells kube-proxy to SNAT everything if using the pure iptables proxy mode.
	//MasqueradeAll bool `json:"masqueradeAll"`
	// master is the address of the Kubernetes API server (overrides any value in kubeconfig)
	Master string `json:"master,omitempty" flag:"master"`
	//// oomScoreAdj is the oom-score-adj value for kube-proxy process. Values must be within
	//// the range [-1000, 1000]
	//OOMScoreAdj *int32 `json:"oomScoreAdj"`
	//// mode specifies which proxy mode to use.
	//Mode ProxyMode `json:"mode"`
	//// portRange is the range of host ports (beginPort-endPort, inclusive) that may be consumed
	//// in order to proxy service traffic. If unspecified (0-0) then ports will be randomly chosen.
	//PortRange string `json:"portRange"`
	//// resourceContainer is the bsolute name of the resource-only container to create and run
	//// the Kube-proxy in (Default: /kube-proxy).
	//ResourceContainer string `json:"resourceContainer"`
	//// udpIdleTimeout is how long an idle UDP connection will be kept open (e.g. '250ms', '2s').
	//// Must be greater than 0. Only applicable for proxyMode=userspace.
	//UDPIdleTimeout unversioned.Duration `json:"udpTimeoutMilliseconds"`
	//// conntrackMax is the maximum number of NAT connections to track (0 to leave as-is)")
	//ConntrackMax int32 `json:"conntrackMax"`
	//// conntrackTCPEstablishedTimeout is how long an idle UDP connection will be kept open
	//// (e.g. '250ms', '2s').  Must be greater than 0. Only applicable for proxyMode is Userspace
	//ConntrackTCPEstablishedTimeout unversioned.Duration `json:"conntrackTCPEstablishedTimeout"`
}

type KubeAPIServerConfig struct {
	PathSrvKubernetes string `json:"pathSrvKubernetes,omitempty"`
	PathSrvSshproxy   string `json:"pathSrvSshproxy,omitempty"`
	Image             string `json:"image,omitempty"`

	LogLevel int32 `json:"logLevel,omitempty" flag:"v"`

	CloudProvider         string            `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	SecurePort            int32             `json:"securePort,omitempty" flag:"secure-port"`
	Address               string            `json:"address,omitempty" flag:"address"`
	EtcdServers           []string          `json:"etcdServers,omitempty" flag:"etcd-servers"`
	EtcdServersOverrides  []string          `json:"etcdServersOverrides,omitempty" flag:"etcd-servers-overrides"`
	AdmissionControl      []string          `json:"admissionControl,omitempty" flag:"admission-control"`
	ServiceClusterIPRange string            `json:"serviceClusterIPRange,omitempty" flag:"service-cluster-ip-range"`
	ClientCAFile          string            `json:"clientCAFile,omitempty" flag:"client-ca-file"`
	BasicAuthFile         string            `json:"basicAuthFile,omitempty" flag:"basic-auth-file"`
	TLSCertFile           string            `json:"tlsCertFile,omitempty" flag:"tls-cert-file"`
	TLSPrivateKeyFile     string            `json:"tlsPrivateKeyFile,omitempty" flag:"tls-private-key-file"`
	TokenAuthFile         string            `json:"tokenAuthFile,omitempty" flag:"token-auth-file"`
	AllowPrivileged       *bool             `json:"allowPrivileged,omitempty" flag:"allow-privileged"`
	APIServerCount        *int32            `json:"apiServerCount,omitempty" flag:"apiserver-count"`
	RuntimeConfig         map[string]string `json:"runtimeConfig,omitempty" flag:"runtime-config"`

	AnonymousAuth *bool `json:"anonymousAuth,omitempty" flag:"anonymous-auth"`

	KubeletPreferredAddressTypes []string `json:"kubeletPreferredAddressTypes,omitempty" flag:"kubelet-preferred-address-types"`

	StorageBackend *string `json:"storageBackend,omitempty" flag:"storage-backend"`

	// The OpenID claim to use as the user name.
	// Note that claims other than the default ('sub') is not guaranteed to be unique and immutable.
	OIDCUsernameClaim *string `json:"oidcUsernameClaim,omitempty" flag:"oidc-username-claim"`
	// If provided, the name of a custom OpenID Connect claim for specifying user groups.
	// The claim value is expected to be a string or array of strings.
	OIDCGroupsClaim *string `json:"oidcGroupsClaim,omitempty" flag:"oidc-groups-claim"`
	// The URL of the OpenID issuer, only HTTPS scheme will be accepted.
	// If set, it will be used to verify the OIDC JSON Web Token (JWT).
	OIDCIssuerURL *string `json:"oidcIssuerURL,omitempty" flag:"oidc-issuer-url"`
	// The client ID for the OpenID Connect client, must be set if oidc-issuer-url is set.
	OIDCClientID *string `json:"oidcClientID,omitempty" flag:"oidc-client-id"`
	// If set, the OpenID server's certificate will be verified by one of the authorities in the oidc-ca-file
	// otherwise the host's root CA set will be used.
	OIDCCAFile *string `json:"oidcCAFile,omitempty" flag:"oidc-ca-file"`

	// If set, all requests coming to the apiserver will be logged to this file.
	AuditLogPath *string `json:"auditLogPath,omitempty" flag:"audit-log-path"`
	// The maximum number of days to retain old audit log files based on the timestamp encoded in their filename.
	AuditLogMaxAge *int32 `json:"auditLogMaxAge,omitempty" flag:"audit-log-maxage"`
	// The maximum number of old audit log files to retain.
	AuditLogMaxBackups *int32 `json:"auditLogMaxBackups,omitempty" flag:"audit-log-maxbackup"`
	// The maximum size in megabytes of the audit log file before it gets rotated. Defaults to 100MB.
	AuditLogMaxSize *int32 `json:"auditLogMaxSize,omitempty" flag:"audit-log-maxsize"`

	AuthorizationMode          *string `json:"authorizationMode,omitempty" flag:"authorization-mode"`
	AuthorizationRBACSuperUser *string `json:"authorizationRbacSuperUser,omitempty" flag:"authorization-rbac-super-user"`
}

type KubeControllerManagerConfig struct {
	Master   string `json:"master,omitempty" flag:"master"`
	LogLevel int32  `json:"logLevel,omitempty" flag:"v"`

	ServiceAccountPrivateKeyFile string `json:"serviceAccountPrivateKeyFile,omitempty" flag:"service-account-private-key-file"`

	Image string `json:"image,omitempty"`

	PathSrvKubernetes string `json:"pathSrvKubernetes,omitempty"`

	// Configuration flags - a subset of https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

	//// port is the port that the controller-manager's http service runs on.
	//Port int32 `json:"port"`
	//// address is the IP address to serve on (set to 0.0.0.0 for all interfaces).
	//Address string `json:"address"`
	// cloudProvider is the provider for cloud services.
	CloudProvider string `json:"cloudProvider,omitempty" flag:"cloud-provider"`
	//// concurrentEndpointSyncs is the number of endpoint syncing operations
	//// that will be done concurrently. Larger number = faster endpoint updating,
	//// but more CPU (and network) load.
	//ConcurrentEndpointSyncs int32 `json:"concurrentEndpointSyncs"`
	//// concurrentRSSyncs is the number of replica sets that are  allowed to sync
	//// concurrently. Larger number = more responsive replica  management, but more
	//// CPU (and network) load.
	//ConcurrentRSSyncs int32 `json:"concurrentRSSyncs"`
	//// concurrentRCSyncs is the number of replication controllers that are
	//// allowed to sync concurrently. Larger number = more responsive replica
	//// management, but more CPU (and network) load.
	//ConcurrentRCSyncs int32 `json:"concurrentRCSyncs"`
	//// concurrentResourceQuotaSyncs is the number of resource quotas that are
	//// allowed to sync concurrently. Larger number = more responsive quota
	//// management, but more CPU (and network) load.
	//ConcurrentResourceQuotaSyncs int32 `json:"concurrentResourceQuotaSyncs"`
	//// concurrentDeploymentSyncs is the number of deployment objects that are
	//// allowed to sync concurrently. Larger number = more responsive deployments,
	//// but more CPU (and network) load.
	//ConcurrentDeploymentSyncs int32 `json:"concurrentDeploymentSyncs"`
	//// concurrentDaemonSetSyncs is the number of daemonset objects that are
	//// allowed to sync concurrently. Larger number = more responsive daemonset,
	//// but more CPU (and network) load.
	//ConcurrentDaemonSetSyncs int32 `json:"concurrentDaemonSetSyncs"`
	//// concurrentJobSyncs is the number of job objects that are
	//// allowed to sync concurrently. Larger number = more responsive jobs,
	//// but more CPU (and network) load.
	//ConcurrentJobSyncs int32 `json:"concurrentJobSyncs"`
	//// concurrentNamespaceSyncs is the number of namespace objects that are
	//// allowed to sync concurrently.
	//ConcurrentNamespaceSyncs int32 `json:"concurrentNamespaceSyncs"`
	//// lookupCacheSizeForRC is the size of lookup cache for replication controllers.
	//// Larger number = more responsive replica management, but more MEM load.
	//LookupCacheSizeForRC int32 `json:"lookupCacheSizeForRC"`
	//// lookupCacheSizeForRS is the size of lookup cache for replicatsets.
	//// Larger number = more responsive replica management, but more MEM load.
	//LookupCacheSizeForRS int32 `json:"lookupCacheSizeForRS"`
	//// lookupCacheSizeForDaemonSet is the size of lookup cache for daemonsets.
	//// Larger number = more responsive daemonset, but more MEM load.
	//LookupCacheSizeForDaemonSet int32 `json:"lookupCacheSizeForDaemonSet"`
	//// serviceSyncPeriod is the period for syncing services with their external
	//// load balancers.
	//ServiceSyncPeriod unversioned.Duration `json:"serviceSyncPeriod"`
	//// nodeSyncPeriod is the period for syncing nodes from cloudprovider. Longer
	//// periods will result in fewer calls to cloud provider, but may delay addition
	//// of new nodes to cluster.
	//NodeSyncPeriod unversioned.Duration `json:"nodeSyncPeriod"`
	//// resourceQuotaSyncPeriod is the period for syncing quota usage status
	//// in the system.
	//ResourceQuotaSyncPeriod unversioned.Duration `json:"resourceQuotaSyncPeriod"`
	//// namespaceSyncPeriod is the period for syncing namespace life-cycle
	//// updates.
	//NamespaceSyncPeriod unversioned.Duration `json:"namespaceSyncPeriod"`
	//// pvClaimBinderSyncPeriod is the period for syncing persistent volumes
	//// and persistent volume claims.
	//PVClaimBinderSyncPeriod unversioned.Duration `json:"pvClaimBinderSyncPeriod"`
	//// minResyncPeriod is the resync period in reflectors; will be random between
	//// minResyncPeriod and 2*minResyncPeriod.
	//MinResyncPeriod unversioned.Duration `json:"minResyncPeriod"`
	//// horizontalPodAutoscalerSyncPeriod is the period for syncing the number of
	//// pods in horizontal pod autoscaler.
	//HorizontalPodAutoscalerSyncPeriod unversioned.Duration `json:"horizontalPodAutoscalerSyncPeriod"`
	//// deploymentControllerSyncPeriod is the period for syncing the deployments.
	//DeploymentControllerSyncPeriod unversioned.Duration `json:"deploymentControllerSyncPeriod"`
	//// podEvictionTimeout is the grace period for deleting pods on failed nodes.
	//PodEvictionTimeout unversioned.Duration `json:"podEvictionTimeout"`
	//// deletingPodsQps is the number of nodes per second on which pods are deleted in
	//// case of node failure.
	//DeletingPodsQps float32 `json:"deletingPodsQps"`
	//// deletingPodsBurst is the number of nodes on which pods are bursty deleted in
	//// case of node failure. For more details look into RateLimiter.
	//DeletingPodsBurst int32 `json:"deletingPodsBurst"`
	//// nodeMontiorGracePeriod is the amount of time which we allow a running node to be
	//// unresponsive before marking it unhealthy. Must be N times more than kubelet's
	//// nodeStatusUpdateFrequency, where N means number of retries allowed for kubelet
	//// to post node status.
	//NodeMonitorGracePeriod unversioned.Duration `json:"nodeMonitorGracePeriod"`
	//// registerRetryCount is the number of retries for initial node registration.
	//// Retry interval equals node-sync-period.
	//RegisterRetryCount int32 `json:"registerRetryCount"`
	//// nodeStartupGracePeriod is the amount of time which we allow starting a node to
	//// be unresponsive before marking it unhealthy.
	//NodeStartupGracePeriod unversioned.Duration `json:"nodeStartupGracePeriod"`
	//// nodeMonitorPeriod is the period for syncing NodeStatus in NodeController.
	//NodeMonitorPeriod unversioned.Duration `json:"nodeMonitorPeriod"`
	//// serviceAccountKeyFile is the filename containing a PEM-encoded private RSA key
	//// used to sign service account tokens.
	//ServiceAccountKeyFile string `json:"serviceAccountKeyFile"`
	//// enableProfiling enables profiling via web interface host:port/debug/pprof/
	//EnableProfiling bool `json:"enableProfiling"`
	// clusterName is the instance prefix for the cluster.
	ClusterName string `json:"clusterName,omitempty" flag:"cluster-name"`
	// clusterCIDR is CIDR Range for Pods in cluster.
	ClusterCIDR string `json:"clusterCIDR,omitempty" flag:"cluster-cidr"`
	//// serviceCIDR is CIDR Range for Services in cluster.
	//ServiceCIDR string `json:"serviceCIDR"`
	//// NodeCIDRMaskSize is the mask size for node cidr in cluster.
	//NodeCIDRMaskSize int32 `json:"nodeCIDRMaskSize"`
	// allocateNodeCIDRs enables CIDRs for Pods to be allocated and, if
	// ConfigureCloudRoutes is true, to be set on the cloud provider.
	AllocateNodeCIDRs *bool `json:"allocateNodeCIDRs,omitempty" flag:"allocate-node-cidrs"`
	// configureCloudRoutes enables CIDRs allocated with allocateNodeCIDRs
	// to be configured on the cloud provider.
	ConfigureCloudRoutes *bool `json:"configureCloudRoutes,omitempty" flag:"configure-cloud-routes"`
	// rootCAFile is the root certificate authority will be included in service
	// account's token secret. This must be a valid PEM-encoded CA bundle.
	RootCAFile string `json:"rootCAFile,omitempty" flag:"root-ca-file"`
	//// contentType is contentType of requests sent to apiserver.
	//ContentType string `json:"contentType"`
	//// kubeAPIQPS is the QPS to use while talking with kubernetes apiserver.
	//KubeAPIQPS float32 `json:"kubeAPIQPS"`
	//// kubeAPIBurst is the burst to use while talking with kubernetes apiserver.
	//KubeAPIBurst int32 `json:"kubeAPIBurst"`
	// leaderElection defines the configuration of leader election client.
	LeaderElection *LeaderElectionConfiguration `json:"leaderElection,omitempty"`
	//// volumeConfiguration holds configuration for volume related features.
	//VolumeConfiguration VolumeConfiguration `json:"volumeConfiguration"`
	//// How long to wait between starting controller managers
	//ControllerStartInterval unversioned.Duration `json:"controllerStartInterval"`
	//// enables the generic garbage collector. MUST be synced with the
	//// corresponding flag of the kube-apiserver. WARNING: the generic garbage
	//// collector is an alpha feature.
	//EnableGarbageCollector bool `json:"enableGarbageCollector"`

	// ReconcilerSyncLoopPeriod is the amount of time the reconciler sync states loop
	// wait between successive executions. Is set to 1 min by kops by default
	AttachDetachReconcileSyncPeriod *metav1.Duration `json:"attachDetachReconcileSyncPeriod,omitempty" flag:"attach-detach-reconcile-sync-period"`

	// terminatedPodGCThreshold is the number of terminated pods that can exist
	// before the terminated pod garbage collector starts deleting terminated pods.
	// If <= 0, the terminated pod garbage collector is disabled.
	TerminatedPodGCThreshold *int32 `json:"terminatedPodGCThreshold,omitempty" flag:"terminated-pod-gc-threshold"`
}

type KubeSchedulerConfig struct {
	Master   string `json:"master,omitempty" flag:"master"`
	LogLevel int32  `json:"logLevel,omitempty" flag:"v"`

	Image string `json:"image,omitempty"`

	// Configuration flags - a subset of https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/componentconfig/types.go

	//// port is the port that the scheduler's http service runs on.
	//Port int32 `json:"port"`
	//// address is the IP address to serve on.
	//Address string `json:"address"`
	//// algorithmProvider is the scheduling algorithm provider to use.
	//AlgorithmProvider string `json:"algorithmProvider"`
	//// policyConfigFile is the filepath to the scheduler policy configuration.
	//PolicyConfigFile string `json:"policyConfigFile"`
	//// enableProfiling enables profiling via web interface.
	//EnableProfiling bool `json:"enableProfiling"`
	//// contentType is contentType of requests sent to apiserver.
	//ContentType string `json:"contentType"`
	//// kubeAPIQPS is the QPS to use while talking with kubernetes apiserver.
	//KubeAPIQPS float32 `json:"kubeAPIQPS"`
	//// kubeAPIBurst is the QPS burst to use while talking with kubernetes apiserver.
	//KubeAPIBurst int32 `json:"kubeAPIBurst"`
	//// schedulerName is name of the scheduler, used to select which pods
	//// will be processed by this scheduler, based on pod's annotation with
	//// key 'scheduler.alpha.kubernetes.io/name'.
	//SchedulerName string `json:"schedulerName"`
	//// RequiredDuringScheduling affinity is not symmetric, but there is an implicit PreferredDuringScheduling affinity rule
	//// corresponding to every RequiredDuringScheduling affinity rule.
	//// HardPodAffinitySymmetricWeight represents the weight of implicit PreferredDuringScheduling affinity rule, in the range 0-100.
	//HardPodAffinitySymmetricWeight int `json:"hardPodAffinitySymmetricWeight"`
	//// Indicate the "all topologies" set for empty topologyKey when it's used for PreferredDuringScheduling pod anti-affinity.
	//FailureDomains string `json:"failureDomains"`
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
	//// leaseDuration is the duration that non-leader candidates will wait
	//// after observing a leadership renewal until attempting to acquire
	//// leadership of a led but unrenewed leader slot. This is effectively the
	//// maximum duration that a leader can be stopped before it is replaced
	//// by another candidate. This is only applicable if leader election is
	//// enabled.
	//LeaseDuration unversioned.Duration `json:"leaseDuration"`
	//// renewDeadline is the interval between attempts by the acting master to
	//// renew a leadership slot before it stops leading. This must be less
	//// than or equal to the lease duration. This is only applicable if leader
	//// election is enabled.
	//RenewDeadline unversioned.Duration `json:"renewDeadline"`
	//// retryPeriod is the duration the clients should wait between attempting
	//// acquisition and renewal of a leadership. This is only applicable if
	//// leader election is enabled.
	//RetryPeriod unversioned.Duration `json:"retryPeriod"`
}

type CloudConfiguration struct {
	// GCE cloud-config options
	Multizone          *bool   `json:"multizone,omitempty"`
	NodeTags           *string `json:"nodeTags,omitempty"`
	NodeInstancePrefix *string `json:"nodeInstancePrefix,omitempty"`
}
