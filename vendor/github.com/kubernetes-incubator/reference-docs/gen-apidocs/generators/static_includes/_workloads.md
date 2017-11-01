# <strong>WORKLOADS</strong>

Worloads resources are responsible for managing and running your containers on the cluster.  [Containers](#container-v1-core) are created
by Controllers through [Pods](#pod-v1-core).  Pods run Containers and provide environmental dependencies such as shared or
persistent storage [Volumes](#volume-v1-core) and [Configuration](#configmap-v1-core) or [Secret](#secret-v1-core) data injected into the
container.

The most common Controllers are:

- [Deployments](#deployment-v1beta1-apps) for stateless persistent apps (e.g. http servers)
- [StatefulSets](#statefulset-v1beta1-apps) for stateful persistent apps (e.g. databases)
- [Jobs](#job-v1-batch) for run-to-completion apps (e.g. batch jobs).

------------
