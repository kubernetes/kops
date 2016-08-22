## Kubernetes Networking Options

kops sets up networking on AWS using VPC networking, where the master allocates a /24 CIDR to each Pod,
drawing from the Pod network.  Routes for each node are then configured in the AWS VPC routing tables.

One important limitation to note is that an AWS routing table cannot have more than 50 entries, which sets a limit of
50 nodes per cluster.  AWS support will sometimes raise the limit to 100, but performance limitations mean
they are unlikely to raise it further.

Because k8s modifies the AWS routing table, this means that realistically kubernetes needs to own the
routing table, and thus it requires its own subnet.  It is theoretically possible to share a routing table
with other infrastructure (but not a second cluster!), but this is not really recommended.

kops will support other networking options as they add support for the daemonset method of deployment.


kops currently supports 3 networking modes:

* `classic` kubernetes native networking, done in-process
* `kubenet` kubernetes native networking via a CNI plugin.  Also has less reliance on Docker's networking.
* `external` networking is done via a Daemonset

TODO: Explain the difference between pod networking & inter-pod networking.