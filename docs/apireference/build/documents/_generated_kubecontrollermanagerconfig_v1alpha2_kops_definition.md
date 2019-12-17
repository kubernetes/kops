## KubeControllerManagerConfig v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | KubeControllerManagerConfig



KubeControllerManagerConfig is the configuration for the controller

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
allocateNodeCIDRs <br /> *boolean*    | AllocateNodeCIDRs enables CIDRs for Pods to be allocated and, if ConfigureCloudRoutes is true, to be set on the cloud provider.
attachDetachReconcileSyncPeriod <br /> *[Duration](#duration-v1-meta)*    | ReconcilerSyncLoopPeriod is the amount of time the reconciler sync states loop wait between successive executions. Is set to 1 min by kops by default
cidrAllocatorType <br /> *string*    | CIDRAllocatorType specifies the type of CIDR allocator to use.
cloudProvider <br /> *string*    | CloudProvider is the provider for cloud services.
clusterCIDR <br /> *string*    | ClusterCIDR is CIDR Range for Pods in cluster.
clusterName <br /> *string*    | ClusterName is the instance prefix for the cluster.
configureCloudRoutes <br /> *boolean*    | ConfigureCloudRoutes enables CIDRs allocated with to be configured on the cloud provider.
featureGates <br /> *object*    | FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
horizontalPodAutoscalerDownscaleDelay <br /> *[Duration](#duration-v1-meta)*    | HorizontalPodAutoscalerDownscaleDelay is a duration that specifies how long the autoscaler has to wait before another downscale operation can be performed after the current one has completed.
horizontalPodAutoscalerSyncPeriod <br /> *[Duration](#duration-v1-meta)*    | HorizontalPodAutoscalerSyncPeriod is the amount of time between syncs During each period, the controller manager queries the resource utilization against the metrics specified in each HorizontalPodAutoscaler definition.
horizontalPodAutoscalerUpscaleDelay <br /> *[Duration](#duration-v1-meta)*    | HorizontalPodAutoscalerUpscaleDelay is a duration that specifies how long the autoscaler has to wait before another upscale operation can be performed after the current one has completed.
horizontalPodAutoscalerUseRestClients <br /> *boolean*    | HorizontalPodAutoscalerUseRestClients determines if the new-style clients should be used if support for custom metrics is enabled.
image <br /> *string*    | Image is the docker image to use
leaderElection <br /> *[LeaderElectionConfiguration](#leaderelectionconfiguration-v1alpha2-kops)*    | LeaderElection defines the configuration of leader election client.
logLevel <br /> *integer*    | LogLevel is the defined logLevel
master <br /> *string*    | Master is the url for the kube api master
nodeMonitorGracePeriod <br /> *[Duration](#duration-v1-meta)*    | NodeMonitorGracePeriod is the amount of time which we allow running Node to be unresponsive before marking it unhealthy. (default 40s) Must be N-1 times more than kubelet's nodeStatusUpdateFrequency, where N means number of retries allowed for kubelet to post node status.
nodeMonitorPeriod <br /> *[Duration](#duration-v1-meta)*    | NodeMonitorPeriod is the period for syncing NodeStatus in NodeController. (default 5s)
podEvictionTimeout <br /> *[Duration](#duration-v1-meta)*    | PodEvictionTimeout is the grace period for deleting pods on failed nodes. (default 5m0s)
rootCAFile <br /> *string*    | rootCAFile is the root certificate authority will be included in service account's token secret. This must be a valid PEM-encoded CA bundle.
serviceAccountPrivateKeyFile <br /> *string*    | ServiceAccountPrivateKeyFile the location for a certificate for service account signing
terminatedPodGCThreshold <br /> *integer*    | TerminatedPodGCThreshold is the number of terminated pods that can exist before the terminated pod garbage collector starts deleting terminated pods. If <= 0, the terminated pod garbage collector is disabled.
useServiceAccountCredentials <br /> *boolean*    | UseServiceAccountCredentials controls whether we use individual service account credentials for each controller.

