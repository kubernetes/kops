## KubeProxyConfig v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | KubeProxyConfig



KubeProxyConfig defined the configuration for a proxy

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
clusterCIDR <br /> *string*    | ClusterCIDR is the CIDR range of the pods in the cluster
cpuRequest <br /> *string*    | 
featureGates <br /> *object*    | FeatureGates is a series of key pairs used to switch on features for the proxy
hostnameOverride <br /> *string*    | HostnameOverride, if non-empty, will be used as the identity instead of the actual hostname.
image <br /> *string*    | 
logLevel <br /> *integer*    | LogLevel is the logging level of the proxy
master <br /> *string*    | Master is the address of the Kubernetes API server (overrides any value in kubeconfig)

