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



## Switching between networking providers

Make sure you are running the latest kops: `git pull && make`.  `kops version` should be `Version git-2f4ac90`

`kops edit cluster` and you should see a block like:

```
  networking:
    classic: {}
```

That means you are running with `classic` networking.  The `{}` means there are no configuration options
(but we have to put something in there so that we do choose classic).

To switch to kubenet, edit to be:
```
  networking:
    kubenet: {}
```

Now follow the normal update / rolling-update procedure:

* `kops update cluster` to preview
* `kops update cluster --yes` to apply
* `kops rolling-update cluster` to preview the rolling-update
* `kops rolling-update cluster --yes` to roll all your instances

Your cluster should be ready in a few minutes.

It is not trivial to see that this has worked; the easiest way seems to be to SSH to the master and verify
that kubelet has been run with `--network-plugin=kubenet`