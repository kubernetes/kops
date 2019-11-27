# Etcd

## Backing up etcd

Kubernetes is relying on etcd for state storage. More details about the usage
can be found [here](https://kubernetes.io/docs/admin/etcd/) and
[here](https://coreos.com/etcd/docs/latest/).

### Backup requirement

A Kubernetes cluster deployed with kops stores the etcd state in two different
AWS EBS volumes per master node. One volume is used to store the Kubernetes
main data, the other one for events. For a HA master with three nodes this will
result in six volumes for etcd data (one in each AZ). An EBS volume is designed
to have a [failure rate](https://aws.amazon.com/ebs/details/#AvailabilityandDurability)
of 0.1%-0.2% per year.

### Backups using etcd-manager

Backups are done periodically and before cluster modifications using [etcd-manager](../etcd/manager.md)
(introduced in kops 1.12). Backups for both the `main` and `events` etcd clusters
are stored in object storage (like S3) together with the cluster configuration.

### Volume backups (legacy etcd)

If you are running your cluster in legacy etcd mode (without etcd-manager),
backups can be done through snapshots of the etcd volumes.

You can for example use CloudWatch to trigger an AWS Lambda with a defined schedule (e.g. once per
hour). The Lambda will then create a new snapshot of all etcd volumes. A complete
guide on how to setup automated snapshots can be found [here](https://serverlesscode.com/post/lambda-schedule-ebs-snapshot-backups/).

Note: this is one of many examples on how to do scheduled snapshots.

## Restore using etcd-manager

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
etcd-manager-ctl --backup-store=s3://my.clusters/test.my.clusters/backups/etcd/main restore-backup [main backup file]
etcd-manager-ctl --backup-store=s3://my.clusters/test.my.clusters/backups/etcd/events restore-backup [events backup file]
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

After the restore is complete, api server should come back up, and you should have a working cluster.
Note that the api server might be very busy for a while as it changes the cluster back to the state of the backup.
It's a good idea to temporarily increase the instance size of your masters and roll your worker nodes.

For more information and troubleshooting, please check the [etcd-manager documentation](https://github.com/kopeio/etcd-manager).

### Restore volume backups (legacy etcd)

If you're using legacy etcd (without etcd-manager), it is possible to restore the volume from a snapshot we created
earlier. Details about creating a volume from a snapshot can be found in the
[AWS documentation](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-restoring-volume.html).

Kubernetes uses protokube to identify the right volumes for etcd. Therefore it
is important to tag the EBS volumes with the correct tags after restoring them
from a EBS snapshot.

protokube will look for the following tags:

* `KubernetesCluster` containing the cluster name (e.g. `k8s.mycompany.tld`)
* `Name` containing the volume name (e.g. `eu-central-1a.etcd-main.k8s.mycompany.tld`)
* `k8s.io/etcd/main` containing the availability zone of the volume (e.g. `eu-central-1a/eu-central-1a`)
* `k8s.io/role/master` with the value `1`

After fully restoring the volume ensure that the old volume is no longer there,
or you've removed the tags from the old volume. After restarting the master node
Kubernetes should pick up the new volume and start running again.


## Etcd Volume Encryption

You must configure etcd volume encryption before bringing up your cluster. You cannot add etcd volume encryption to an already running cluster.

### Encrypting Etcd Volumes Using the Default AWS KMS Key

Edit your cluster to add `encryptedVolume: true` to each etcd volume:

`kops edit cluster ${CLUSTER_NAME}`

```
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

```
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
