## ClusterSubnetSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | ClusterSubnetSpec





<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
cidr <br /> *string*    | 
egress <br /> *string*    | Egress defines the method of traffic egress for this subnet
id <br /> *string*    | ProviderID is the cloud provider id for the objects associated with the zone (the subnet on AWS)
name <br /> *string*    | 
publicIP <br /> *string*    | PublicIP to attach to NatGateway
region <br /> *string*    | Region is the region the subnet is in, set for subnets that are regionally scoped
type <br /> *string*    | 
zone <br /> *string*    | Zone is the zone the subnet is in, set for subnets that are zonally scoped

