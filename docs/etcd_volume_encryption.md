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

At this point you must edit the Key Users list to add the `masters` role. 
This has to be done before the master(s) attempt to to mount the volumes. 
You should have at least a several minute window between the `masters` role being created by kops and the master(s) 
mounting the volume, but if you somehow miss this window, you can just delete the master(s) and the ASG will kick in 
and once new masters start up they should be able to mount successfully.

Adding the `masters` role to the Key Users group via the AWS Console:

1. Navigate to the IAM page
2. Click on `Encryption keys` on the left sidebar
3. Select the KMS key that you are using to encrypt the etcd volumes
4. Scroll down to Key Users and click Add
5. Select the `masters.<your.domain>` role and click Attach
