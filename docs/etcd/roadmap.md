# Kops and etcd

Kops currently manages etcd using a daemon that runs on the masters called protokube.  protokube
is responsible for:

* mapping volumes so that exactly one master runs each member of the etcd cluster,
even when the masters are dynamic (in autoscaling groups).
* mapping DNS names so that etcd nodes can discover each other even as they move around.

We are following the [approach recommended by the authors of etcd](https://github.com/coreos/etcd/issues/5418)

protokube has additional responsibilities also, so we call this the protokube-integrated etcd support.

Generally, splitting up protokube is part of the general kops roadmap, where we aim to split out our integrated
tooling into composable tooling that can be upgraded (or even used) separately.

## Limitations

The current approach for managing etcd makes certain tasks hard:

* upgrades/downgrades between etcd versions
* resizing the cluster

To address these limitations, we plan to adopt the [etcd-manager](https://github.com/kopeio/etcd-manager) as
it matures.

To make adoption easier, the etcd-manager has added a standalone backup tool, that can backup etcd into the
[expected structure](https://github.com/kopeio/etcd-manager/blob/master/docs/backupstructure.md), even if you are not running the etcd-manager.  It should be possible to then use
the etcd-manager from that backup.

## Roadmap

### _kops 1.9_

* Use the etcd-backup tool to allow users to opt-in to backups with the protokube-integrated etcd support, in the format that etcd-manager expects

Goal: Users can enable backups on a running cluster.

### _kops 1.10_

* Make the etcd-backup tool enabled-by-default, so everyone should have backups.
* Allow users to opt-in to the full etcd-manager.
* Make etcd3 the default for new clusters, now that we have an upgrade path.

Goal: Users that want to move from etcd2 to etcd3 can enable backups
on an existing cluster (running kops 1.9 or later), then enable the etcd-manager (with kops 1.10 or later).

### _kops 1.11_

* Make the etcd-manager the default, deprecate the protokube-integrated approach

Goal: Users are fully able to manage etcd - moving between versions or resizing their clusters.

### _untargeted_

* Remove the protokube-integrated etcd support (_untargeted_)
