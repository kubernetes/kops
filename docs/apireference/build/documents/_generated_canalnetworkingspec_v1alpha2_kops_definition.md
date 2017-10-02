## CanalNetworkingSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | CanalNetworkingSpec



CanalNetworkingSpec declares that we want Canal networking

<aside class="notice">
Appears In:

<ul> 
<li><a href="#networkingspec-v1alpha2-kops">NetworkingSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
chainInsertMode <br /> *string*    | ChainInsertMode controls whether Felix inserts rules to the top of iptables chains, or appends to the bottom. Leaving the default option is safest to prevent accidentally breaking connectivity. Default: 'insert' (other options: 'append')
defaultEndpointToHostAction <br /> *string*    | DefaultEndpointToHostAction allows users to configure the default behaviour for traffic between pod to host after calico rules have been processed. Default: ACCEPT (other options: DROP, RETURN)
prometheusGoMetricsEnabled <br /> *boolean*    | PrometheusGoMetricsEnabled enables Prometheus Go runtime metrics collection
prometheusMetricsEnabled <br /> *boolean*    | PrometheusMetricsEnabled can be set to enable the experimental Prometheus metrics server (default: false)
prometheusMetricsPort <br /> *integer*    | PrometheusMetricsPort is the TCP port that the experimental Prometheus metrics server should bind to (default: 9091)
prometheusProcessMetricsEnabled <br /> *boolean*    | PrometheusProcessMetricsEnabled enables Prometheus process metrics collection

