# Etcd Administration Tasks

## etcd-manager

[etcd-manager](https://github.com/kubernetes-sigs/etcdadm/tree/master/etcd-manager) is a kubernetes-sigs project that kOps uses to manage
etcd.

It handles graceful upgrades of etcd, TLS, and backups. If a Kubernetes cluster needs more redundant control plane, it also takes care of resizing the etcd cluster.

## Backups

Backups and restores of etcd on kOps are covered in [etcd_backup_restore_encryption.md](etcd_backup_restore_encryption.md)

## Direct Data Access

It's not typically necessary to view or manipulate the data inside of etcd directly with etcdctl, because all operations usually go through kubectl commands. However, it can be informative during troubleshooting, or just to understand kubernetes better. Here are the steps to accomplish that on kOps.

1\. Determine which version of etcd is running

```bash
kops get cluster --full -o yaml
```

Look at the `etcdCluster` configuration's `version` for the given cluster.


2\. Connect to an etcd-manager pod

```bash
CONTAINER=$(kubectl get pods -n kube-system | grep etcd-manager-main | head -n 1 | awk '{print $1}')
kubectl exec -it -n kube-system $CONTAINER -- sh
```

``

3\. Run etcdctl

```bash
ETCD_VERSION=3.5.1
ETCDDIR=/opt/etcd-v$ETCD_VERSION-linux-amd64 # Replace with arm64 if you are running an arm control plane
CERTDIR=/rootfs/srv/kubernetes/kube-apiserver/
alias etcdctl="ETCDCTL_API=3 $ETCDDIR/etcdctl --cacert=$CERTDIR/etcd-ca.crt --cert=$CERTDIR/etcd-client.crt --key=$CERTDIR/etcd-client.key --endpoints=https://127.0.0.1:4001"
```

Test the client by running the following:

```bash
etcdctl member list
```

If successful, this should output the members of the etcd cluster.
