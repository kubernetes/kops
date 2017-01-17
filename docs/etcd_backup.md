# Backing up etcd

Kubernetes is relying on etcd for state storage. More details about the usage
can be found [here](https://kubernetes.io/docs/admin/etcd/).

## Backup requirement

A Kubernetes cluster deployed with kops stores the etcd state in two different
AWS EBS volumes for each master node. An EBS volume are designed to have a
[failure rate](https://aws.amazon.com/ebs/details/#AvailabilityandDurability)
of 0.1%-0.2% per year. This results in the requirement that these volumes should
be backuped like any other persistent data.

## Create volume backups

Kubernetes does currently not provide any option to do regular backups of etcd
out of the box.

It is possible to setup a scheduled backup with EBS snapshots and AWS Lambda,
more can be found [here](https://serverlesscode.com/post/lambda-schedule-ebs-snapshot-backups/).

## Restore volume backups

In case of a lost EBS volume it is possible to restore the volume from a
snapshot, more details [here](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-restoring-volume.html).

Kubernetes uses protokube to identify the right volumes for etcd. Therefore it
is important to tag the EBS volumes with the correct tags after restoring them
from a EBS snapshot.

protkube will look for the following tags:

* `KubernetesCluster` containing the cluster name (e.g. `k8s.mycompany.tld`)
* `Name` containing the volume name (e.g. `eu-central-1a.etcd-main.k8s.mycompany.tld`)
* `k8s.io/etcd/main` containg the availability zone of the volume (e.g. `
eu-central-1a/eu-central-1a`)
* `k8s.io/role/master` with the value `1`

After fully restoring the volume ensure that the old volume is no longer there,
or you've removed the tags from the old volume. After restarting the master node
Kubernetes should pick up the new volume and start running again.
