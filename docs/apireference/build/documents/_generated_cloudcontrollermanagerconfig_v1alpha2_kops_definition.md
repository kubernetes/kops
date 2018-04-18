## CloudControllerManagerConfig v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | CloudControllerManagerConfig





<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
allocateNodeCIDRs <br /> *boolean*    | AllocateNodeCIDRs enables CIDRs for Pods to be allocated and, if ConfigureCloudRoutes is true, to be set on the cloud provider.
cidrAllocatorType <br /> *string*    | CIDRAllocatorType specifies the type of CIDR allocator to use.
cloudProvider <br /> *string*    | CloudProvider is the provider for cloud services.
clusterCIDR <br /> *string*    | ClusterCIDR is CIDR Range for Pods in cluster.
clusterName <br /> *string*    | ClusterName is the instance prefix for the cluster.
configureCloudRoutes <br /> *boolean*    | ConfigureCloudRoutes enables CIDRs allocated with to be configured on the cloud provider.
image <br /> *string*    | Image is the OCI image of the cloud controller manager.
leaderElection <br /> *[LeaderElectionConfiguration](#leaderelectionconfiguration-v1alpha2-kops)*    | LeaderElection defines the configuration of leader election client.
logLevel <br /> *integer*    | LogLevel is the verbosity of the logs.
master <br /> *string*    | Master is the url for the kube api master.
useServiceAccountCredentials <br /> *boolean*    | UseServiceAccountCredentials controls whether we use individual service account credentials for each controller.

