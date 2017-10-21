# <strong>DISCOVERY & LOAD BALANCING</strong>

Discovery and Load Balancing resources are responsible for stitching your workloads together into an accessible Loadbalanced Service.  By default,
[Workloads](#workloads) are only accessible within the cluster, and they must be exposed externally using a either
a *LoadBalancer* or *NodePort* [Service](#service-v1-core).  For development, internally accessible
Workloads can be accessed via proxy through the api master using the `kubectl proxy` command.

Common resource types:

- [Services](#service-v1-core) for providing a single ip endpoint loadbalanced across multiple Workload replicas.
- [Ingress](#ingress-v1beta1-extensions) for providing a https(s) endpoint http(s) routed to one or more *Services*

------------
