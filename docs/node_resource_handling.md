## Node Resource Handling In Kubernetes

An aspect of Kubernetes clusters that is often overlooked is the resources non-
pod components require to run, such as:

* Operating system components i.e. `sshd`, `udev` etc.
* Kubernetes system components i.e. `kubelet`, `container runtime` (e.g.
  Docker), `node problem detector`, `journald` etc.

As you manage your cluster, it's important that you are cognisant of these
components because if your critical non-pod components don't have enough
resources, you might end up with a very unstable cluster.

### Understanding Node Resources

Each node in a cluster has resources available to it and pods scheduled to run
on the node may or may not have resource requests or limits set on them.
Kubernetes schedules pods on nodes that have resources that satisfy the pod's
specified requirements. Broadly, pods are [bin-packed][4] onto the nodes in a
best effort attempt to utilize as much of the resources available with as few
nodes as possible.

```
      Node Capacity
---------------------------
|     kube-reserved       |
|-------------------------|
|     system-reserved     |
|-------------------------|
|    eviction-threshold   |
|-------------------------|
|                         |
|      allocatable        |
|   (available for pods)  |
|                         |
|                         |
---------------------------
```

Node resources can be categorised into 4 (as shown above):

* `kube-reserved` – reserves resources for kubernetes system daemons.
* `system-reserved` – reserves resources for operating system components.
* `eviction-threshold` – specifies limits that trigger evictions when node
  resources drop below the reserved value.
* `allocatable` – the remaining node resources available for scheduling of pods
  when `kube-reserved`, `system-reserved` and `eviction-threshold` resources
  have been accounted for.

For example, with a 30.5 GB, 4 vCPUs machine with only `eviction-thresholds` set
as `--eviction-hard=memory.available<100Mi` we'd get the following `Capacity`
and `Allocatable` resources:

```
$ kubectl describe node/ip-xx-xx-xx-xxx.internal
...
Capacity:
 cpu:   4
 memory:  31402412Ki
 ...
Allocatable:
 cpu:   4
 memory:  31300012Ki
 ...
```

### So, What Could Possibly Go Wrong?

The scheduler ensures that for each resource type, the sum of the resources
scheduled does not surpass the sum of allocatable resources. But suppose you
have a couple of applications deployed in your cluster that are constantly using
up way more resources set in their resource requests (burst above requests but
below limits during workload). You end up with a node with pods that are each
attempting to take up more resources than there are available on the node!

This is particularly an issue with non-compressible resources like memory. For
example, in the aforementioned case, with an eviction threshold of only
`memory.available<100Mi` and no `kube-reserved` nor `system-reserved`
reservations set, it is possible for a node to OOM prior to when `kubelet` is
able to reclaim memory (because it may not observe memory pressure right away,
since it polls `cAdvisor` to collect memory usage stats at a regular interval).

All the while, keep in mind that without `kube-reserved` nor `system-reserved`
reservations set (which is most clusters i.e. [GKE][5], [Kops][6]), the
scheduler doesn't account for resources that non-pod components would require to
function properly because `Capacity` and `Allocatable` resources are more or
less equal.

### Where Do We Go From Here?

It's difficult to give a one size fits all answer to node resource allocation.
The behaviour of your cluster depends on the resource requirements of the apps
running on the cluster, the pod density and the cluster size. But there's a
[node performance dashboard][7] that exposes `cpu` and `memory` usage profiles
of `kubelet` and `docker` engine at multiple levels of pod density which may
serve as a guide for what values would be appropriate for your cluster.

But, it seems fitting to recommend the following:

1. Always set requests with some breathing room – do not set requests to match
   your application's resource profile during idle time too closely.
2. Always set limits – so that your application doesn't hog all the memory on a
   node during a spike.
3. Don't set your limits for incompressible resources too high - at the end of
   the day, the Kubernetes scheduler schedules based on resource requests which
   match what's available on the node. During a spike, your pod technically will
   try to access resources outside what it's guaranteed to have access to. As
   explained before, this can be an issue if a bunch of your pods are all
   bursting at the same time.
4. Increase eviction thresholds if they are too low - while extreme utilization
   is ideal, it might be too close to the edge such that the system doesn't have
   enough time to reclaim resources via evictions if the resource increases
   within that window rapidly.
5. Reserve resources for system components once you've been able to profile your
   nodes i.e. `kube-reserved` and `system-reserved`.

**Further Reading:**

 * [Configure Out Of Resource Handling][2]
 * [Reserve Compute Resources for System Daemons][1]
 * [Managing Compute Resources for Containers][3]
 * [Visualize Kubelet Performance with Node Dashboard][8]

[1]: https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/
[2]: https://kubernetes.io/docs/tasks/administer-cluster/out-of-resource/
[3]: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
[4]: https://en.wikipedia.org/wiki/Bin_packing_problem
[5]: https://cloud.google.com/container-engine/
[6]: https://github.com/kubernetes/kops
[7]: http://node-perf-dash.k8s.io/#/builds
[8]: http://kubernetes.io/blog/2016/11/visualize-kubelet-performance-with-node-dashboard.html
