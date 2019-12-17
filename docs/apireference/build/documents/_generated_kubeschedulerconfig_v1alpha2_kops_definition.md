## KubeSchedulerConfig v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | KubeSchedulerConfig



KubeSchedulerConfig is the configuration for the kube-scheduler

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
featureGates <br /> *object*    | FeatureGates is set of key=value pairs that describe feature gates for alpha/experimental features.
image <br /> *string*    | Image is the docker image to use
leaderElection <br /> *[LeaderElectionConfiguration](#leaderelectionconfiguration-v1alpha2-kops)*    | LeaderElection defines the configuration of leader election client.
logLevel <br /> *integer*    | LogLevel is the logging level
master <br /> *string*    | Master is a url to the kube master
usePolicyConfigMap <br /> *boolean*    | UsePolicyConfigMap enable setting the scheduler policy from a configmap

