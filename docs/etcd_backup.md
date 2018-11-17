# Backing up etcd

Kubernetes is relying on etcd for state storage. More details about the usage
can be found [here](https://kubernetes.io/docs/admin/etcd/) and
[here](https://coreos.com/etcd/docs/latest/v2/README.html).

## Backup requirement

A Kubernetes cluster deployed with kops stores the etcd state in two different
AWS EBS volumes per master node. One volume is used to store the Kubernetes
main data, the other one for events. For a HA master with three nodes this will
result in six volumes for etcd data (one in each AZ). An EBS volume is designed
to have a [failure rate](https://aws.amazon.com/ebs/details/#AvailabilityandDurability)
of 0.1%-0.2% per year.

## Create volume backups

Kubernetes does currently not provide any option to do regular backups of etcd
out of the box.

Therefore we have to either manually backup the etcd volumes regularly or use
other AWS services to do this in a automated, scheduled way. You can for example
use CloudWatch to trigger an AWS Lambda with a defined schedule (e.g. once per
hour). The Lambda will then create a new snapshot of all etcd volumes. A complete
guide on how to setup automated snapshots can be found [here](https://serverlesscode.com/post/lambda-schedule-ebs-snapshot-backups/).

Note: this is one of many examples on how to do scheduled snapshots.

## Restore volume backups

In case the Kubernetes cluster fails in a way that too many master nodes can't
access their etcd volumes it is impossible to get a etcd quorum.

In this case it is now possible to restore the volume from a snapshot we created
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
