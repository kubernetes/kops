# Calico Version 3
In early 2018 Version 3 of Calico was released, it included a reworked data
model and with that a switch from the etcd v2 to v3 API. This document covers
the requirements, upgrade process, and configuration to install
Calico Version 3.

## Requirements

- The main requirement needed for Calico Version 3 is the etcd v3 API available
  with etcd server version 3.
- Another requirement is for the Kubernetes version to be a minimum of v1.7.0.

### etcd
Due to the etcd v3 API being a requirement of Calico Version 3
(when using etcd as the datastore) not all Kops installations will be
upgradable to Calico V3. Installations using etcd v2 (or earlier) will need
to remain on Calico V2 or update to etcdv3.

## Configuration of a new cluster
To ensure a new cluster will have Calico Version 3 installed the following
two configurations options should be set:

- `spec.etcdClusters.etcdMembers[0].Version` (Main cluster) should be
  set to a Version of etcd greater than 3.x or the default version
  needs to be greater than 3.x.
- The Networking config must have the Calico MajorVersion set to `v3` like
  the following:
  ```
  spec:
    networking:
      calico:
        majorVersion: v3
  ```

Both of the above two settings can be set by doing a `kops edit cluster ...`
before bringing the cluster up for the first time.

With the above two settings your Kops deployed cluster will be running with
Calico Version 3.

### Create cluster networking flag

When enabling Calico with the `--networking calico` flag, etcd will be set to
a v3 version. Feel free to change to a different v3 version of etcd.

## Upgrading an existing cluster
Assuming your cluster meets the requirements it is possible to upgrade
your Calico Kops cluster.

A few notes about the upgrade:

- During the first portion of the migration, while the calico-kube-controllers
  pod is running its Init, no new policies will be applied though already
  applied policy will be active.
- During the migration no new pods will be scheduled as adding new workloads
  to Calico is blocked. Once the calico-complete-upgrade job has completed
  pods will once again be schedulable.
- The upgrade process that has been automated in kops can be found in
  [the Upgrading Calico docs](https://docs.projectcalico.org/v3.1/getting-started/kubernetes/upgrade/upgrade).

Perform the upgrade with the following steps:

1. First you must ensure that you are running Calico V2.6.5+. With the
   latest Kops (greater than 1.9) ensuring your cluster is updated can be
   done by doing a `kops update` on the cluster.
1. Verify your Calico data will migrate successfully by installing and
   configuring the
   [calico-upgrade command](https://docs.projectcalico.org/v3.1/getting-started/kubernetes/upgrade/setup)
   and then run `calico-upgrade dry-run` and verify it reports that the
   migration can be completed successfully.
1. Set `majorVersion` field as below by editing
   your cluster configuration with `kops edit cluster`.
   ```
   spec:
     networking:
       calico:
         majorVersion: v3
   ```
1. Update your cluster with `kops update` like you would normally update.
1. Monitor the progress of the migration by using
   `kubectl get pods -n kube-system` and checking the status of the following pods:
   - calico-node pods should restart one at a time and all becoming Running
   - calico-kube-controllers pod will restart and after the first calico-node
     pod starts running it will start running
   - calico-complete-upgrade pod will be Completed after all the calico-node
     pods start running
   If any of the above fail by entering a crash loop you should investigate
   by checking the logs with `kubectl -n kube-system logs <pod name>`.
1. Once the calico-node and calico-kube-controllers are running and the
   calico-complete-upgrade pod has completed the migration has finished
   successfully.

### Recovering from a partial migration

The InitContainer of the first calico-node pod that starts will perform the
datastore migration necessary for upgrading from Calico v2 to Calico v3, if
this InitContainer is killed or restarted when the new datastore is being
populated it will be necessary to manually remove the Calico data in the
etcd v3 API before the migration will be successful.
