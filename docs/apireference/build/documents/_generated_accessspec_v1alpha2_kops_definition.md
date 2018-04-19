## AccessSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | AccessSpec



AccessSpec provides configuration details related to kubeapi dns and ELB access

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
dns <br /> *[DNSAccessSpec](#dnsaccessspec-v1alpha2-kops)*    | DNS wil be used to provide config on kube-apiserver elb dns
loadBalancer <br /> *[LoadBalancerAccessSpec](#loadbalanceraccessspec-v1alpha2-kops)*    | LoadBalancer is the configuration for the kube-apiserver ELB

