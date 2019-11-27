# Architecture: kops-controller

kops-controller runs as a container on the master node(s).  It is a kubebuilder
controller, that performs runtime reconciliation for kops.

Controllers in kops-controller:

* NodeController


## NodeController

The NodeController watches Node objects, and applies labels to them from a
controller.  Previously this was done by the kubelet, but the fear was that this
was insecure and so [this functionality was removed](https://github.com/kubernetes/enhancements/blob/master/keps/sig-auth/0000-20170814-bounding-self-labeling-kubelets.md).

The main difficulty here is mapping from a Node to an InstanceGroup in a way
that does not render the system just as vulnerable to spoofing as it was
previously.

NodeController uses the cloud APIs to make this link (in future, cluster-api may
offer an alternative).  The theory is that we can then work to prevent spoofing
of the Node's `providerID`, and further we assume that an attacker that has
gained the ability to manipulate the underlying cloud itself has already
bypassed our protections.

On AWS, tags are not easily mutable from a Node; so we simply set a tag
with the name of the instance-group.  When we see a node, we query EC2 for the
instance defined in `providerID`, and we get the instance group name from the
tag.  We then query the instance group definition from the underlying store
(typically S3), construct the correct tags, and apply the tags.

On GCE, the metadata is more mutable.  So we query the instance, but then we
find the owning MIG, query the instances that are part of that MIG to verify
that the instance is indeed part of the MIG, and then we get the metadata from
the instance template (which is not easily mutated from the instance).  We then
get the instance group definition from the underlying store, as elsewhere.
