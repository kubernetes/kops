## CalicoNetworkingSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | CalicoNetworkingSpec



CalicoNetworkingSpec declares that we want Calico networking

<aside class="notice">
Appears In:

<ul> 
<li><a href="#networkingspec-v1alpha2-kops">NetworkingSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
crossSubnet <br /> *boolean*    | 
logSeverityScreen <br /> *string*    | LogSeverityScreen lets us set the desired log level. (Default: info)
prometheusGoMetricsEnabled <br /> *boolean*    | PrometheusGoMetricsEnabled enables Prometheus Go runtime metrics collection
prometheusMetricsEnabled <br /> *boolean*    | PrometheusMetricsEnabled can be set to enable the experimental Prometheus metrics server (default: false)
prometheusMetricsPort <br /> *integer*    | PrometheusMetricsPort is the TCP port that the experimental Prometheus metrics server should bind to (default: 9091)
prometheusProcessMetricsEnabled <br /> *boolean*    | PrometheusProcessMetricsEnabled enables Prometheus process metrics collection

