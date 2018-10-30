## KubeletConfigSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | KubeletConfigSpec



KubeletConfigSpec defines the kubelet configuration

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
<li><a href="#instancegroupspec-v1alpha2-kops">InstanceGroupSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
allowPrivileged <br /> *boolean*    | AllowPrivileged enables containers to request privileged mode (defaults to false)
anonymousAuth <br /> *boolean*    | AnonymousAuth permits you to control auth to the kubelet api
apiServers <br /> *string*    | APIServers is not used for clusters version 1.6 and later - flag removed
authorizationMode <br /> *string*    | AuthorizationMode is the authorization mode the kubelet is running in
babysitDaemons <br /> *boolean*    | The node has babysitter process monitoring docker and kubelet. Removed as of 1.7
cgroupRoot <br /> *string*    | cgroupRoot is the root cgroup to use for pods. This is handled by the container runtime on a best effort basis.
clientCaFile <br /> *string*    | ClientCAFile is the path to a CA certificate
cloudProvider <br /> *string*    | CloudProvider is the provider for cloud services.
clusterDNS <br /> *string*    | ClusterDNS is the IP address for a cluster DNS server
clusterDomain <br /> *string*    | ClusterDomain is the DNS domain for this cluster
configureCbr0 <br /> *boolean*    | configureCBR0 enables the kubelet to configure cbr0 based on Node.Spec.PodCIDR.
enableCustomMetrics <br /> *boolean*    | Enable gathering custom metrics.
enableDebuggingHandlers <br /> *boolean*    | EnableDebuggingHandlers enables server endpoints for log collection and local running of containers and commands
enforceNodeAllocatable <br /> *string*    | Enforce Allocatable across pods whenever the overall usage across all pods exceeds Allocatable.
evictionHard <br /> *string*    | Comma-delimited list of hard eviction expressions.  For example, 'memory.available<300Mi'.
evictionMaxPodGracePeriod <br /> *integer*    | Maximum allowed grace period (in seconds) to use when terminating pods in response to a soft eviction threshold being met.
evictionMinimumReclaim <br /> *string*    | Comma-delimited list of minimum reclaims (e.g. imagefs.available=2Gi) that describes the minimum amount of resource the kubelet will reclaim when performing a pod eviction if that resource is under pressure.
evictionPressureTransitionPeriod <br /> *[Duration](#duration-v1-meta)*    | Duration for which the kubelet has to wait before transitioning out of an eviction pressure condition.
evictionSoft <br /> *string*    | Comma-delimited list of soft eviction expressions.  For example, 'memory.available<300Mi'.
evictionSoftGracePeriod <br /> *string*    | Comma-delimited list of grace periods for each soft eviction signal.  For example, 'memory.available=30s'.
failSwapOn <br /> *boolean*    | Tells the Kubelet to fail to start if swap is enabled on the node.
featureGates <br /> *object*    | FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
hairpinMode <br /> *string*    | How should the kubelet configure the container bridge for hairpin packets. Setting this flag allows endpoints in a Service to loadbalance back to themselves if they should try to access their own Service. Values:   "promiscuous-bridge": make the container bridge promiscuous.   "hairpin-veth":       set the hairpin flag on container veth interfaces.   "none":               do nothing. Setting --configure-cbr0 to false implies that to achieve hairpin NAT one must set --hairpin-mode=veth-flag, because bridge assumes the existence of a container bridge named cbr0.
hostnameOverride <br /> *string*    | HostnameOverride is the hostname used to identify the kubelet instead of the actual hostname.
imageGCHighThresholdPercent <br /> *integer*    | ImageGCHighThresholdPercent is the percent of disk usage after which image garbage collection is always run.
imageGCLowThresholdPercent <br /> *integer*    | ImageGCLowThresholdPercent is the percent of disk usage before which image garbage collection is never run. Lowest disk usage to garbage collect to.
imagePullProgressDeadline <br /> *[Duration](#duration-v1-meta)*    | ImagePullProgressDeadline is the timeout for image pulls If no pulling progress is made before this deadline, the image pulling will be cancelled. (default 1m0s)
kubeReserved <br /> *object*    | Resource reservation for kubernetes system daemons like the kubelet, container runtime, node problem detector, etc.
kubeReservedCgroup <br /> *string*    | Control group for kube daemons.
kubeconfigPath <br /> *string*    | KubeconfigPath is the path of kubeconfig for the kubelet
kubeletCgroups <br /> *string*    | KubeletCgroups is the absolute name of cgroups to isolate the kubelet in.
logLevel <br /> *integer*    | LogLevel is the logging level of the kubelet
maxPods <br /> *integer*    | MaxPods is the number of pods that can run on this Kubelet.
networkPluginMTU <br /> *integer*    | NetworkPluginMTU is the MTU to be passed to the network plugin, and overrides the default MTU for cases where it cannot be automatically computed (such as IPSEC).
networkPluginName <br /> *string*    | NetworkPluginName is the name of the network plugin to be invoked for various events in kubelet/pod lifecycle
nodeLabels <br /> *object*    | NodeLabels to add when registering the node in the cluster.
nodeStatusUpdateFrequency <br /> *[Duration](#duration-v1-meta)*    | NodeStatusUpdateFrequency Specifies how often kubelet posts node status to master (default 10s) must work with nodeMonitorGracePeriod in KubeControllerManagerConfig.
nonMasqueradeCIDR <br /> *string*    | NonMasqueradeCIDR configures masquerading: traffic to IPs outside this range will use IP masquerade.
nvidiaGPUs <br /> *integer*    | NvidiaGPUs is the number of NVIDIA GPU devices on this node.
podCIDR <br /> *string*    | PodCIDR is the CIDR to use for pod IP addresses, only used in standalone mode. In cluster mode, this is obtained from the master.
podInfraContainerImage <br /> *string*    | PodInfraContainerImage is the image whose network/ipc containers in each pod will use.
podManifestPath <br /> *string*    | config is the path to the config file or directory of files
readOnlyPort <br /> *integer*    | ReadOnlyPort is the port used by the kubelet api for read-only access (default 10255)
reconcileCIDR <br /> *boolean*    | ReconcileCIDR is Reconcile node CIDR with the CIDR specified by the API server. No-op if register-node or configure-cbr0 is false.
registerNode <br /> *boolean*    | RegisterNode enables automatic registration with the apiserver.
registerSchedulable <br /> *boolean*    | registerSchedulable tells the kubelet to register the node as schedulable. No-op if register-node is false.
requireKubeconfig <br /> *boolean*    | RequireKubeconfig indicates a kubeconfig is required
resolvConf <br /> *string*    | ResolverConfig is the resolver configuration file used as the basis for the container DNS resolution configuration."), []
runtimeCgroups <br /> *string*    | Cgroups that container runtime is expected to be isolated in.
runtimeRequestTimeout <br /> *[Duration](#duration-v1-meta)*    | RuntimeRequestTimeout is timeout for runtime requests on - pull, logs, exec and attach
seccompProfileRoot <br /> *string*    | SeccompProfileRoot is the directory path for seccomp profiles.
serializeImagePulls <br /> *boolean*    | // SerializeImagePulls when enabled, tells the Kubelet to pull images one // at a time. We recommend *not* changing the default value on nodes that // run docker daemon with version  < 1.9 or an Aufs storage backend. // Issue #10959 has more details.
systemCgroups <br /> *string*    | SystemCgroups is absolute name of cgroups in which to place all non-kernel processes that are not already in a container. Empty for no container. Rolling back the flag requires a reboot.
systemReserved <br /> *object*    | Capture resource reservation for OS system daemons like sshd, udev, etc.
systemReservedCgroup <br /> *string*    | Parent control group for OS system daemons.
taints <br /> *string array*    | Taints to add when registering a node in the cluster
volumePluginDirectory <br /> *string*    | The full path of the directory in which to search for additional third party volume plugins
volumeStatsAggPeriod <br /> *[Duration](#duration-v1-meta)*    | VolumeStatsAggPeriod is the interval for kubelet to calculate and cache the volume disk usage for all pods and volumes

