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
cloudProvider <br /> *string*    | CloudProvider is the provider for cloud services.
clusterCIDR <br /> *string*    | ClusterCIDR is CIDR Range for Pods in cluster.
clusterName <br /> *string*    | ClusterName is the instance prefix for the cluster.
configureCloudRoutes <br /> *boolean*    | ConfigureCloudRoutes enables CIDRs allocated with to be configured on the cloud provider.
image <br /> *string*    | Image is the docker image to use
leaderElection <br /> *[LeaderElectionConfiguration](#leaderelectionconfiguration-v1alpha2-kops)*    | LeaderElection defines the configuration of leader election client.
logLevel <br /> *integer*    | LogLevel is the defined logLevel
master <br /> *string*    | Master is the url for the kube api master
rootCAFile <br /> *string*    | rootCAFile is the root certificate authority will be included in service account's token secret. This must be a valid PEM-encoded CA bundle.
serviceAccountPrivateKeyFile <br /> *string*    | ServiceAccountPrivateKeyFile the location for a certificate for service account signing
terminatedPodGCThreshold <br /> *integer*    | TerminatedPodGCThreshold is the number of terminated pods that can exist before the terminated pod garbage collector starts deleting terminated pods. If <= 0, the terminated pod garbage collector is disabled.
useServiceAccountCredentials <br /> *boolean*    | UseServiceAccountCredentials controls whether we use individual service account credentials for each controller.

