kops: Operate Kubernetes the Kubernetes Way

kops (Kubernetes-Ops) is a set of tools for installing, operating and deleting Kubernetes clusters.

It follows the Kubernetes design philosophy: the user creates a Cluster configuration object in JSON/YAML,
and then controllers create the Cluster.

Each component (kubelet, kube-apiserver...) is explicitly configured: We reuse the k8s componentconfig types
where we can, and we create additional types for the configuration of additional components.

kops can:

* create a cluster
* upgrade a cluster
* reconfigure the components
* add, remove or reconfigure groups of machines (InstanceGroups)
* manage cluster add-ons
* delete a cluster

Some users will need or prefer to use tools like Terraform for cluster configuration,
so kops can also output the equivalent configuration for those tools also (currently just Terraform, others
planned).  After creation with your preferred tool, you can still use the rest of the kops tooling to operate
your cluster.

## Primary API types

There are two primary types:

* Cluster represents the overall cluster configuration (such as the version of kubernetes we are running), and contains default values for the individual nodes.

* InstanceGroup is a group of instances with similar configuration that are managed together.
  Typically this is a group of Nodes or a single master instance.  On AWS, it is currently implemented by an AutoScalingGroup.

## State Store

The API objects are currently stored in an abstraction called a ["state store"](/state.md) has more detail.

## Configuration inference

Configuration of a kubernetes cluster is actually relatively complicated: there are a lot of options, and many combinations
must be configured consistently with each other.

Similar to the way creating a Kubernetes object populates other spec values, the `kops create cluster` command will infer other values
that are not set, so that you can specify a minimal set of values (but if you don't want to override the default value, you simply specify the fields!).

Because more values are inferred than with simpler k8s objects, we record the user-created spec separately from the
complete inferred specification.  This means we can keep track of which values were actually set by the user, vs just being
default values; this lets us avoid some of the problems e.g. with ClusterIP on a Service.

We aim to remove any computation logic from the downstream pieces (i.e. nodeup & protokube); this means there is a
single source of truth and it is practical to implement alternatives to nodeup & protokube.  For example, components
such as kubelet might read their configuration directly from the state store in future, eliminating the need to
have a management process that copies values around.

Currently the 'completed' cluster specification is stored in the state store in a file called `cluster.spec`
