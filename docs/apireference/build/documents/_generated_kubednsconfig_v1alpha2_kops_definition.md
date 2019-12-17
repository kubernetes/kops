## KubeDNSConfig v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | KubeDNSConfig





<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterspec-v1alpha2-kops">ClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
cacheMaxConcurrent <br /> *integer*    | CacheMaxConcurrent is the maximum number of concurrent queries for dnsmasq
cacheMaxSize <br /> *integer*    | CacheMaxSize is the maximum entries to keep in dnsmaq
domain <br /> *string*    | 
image <br /> *string*    | Image is the name of the docker image to run Deprecated as this is now in the addon
replicas <br /> *integer*    | Deprecated as this is now in the addon, and controlled by autoscaler
serverIP <br /> *string*    | 

