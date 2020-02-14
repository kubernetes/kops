## Kubernetes Bootstrap

This is an overview of how a Kubernetes cluster comes up, when using kops.

## From spec to complete configuration

The kops tool itself takes the (minimal) spec of a cluster that the user specifies,
and computes a complete configuration, setting defaults where values are not specified,
and deriving appropriate dependencies.  The "complete" specification includes the set
of all flags that will be passed to all components.  All decisions about how to install the
cluster are made at this stage, and thus every decision can in theory be changed if the user
specifies a value in the spec.

This complete specification is set in the launch configuration for the AutoScaling Group (on AWS),
or the Managed Instance Group (on GCE).

On both AWS & GCE, everything (nodes & masters) runs in an ASG/MIG; this means that failures
(or the user) can terminate machines and the system will self-heal.

## nodeup: from image to kubelet

nodeup is the component that installs packages and sets up the OS, sufficiently for
Kubelet.  The core requirements are:

* Docker must be installed.  nodeup will install Docker 1.13.1, the version of Docker tested with Kubernetes 1.8
* Kubelet, which is installed a systemd service

In addition, nodeup installs:

* Protokube, which is a kops-specific component

## /etc/kubernetes/manifests

kubelet starts pods as controlled by the files in /etc/kubernetes/manifests  These files are created
by nodeup and protokube (ideally all by protokube, but currently split between the two).

These pods are declared using the standard k8s manifests, just as if they were stored in the API.
But these are used to break the circular dependency for the bring-up of our core components, such
as etcd & kube-apiserver.

On masters:

* kube-apiserver
* kube-controller-manager (which runs miscellaneous controllers)
* kube-scheduler (which assigns pods to nodes)
* etcd (this is actually created by protokube though)
* dns-controller

On nodes:

* kube-proxy (which configures iptables so that the k8s-network will work)

It is possible to add custom static pods by using `fileAssets` in the
cluster spec. This might be useful for any custom bootstraping that
doesn't fit into `additionalUserData` or `hooks`.

## kubelet start

Kubelet starts up, starts (and restarts) all the containers in /etc/kubernetes/manifests.

It also tries to contact the API server (which the master kubelet will itself eventually start),
register the node.  Once a node is registered, kube-controller-manager will allocate it a PodCIDR,
which is an allocation of the k8s-network IP range.  kube-controller-manager updates the node, setting
the PodCIDR field.  Once kubelet sees this allocation, it will set up the
local bridge with this CIDR, which allows docker to start.  Before this happens, only pods
that have hostNetwork will work - so all the "core" containers run with hostNetwork=true.

## api-server bringup

The api-server will listen on localhost:8080 on the master.  This is an unsecured endpoint,
but is only reachable from the master, and only for pods running with hostNetwork=true.  This
is how components like kube-scheduler and kube-controller-manager can reach the API without
requiring a token.

APIServer also listens on the HTTPS port (443) on all interfaces.  This is a secured endpoint,
and requires valid authentication/authorization to use it.  This is the endpoint that node kubelet
will reach, and also that end-users will reach.

kops uses DNS to allow nodes and end-users to discover the api-server.  The apiserver pod manifest (in
 /etc/kubernetes/manifests) includes annotations that will cause the dns-controller to create the
 records.  It creates `api.internal.mycluster.com` for use inside the cluster (using InternalIP addresses),
 and it creates `api.mycluster.com` for use outside the cluster (using ExternalIP addresses).

## etcd bringup

etcd is where we have put all of our synchronization logic, so it is more complicated than most other pieces,
and we must be really careful when bringing it up.

kops follows CoreOS's recommend procedure for [bring-up of etcd on clouds](https://github.com/coreos/etcd/issues/5418):

* We have one EBS volume for each etcd cluster member (in different nodes)
* We attach the EBS volume to a master, and bring up etcd on that master
* We set up DNS names pointing to the etcd process
* We set up etcd with a static cluster, with those DNS names

Because the data is persistent and the cluster membership is also a static set of DNS names, this
means we don't need to manage etcd directly.  We just try to make sure that some master always have
each volume mounted with etcd running and DNS set correctly.  That is the job of protokube.

Protokube:

* discovers EBS volumes that hold etcd data (using tags)
* tries to safe_format_and_mount them
* if successful in mounting the volume, it will write a manifest for etcd into /etc/kubernetes/manifests
* configures DNS for the etcd nodes (we can't use dns-controller, because the API is not yet up)
* kubelet then starts and runs etcd

## node bringup

Most of this has focused on things that happen on the master, but the node bringup is very similar but simplified:

* nodeup installs docker & kubelet
* in /etc/kubernetes/manifests, we have kube-proxy

So kubelet will start up, as will kube-proxy.  It will try to reach the api-server on the internal DNS name,
and once the master is up it will succeed.  Then:

* kubelet creates a Node object representing itself
* kube-controller-manager sees the node creation and assigns it a PodCIDR
* kubelet sees the PodCIDR assignment and configures the local docker bridge (cbr0)
* the node will be marked as Ready, and kube-scheduler will start assigning pods to the node
* when kubelet sees assigned pods it will run them
