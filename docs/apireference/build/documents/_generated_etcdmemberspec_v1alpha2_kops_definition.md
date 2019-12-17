## EtcdMemberSpec v1alpha2 kops

Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | EtcdMemberSpec



EtcdMemberSpec is a specification for a etcd member

<aside class="notice">
Appears In:

<ul> 
<li><a href="#etcdclusterspec-v1alpha2-kops">EtcdClusterSpec kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
encryptedVolume <br /> *boolean*    | EncryptedVolume indicates you want to encrypt the volume
instanceGroup <br /> *string*    | InstanceGroup is the instanceGroup this volume is associated
kmsKeyId <br /> *string*    | KmsKeyId is a AWS KMS ID used to encrypt the volume
name <br /> *string*    | Name is the name of the member within the etcd cluster
volumeIops <br /> *integer*    | If volume type is io1, then we need to specify the number of Iops.
volumeSize <br /> *integer*    | VolumeSize is the underlying cloud volume size
volumeType <br /> *string*    | VolumeType is the underlying cloud storage class

