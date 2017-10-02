## RomanaNetworkingSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | RomanaNetworkingSpec



RomanaNetworkingSpec declares that we want Romana networking

<aside class="notice">
Appears In:

<ul> 
<li><a href="#networkingspec-v1alpha2-kops">NetworkingSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
daemonServiceIP <br /> *string*    | DaemonServiceIP is the Kubernetes Service IP for the romana-daemon pod
etcdServiceIP <br /> *string*    | EtcdServiceIP is the Kubernetes Service IP for the etcd backend used by Romana

