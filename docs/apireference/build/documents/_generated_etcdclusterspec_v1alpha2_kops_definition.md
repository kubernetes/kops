## EtcdClusterSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | EtcdClusterSpec



EtcdClusterSpec is the etcd cluster specification

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
enableEtcdTLS <br /> *boolean*    | EnableEtcdTLS indicates the etcd service should use TLS between peers and clients
etcdMembers <br /> *[EtcdMemberSpec](#etcdmemberspec-v1alpha2-kops) array*    | Members stores the configurations for each member of the cluster (including the data volume)
name <br /> *string*    | Name is the name of the etcd cluster (main, events etc)
version <br /> *string*    | Version is the version of etcd to run i.e. 2.1.2, 3.0.17 etcd

