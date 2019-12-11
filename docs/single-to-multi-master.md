# Migrating from single to multi-master

This document describes how to go from a single-master cluster (created by kops)
to a multi-master cluster. If you are using etcd-manager you just need to perform some steps of the migration.  

# etcd-manager

If you are using etcd-manager, just perform the steps in this section. Etcd-manager is default for kops 1.12. Etcd-manager makes the upgrade to multi-master much smoother. 

The list references steps of the next section. To upgrade from a single master to a cluster with three masters:

- Skip Step 1 since Etcd-manager is doing automatic backups to S3
- Create Instance Groups (Section 2 below)
  - create the subnets
  - create the instance groups (no need to disable the third master, leave minSize and maxSize at 1)
  - add the masters to your etcd cluster definition (both in section named main and events)
- Skip Step 3 and 4
- Now you are ready to update the AWS configuration:
  - `kops update cluster your-cluster-name`
- AWS will launch two new masters, they will be discovered and then configured by etcd-manager
- check with `kubectl get nodes` to see everything is ready
- Cleanup (Step 5) to do a rolling restart of all masters (just in case)

# Etcd without etcd-manager

## 0 - Warnings

This is a risky procedure that **can lead to data-loss** in the etcd cluster.
Please follow all the backup steps before attempting it. Please read the
[etcd admin guide](https://github.com/coreos/etcd/blob/v2.2.1/Documentation/admin_guide.md)
before attempting it.

We can migrate from a single-master cluster to a multi-master cluster, but this is a complicated operation. It is easier to create a multi-master cluster using Kops (described [here](operations/high_availability.md)). If possible, try to plan this at time of cluster creation.

During this procedure, you will experience **downtime** on the API server, but
not on the end user services. During this downtime, existing pods will continue
to work, but you will not be able to create new pods and any existing pod that
dies will not be restarted.

## 1 - Backups

### a - Backup main etcd cluster

```bash
$ kubectl --namespace=kube-system get pods | grep etcd
etcd-server-events-ip-172-20-36-161.ec2.internal        1/1       Running   4          2h
etcd-server-ip-172-20-36-161.ec2.internal               1/1       Running   4          2h
$ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -it -- sh
/ # etcdctl backup --data-dir /var/etcd/data --backup-dir /var/etcd/backup
/ # mv /var/etcd/backup/ /var/etcd/data/
/ # exit
$ kubectl --namespace=kube-system get pod etcd-server-ip-172-20-36-161.ec2.internal -o json | jq '.spec.volumes[] | select(.name | contains("varetcdata")) | .hostPath.path'
"/mnt/master-vol-0ea119c15602cbb57/var/etcd/data"
$ ssh admin@<master-node>
admin@ip-172-20-36-161:~$ sudo -i
root@ip-172-20-36-161:~# mv /mnt/master-vol-0ea119c15602cbb57/var/etcd/data/backup /home/admin/
root@ip-172-20-36-161:~# chown -R admin: /home/admin/backup/
root@ip-172-20-36-161:~# exit
admin@ip-172-20-36-161:~$ exit
$ scp -r admin@<master-node>:backup/ .
```

### b - Backup event etcd cluster

```bash
$ kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -it -- sh
/ # etcdctl backup --data-dir /var/etcd/data-events --backup-dir /var/etcd/backup
/ # mv /var/etcd/backup/ /var/etcd/data-events/
/ # exit
$ kubectl --namespace=kube-system get pod etcd-server-events-ip-172-20-36-161.ec2.internal -o json | jq '.spec.volumes[] | select(.name | contains("varetcdata")) | .hostPath.path'
"/mnt/master-vol-0bb5ad222911c6777/var/etcd/data-events"
$ ssh admin@<master-node>
admin@ip-172-20-36-161:~$ sudo -i
root@ip-172-20-36-161:~# mv /mnt/master-vol-0bb5ad222911c6777/var/etcd/data-events/backup/ /home/admin/backup-events
root@ip-172-20-36-161:~# chown -R admin: /home/admin/backup-events/
root@ip-172-20-36-161:~# exit
admin@ip-172-20-36-161:~$ exit
$ scp -r admin@<master-node>:backup-events/ .
```

## 2 - Create instance groups

### a - Create new subnets

Add new subnets for your availability zones. Also create subnets of type Utility if they are defined in your configuration. 
```bash
kops edit cluster
```

Change the subnet section. An example might be: (**Adapt to your configuration!**)
```yaml
  - cidr: 172.20.32.0/19
    name: eu-west-1a
    type: Private
    zone: eu-west-1a
  - cidr: 172.20.64.0/19
    name: eu-west-1b
    type: Private
    zone: eu-west-1b
  - cidr: 172.20.96.0/19
    name: eu-west-1c
    type: Private
    zone: eu-west-1c
  - cidr: 172.20.0.0/22
    name: utility-eu-west-1a
    type: Utility
    zone: eu-west-1a
  - cidr: 172.20.4.0/22
    name: utility-eu-west-1b
    type: Utility
    zone: eu-west-1b
  - cidr: 172.20.8.0/22
    name: utility-eu-west-1c
    type: Utility
    zone: eu-west-1c
```

### b - Create new master instance group

Create 1 kops instance group for the first one of your new masters, in
a different AZ from the existing one. 

```bash
$ kops create instancegroup master-<availability-zone2> --subnet <availability-zone2> --role Master
```
Example:

```bash
$ kops create ig master-eu-west-1b --subnet eu-west-1b --role Master
```

 * ``maxSize`` and ``minSize`` should be 1,
 * only one zone should be listed.
 * adjust the machineType
 * adjust the image to the OS you are using

### c - Create third master instance group

Instance group for the third master, in a different AZ from the existing one, is
also required. However, real EC2 instance is not required until the second master launches.

```bash
$ kops create instancegroup master-<availability-zone3> --subnet <availability-zone3> --role Master
```

Example:

```bash
$ kops create ig master-eu-west-1c --subnet eu-west-1c --role Master
```

 * ``maxSize`` and ``minSize`` should be **0**,
 * if you are using etcd-manager, you just can leave the `maxSize` and `minSize` at **1**.
 * only one zone should be listed.
 * adjust the machineType
 * adjust the image to the OS you are using

### d - Reference the new masters in your cluster configuration

*kops will refuse to have only 2 members in the etcd clusters, so we have to
reference a third one, even if we have not created it yet.*

```bash
$ kops edit cluster example.com
```

 * In ``.spec.etcdClusters`` add 2 new members in each cluster, one for each new
 availability zone.

```yaml
    - instanceGroup: master-<availability-zone2>
      name: <availability-zone2-name>
    - instanceGroup: master-<availability-zone3>
      name: <availability-zone3-name>
```

Example:

```yaml
etcdClusters:
  - etcdMembers:
    - instanceGroup: master-eu-west-1a
      name: a
    - instanceGroup: master-eu-west-1b
      name: b
    - instanceGroup: master-eu-west-1c
      name: c
    name: main
  - etcdMembers:
    - instanceGroup: master-eu-west-1a
      name: a
    - instanceGroup: master-eu-west-1b
      name: b
    - instanceGroup: master-eu-west-1c
      name: c
    name: events
```

## 3 - Add a new master

### a - Add a new member to the etcd clusters

**The clusters will stop to work until the new member is started**.

```bash
$ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl member add etcd-<availability-zone2-name> http://etcd-<availability-zone2-name>.internal.example.com:2380 \
	&& kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 member add etcd-events-<availability-zone2-name> http://etcd-events-<availability-zone2-name>.internal.example.com:2381
```

Example:

```bash
$ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl member add etcd-b http://etcd-b.internal.example.com:2380 \
	&& kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 member add etcd-events-b http://etcd-events-b.internal.example.com:2381
```

### b - Launch the new master

```bash
$ kops update cluster example.com --yes
# wait for the new master to boot and initialize
$ ssh admin@<new-master>
admin@ip-172-20-116-230:~$ sudo -i
root@ip-172-20-116-230:~# systemctl stop kubelet
root@ip-172-20-116-230:~# systemctl stop protokube
```

Reinitialize the etcd instances:
* In both ``/etc/kubernetes/manifests/etcd-events.manifest`` and
``/etc/kubernetes/manifests/etcd.manifest``, edit the
``ETCD_INITIAL_CLUSTER_STATE`` variable to ``existing``.
* In the same files, remove the third non-existing member from
``ETCD_INITIAL_CLUSTER``.
* Delete the containers and the data directories. Mount paths can be determined with the help of ``docker inspect <container-id> | grep /mnt/master-vol``:

```bash
root@ip-172-20-116-230:~# docker stop $(docker ps | grep "k8s.gcr.io/etcd" | awk '{print $1}')
root@ip-172-20-116-230:~# rm -r /mnt/master-vol-03b97b1249caf379a/var/etcd/data-events/member/
root@ip-172-20-116-230:~# rm -r /mnt/master-vol-0dbfd1f3c60b8c509/var/etcd/data/member/
```

Launch them again:

```bash
root@ip-172-20-116-230:~# systemctl start kubelet
```

At this point, both etcd clusters should be healthy with two members:

```bash
$ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl member list
$ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl cluster-health
$ kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 member list
$ kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 cluster-health
```

If not, check ``/var/log/etcd.log`` for problems.

Restart protokube on the new master:

```bash
root@ip-172-20-116-230:~# systemctl start protokube
```

## 4 - Add the third master

### a - Edit instance group

Prepare to launch the third master instance:

```bash
$ kops edit instancegroup master-<availability-zone3>
```

* Replace ``maxSize`` and ``minSize`` values to **1**.

### b - Add a new member to the etcd clusters

 ```bash
 $ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl member add etcd-<availability-zone3-name> http://etcd-<availability-zone3-name>.internal.example.com:2380 \
	&& kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 member add etcd-events-<availability-zone3-name> http://etcd-events-<availability-zone3-name>.internal.example.com:2381
 ```

Example:

```bash
$ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl member add etcd-c http://etcd-c.internal.example.com:2380 \
	&& kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 member add etcd-events-c http://etcd-events-c.internal.example.com:2381
```

### c - Launch the third master

 ```bash
 $ kops update cluster example.com --yes
 # wait for the third master to boot and initialize
 $ ssh admin@<third-master>
 admin@ip-172-20-139-130:~$ sudo -i
 root@ip-172-20-139-130:~# systemctl stop kubelet
 root@ip-172-20-139-130:~# systemctl stop protokube
 ```

 Reinitialize the etcd instances:
 * In both ``/etc/kubernetes/manifests/etcd-events.manifest`` and
 ``/etc/kubernetes/manifests/etcd.manifest``, edit the
 ``ETCD_INITIAL_CLUSTER_STATE`` variable to ``existing``.
 * Delete the containers and the data directories. Mount paths can be determined with the help of ``docker inspect <container-id> | grep /mnt/master-vol``:

 ```bash
 root@ip-172-20-139-130:~# docker stop $(docker ps | grep "k8s.gcr.io/etcd" | awk '{print $1}')
 root@ip-172-20-139-130:~# rm -r /mnt/master-vol-019796c3511a91b4f/var/etcd/data-events/member/
 root@ip-172-20-139-130:~# rm -r /mnt/master-vol-0c89fd6f6a256b686/var/etcd/data/member/
 ```

 Launch them again:

 ```bash
 root@ip-172-20-139-130:~# systemctl start kubelet
 ```

 At this point, both etcd clusters should be healthy with three members:

 ```bash
 $ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl member list
 $ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl cluster-health
 $ kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 member list
 $ kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 cluster-health
 ```

 If not, check ``/var/log/etcd.log`` for problems.

 Restart protokube on the third master:

 ```bash
 root@ip-172-20-139-130:~# systemctl start protokube
 ```

## 5 - Cleanup

To be sure that everything runs smoothly and is setup correctly, it is advised
to terminate the masters one after the other (always keeping 2 of them up and
running). They will be restarted with a clean config and should join the others
without any problems.

While optional, this last step allows you to be sure that your masters are
fully configured by Kops and that there is no residual manual configuration.
If there is any configuration problem, they will be detected during this step
and not during a future upgrade or, worse, during a master failure.


## 6 - Restore (if migration to multi-master failed)

In case you failed to upgrade to multi-master you will need to restore from the backup you have taken previously.

Take extra care because kops will not start etcd and etcd-events with the same ID on <master-b> an/or <master-c> for example but will mix them (ex: etcd-b and etcd-events-c on <master-b> & etcd-c and etcd-events-b on <master-c> ); this can be double checked in Route53 where kops will create DNS records for your services.

If your 2nd spinned master failed and cluster becomes inconsistent edit the corresponding kops master instancegroup and switch ``MinSize`` and ``MaxSize`` to "0" and run an update on your cluster.

Next ssh into your primary master:

``systemctl stop kubelet``
``systemctl stop protokube``

Reinitialize the etcd instances:
* In both ``/etc/kubernetes/manifests/etcd-events.manifest`` and
``/etc/kubernetes/manifests/etcd.manifest``, add the
``ETCD_FORCE_NEW_CLUSTER`` variable with value ``1``.
* Delete the containers and the data directories while restoring also from previous backup:

```bash
root@ip-172-20-116-230:~# docker stop $(docker ps | grep "k8s.gcr.io/etcd" | awk '{print $1}')
root@ip-172-20-116-230:~# rm -r /mnt/master-vol-03b97b1249caf379a/var/etcd/data-events/member/
root@ip-172-20-116-230:~# rm -r /mnt/master-vol-0dbfd1f3c60b8c509/var/etcd/data/member/
root@ip-172-20-116-230:~# cp -R /mnt/master-vol-03b97b1249caf379a/var/etcd/data-events/backup/member  /mnt/master-vol-03b97b1249caf379a/var/etcd/data-events/
root@ip-172-20-116-230:~# cp -R /mnt/master-vol-0dbfd1f3c60b8c509/var/etcd/data/backup/member /mnt/master-vol-0dbfd1f3c60b8c509/var/etcd/data/
```

Now start back the services and watch for the logs:

``systemctl start kubelet``
``tail -f /var/log/etcd*`` # for errors, if no errors encountered re-start also protokube
``systemctl start protokube``

Test if your master is reboot-proof:

Go to EC2 and ``Terminate`` the instance and check if your cluster recovers (needed to discard any manual configurations and check that kops handles everything the right way).

Note! Would recommend also to use Amazon Lambda to take daily Snapshots of all your persistent volume so you can have from what to recover in case of failures.