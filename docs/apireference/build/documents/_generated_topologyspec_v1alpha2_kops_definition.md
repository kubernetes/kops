## TopologySpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | TopologySpec





<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
bastion <br /> *[BastionSpec](#bastionspec-v1alpha2-kops)*    | Bastion provide an external facing point of entry into a network containing private network instances. This host can provide a single point of fortification or audit and can be started and stopped to enable or disable inbound SSH communication from the Internet, some call bastion as the "jump server".
dns <br /> *[DNSSpec](#dnsspec-v1alpha2-kops)*    | DNS configures options relating to DNS, in particular whether we use a public or a private hosted zone
masters <br /> *string*    | The environment to launch the Kubernetes masters in public|private
nodes <br /> *string*    | The environment to launch the Kubernetes nodes in public|private

