## ExternalDNSConfig v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | ExternalDNSConfig



ExternalDNSConfig are options of the dns-controller

<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
disable <br /> *boolean*    | Disable indicates we do not wish to run the dns-controller addon
watchIngress <br /> *boolean*    | WatchIngress indicates you want the dns-controller to watch and create dns entries for ingress resources
watchNamespace <br /> *string*    | WatchNamespace is namespace to watch, detaults to all (use to control whom can creates dns entries)

