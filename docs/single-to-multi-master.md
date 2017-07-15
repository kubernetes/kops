# Migrating from single to multi-master

This document describes how to go from a single-master cluster (created by kops)
to a multi-master cluster.

## Warnings

This is a risky procedure that **can lead to data-loss** in the etcd cluster.
Please follow all the backup steps before attempting it. Please read the
[etcd admin guide](https://github.com/coreos/etcd/blob/v2.2.1/Documentation/admin_guide.md)
before attempting it.

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

### a - Create new master instance group

Create 1 kops instance group for the first one of your new masters, in
a different AZ from the existing one.

```bash
$ kops create instancegroup master-<availability-zone2> --subnet <availability-zone2> --role Master
```

 * ``maxSize`` and ``minSize`` should be 1,
 * only one zone should be listed.

### b - Create third master instance group

Instance group for the third master, in a different AZ from the existing one, is
also required. However, real EC2 instance is not required until the second master launches.

```bash
$ kops create instancegroup master-<availability-zone3> --subnet <availability-zone3> --role Master
```

 * ``maxSize`` and ``minSize`` should be **0**,
 * only one zone should be listed.

### c - Reference the new masters in your cluster configuration

*kops will refuse to have only 2 members in the etcd clusters, so we have to
reference a third one, even if we have not created it yet.*

```bash
$ kops edit cluster example.com
```

 * In ``.spec.etcdClusters`` 2 new members in each cluster, one for each new
 availability zone.

```yaml
    - instanceGroup: master-<availability-zone2>
      name: <availability-zone2>
    - instanceGroup: master-<availability-zone3>
      name: <availability-zone3>
```

## 3 - Add a new master

### a - Add a new member to the etcd clusters

**The clusters will stop to work until the new member is started**.

```bash
$ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl member add etcd-<availability-zone2> http://etcd-<availability-zone2>.internal.example.com:2380
$ kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 member add etcd-events-<availability-zone2> http://etcd-events-<availability-zone2>.internal.example.com:2381
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
* Delete the containers and the data directories:

```bash
root@ip-172-20-116-230:~# docker stop $(docker ps | grep "etcd:2.2.1" | awk '{print $1}')
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
 $ kubectl --namespace=kube-system exec etcd-server-ip-172-20-36-161.ec2.internal -- etcdctl member add etcd-<availability-zone3> http://etcd-<availability-zone3>.internal.example.com:2380
 $ kubectl --namespace=kube-system exec etcd-server-events-ip-172-20-36-161.ec2.internal -- etcdctl --endpoint http://127.0.0.1:4002 member add etcd-events-<availability-zone3> http://etcd-events-<availability-zone3>.internal.example.com:2381
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
 * Delete the containers and the data directories:

 ```bash
 root@ip-172-20-139-130:~# docker stop $(docker ps | grep "etcd:2.2.1" | awk '{print $1}')
 root@ip-172-20-139-130:~# rm -r /mnt/master-vol-019796c3511a91b4f//var/etcd/data-events/member/
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
