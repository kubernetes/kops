# Etcd Volume Encryption

You must configure etcd volume encryption before bringing up your cluster. You cannot add etcd volume encryption to an already running cluster.

## Encrypting Etcd Volumes Using the Default AWS KMS Key

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

## Encrypting Etcd Volumes Using a Custom AWS KMS Key

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
