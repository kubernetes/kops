# Rolling Updates

Upgrading and modifying a k8s cluster often requires the replacement of nodes.  To not cause loss of service and other disruptions a kops uses a functionality call rolling updates.  Rolling a cluster is the replacement of the masters and nodes with new cloud instances.

When starting the rolling update, kops will check each instance in the instance group if it needs to be updated, so when you just update nodes, the master will not be updated. When your rolling update is interrupted and you run another rolling update, instances that have been updated before will not be updated again.

![Rolling Update Diagram](images/rolling-update.png?raw=true "Rolling Updates Diagram")

`kops` executes steps 2-4 for all the masters until the masters are replaced.   Then the same process is followed to replaces all nodes.
