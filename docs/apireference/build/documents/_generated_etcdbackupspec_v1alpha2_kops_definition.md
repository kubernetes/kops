## EtcdBackupSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | EtcdBackupSpec



EtcdBackupSpec describes how we want to do backups of etcd

<aside class="notice">
Appears In:

<ul> 
<li><a href="#etcdclusterspec-v1alpha2-kops">EtcdClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
backupStore <br /> *string*    | BackupStore is the VFS path where we will read/write backup data
image <br /> *string*    | Image is the etcd backup manager image to use.  Setting this will create a sidecar container in the etcd pod with the specified image.

