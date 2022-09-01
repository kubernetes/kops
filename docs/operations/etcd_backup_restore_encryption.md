# Etcd

## Backing up etcd

Kubernetes is relying on etcd for state storage. More details about the usage
can be found [here](https://kubernetes.io/docs/admin/etcd/) and
[here](https://coreos.com/etcd/docs/latest/).

### Backup requirement

A Kubernetes cluster deployed with kOps stores the etcd state in two different
AWS EBS volumes per master node. One volume is used to store the Kubernetes
main data, the other one for events. For a HA master with three nodes this will
result in six volumes for etcd data (one in each AZ). An EBS volume is designed
to have a [failure rate](https://aws.amazon.com/ebs/details/#AvailabilityandDurability)
of 0.1%-0.2% per year.

## Taking backups

Backups are done periodically and before cluster modifications using [etcd-manager](etcd_administration.md)
(introduced in kOps 1.12). Backups for both the `main` and `events` etcd clusters
are stored in object storage (like S3) together with the cluster configuration.

By default, backups are taken every 15 min. Hourly backups are kept for 1 week and
daily backups are kept for 1 year, before being automatically removed. The retention
duration for backups [can be adjusted](../cluster_spec.md#etcd-backups-retention)
to suit other needs.

## Restore backups

In case of a disaster situation with etcd (lost data, cluster issues etc.) it's
possible to do a restore of the etcd cluster using `etcd-manager-ctl`.
You can download the `etcd-manager-ctl` binary from the [etcd-manager repository](https://github.com/kopeio/etcd-manager/releases).
It is not necessary to run `etcd-manager-ctl` in your cluster, as long as you have access to cluster state storage (like S3).

Please note that this process involves downtime for your masters (and so the api server).
A restore cannot be undone (unless by restoring again), and you might lose pods, events
and other resources that were created after the backup.

For this example, we assume we have a cluster named `test.my.clusters` in a S3 bucket called `my.clusters`.

List the backups that are stored in your state store (note that backup files are different for the `main` and `events` clusters):

```
etcd-manager-ctl --backup-store=s3://my.clusters/test.my.clusters/backups/etcd/main list-backups
etcd-manager-ctl --backup-store=s3://my.clusters/test.my.clusters/backups/etcd/events list-backups
```

Add a restore command for both clusters:

```
etcd-manager-ctl --backup-store=s3://my.clusters/test.my.clusters/backups/etcd/main restore-backup [main backup dir]
etcd-manager-ctl --backup-store=s3://my.clusters/test.my.clusters/backups/etcd/events restore-backup [events backup dir]
```

Note that this does not start the restore immediately; you need to restart etcd on all masters.
You can do this with a `docker stop` or `kill` on the etcd-manager containers on the masters (the container names start with `k8s_etcd-manager_etcd-manager`).
The etcd-manager containers should restart automatically, and pick up the restore command. You also have the option to roll your masters quickly, but restarting the containers is preferred.

A new etcd cluster will be created and the backup will be
restored onto this new cluster. Please note that this process might take a short while,
depending on the size of your cluster.

You can follow the progress by reading the etcd logs (`/var/log/etcd(-events).log`)
on the master that is the leader of the cluster (you can find this out by checking the etcd logs on all masters).
Note that the leader might be different for the `main` and `events` clusters.

## Verify master lease consistency

[This bug](https://github.com/kubernetes/kubernetes/issues/86812) causes old apiserver leases to get stuck. In order to recover from this you need to remove the leases from etcd directly. 

To verify if you are affect by this bug, check the endpoints resource of the kubernetes apiserver, like this:
```
kubectl get endpoints/kubernetes -o yaml
```

If you see more address than masters, you will need to remove it manually inside the etcd cluster.

See [etcd administation](/operations/etcd_administration) how to obtain access to the etcd cluster.

Once you have a working etcd client, run the following:
```
etcdctl get --prefix --keys-only /registry/masterleases
```

Also you can delete all of the leases in one go... 
```
etcdctl del --prefix /registry/masterleases/
```

The remaining api servers will immediately recreate their own leases. Check again the above-mentioned endpoint to verify the problem has been solved.

Because the state on each of the Nodes may differ from the state in etcd, it is also a good idea to do a rolling-update of the entire cluster:

```sh
kops rolling-update cluster --force --yes
```

For more information and troubleshooting, please check the [etcd-manager documentation](https://github.com/kubernetes-sigs/etcdadm/etcd-manager).

## Etcd Volume Encryption

You must configure etcd volume encryption before bringing up your cluster. You cannot add etcd volume encryption to an already running cluster.

### Encrypting Etcd Volumes Using the Default AWS KMS Key

Edit your cluster to add `encryptedVolume: true` to each etcd volume:

`kops edit cluster ${CLUSTER_NAME}`

```yaml
...
etcdClusters:
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
    encryptedVolume: true
  name: main
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
    encryptedVolume: true
  name: events
...
```

Update your cluster:

```
kops update cluster ${CLUSTER_NAME}
# Review changes before applying
kops update cluster ${CLUSTER_NAME} --yes
```

### Encrypting Etcd Volumes Using a Custom AWS KMS Key

Edit your cluster to add `encryptedVolume: true` to each etcd volume:

`kops edit cluster ${CLUSTER_NAME}`

```yaml
...
etcdClusters:
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
    encryptedVolume: true
    kmsKeyId: <full-arn-of-your-kms-key>
  name: main
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
    encryptedVolume: true
    kmsKeyId: <full-arn-of-your-kms-key>
  name: events
...
```

Update your cluster:

```
kops update cluster ${CLUSTER_NAME}
# Review changes before applying
kops update cluster ${CLUSTER_NAME} --yes
```
