# <strong>METADATA</strong>

Metadata resources are responsible for configuring behavior of your other Resources within the Cluster.

Common resource types:

- [HorizontalPodAutoscaler](#horizontalpodautoscaler-v1-autoscaling) (HPA) for automatically scaling the replicacount of your workloads in response to load
- [PodDisruptionBudget](#poddisruptionbudget-v1alpha1) for configuring how many replicas in a given workload maybe made concurrently unavailable when performing maintenance.
- [ThirdPartyResource](#thirdpartyresource-v1beta1) for extending the Kubernetes APIs with your own types
- [Event](#event-v1-core) for notification of resource lifecycle events in the cluster.

------------
