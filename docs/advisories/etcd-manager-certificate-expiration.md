# etcd-manager Certificate Expiration

etcd-manager configures certificates for TLS communication between kube-apiserver and etcd, as well as between etcd members.
These certificates are signed by the cluster CA and are valid for a duration of 1 year.

Affected versions of etcd-manager currently do NOT automatically rotate these certificates before expiration.
If these certificates are not rotated prior to their expiration, Kubernetes apiserver will become inaccessible and your control-plane will experience downtime.

## How do I know if I'm affected?

Clusters are affected by this issue if they're using a version of etcd-manager < `3.0.20200428`.
The etcd-manager version is set automatically based on the Kops version.
These Kops versions are affected:

* Kops 1.10.0-alpha.1 through 1.15.2
* Kops 1.16.0-alpha.1 through 1.16.1
* Kops 1.17.0-alpha.1 through 1.17.0-beta.1
* Kops 1.18.0-alpha.1 through 1.18.0-alpha.2

The issue can be confirmed by checking for the existence of etcd-manager pods and observing their image tags:

```bash
kubectl get pods -n kube-system -l k8s-app=etcd-manager-main \
  -o jsonpath='{range .items[*].spec.containers[*]}{.image}{"\n"}{end}'
```

* If this outputs `kopeio/etcd-manager` images with tags older than `3.0.20200428`, the cluster is affected.
* If this outputs an image other than `kopeio/etcd-manager`, the cluster may be affected.
* If this does does not output anything or outputs `kopeio/etcd-manager` images with tags >= `3.0.20200428`, the cluster is not affected.

## Solution

Upgrade etcd-manager. etcd-manager version >= `3.0.20200428` manages certificate lifecycle and will automatically request new certificates before expiration.

We have two suggested workflows to upgrade etcd-manager in your cluster. While both workflows require a rolling-update of the master nodes, neither require control-plane downtime (if the clusters have highly available masters).

1. Upgrade to Kops 1.15.3, 1.16.2, 1.17.0-beta.2, or 1.18.0-alpha.3.
   This is the recommended approach.
   Follow the normal steps when upgrading Kops and confirm the etcd-manager image will be updated based on the output of `kops update cluster`.
   ```
   kops update cluster --yes
   kops rolling-update cluster --instance-group-roles=Master
   ```
2. Another solution is to override the etcd-manager image in the ClusterSpec.
   The image will be set in two places, one for each etcdCluster (main and events).
   ```
   kops edit cluster $CLUSTER_NAME
   # Set `spec.etcdClusters[*].manager.image` to `kopeio/etcd-manager:3.0.20200428`
   kops update cluster # confirm the image is being updated
   kops update cluster --yes
   kops rolling-update cluster --instance-group-roles=Master --force
   ```

## Hack/Workaround

**This will not prevent the issue from occurring again, only reset the 1 year expiration.**

A rolling-update of the master nodes can be avoided by manually deleting the certificates to force them to be recreated.
Perform these steps on each master instance at a time.

1. SSH into the instance and delete the pki directory from each volume mount.
   ```
   # Mount paths may vary between cloud providers and OS distributions
   sudo rm -rf /mnt/master-vol-*/pki
   ```
2. Restart the two etcd-manager containers. Alternatively you can reboot the instance or terminate the instance in the autoscaling group.
   ```
   sudo docker restart $(sudo docker ps -q -f "label=io.kubernetes.container.name=etcd-manager")
   ```
   When the container restarts, etcd-manager will reissue the certs and rejoin the etcd cluster. etcd membership can be confirmed with `etcdctl endpoint status` by following the instructions in the [docs](https://kops.sigs.k8s.io/operations/etcd_administration/#direct-data-access).