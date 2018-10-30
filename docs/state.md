# The State Store

kops has the notion of a 'state store'; a location where we store the configuration of your cluster.  State is stored
here not only when you first create a cluster, but also you can change the state and apply changes to a running cluster.

Eventually, kubernetes services will also pull from the state store, so that we don't need to marshal all our
configuration through a channel like user-data.  (This is currently done for secrets and SSL keys, for example,
though we have to copy the data from the state store to a file where components like kubelet can read them).

The state store uses kops's VFS implementation, so can in theory be stored anywhere. 
As of now the following state stores are supported: 
* Amazon AWS S3 (s3://)
* local filesystem (file://)
* Digital Ocean (do://)
* MemFS (memfs://)
* Google Cloud (gs://)
* Kubernetes (k8s://)
* OpenStack Swift (swift://)
* AliCloud (oss://)

The state store is just files; you can copy the files down and put them into git (or your preferred version control system).

## {statestore}/config

One of the most important files in the state store is the top-level config file.  This file stores the main
configuration for your cluster (instance types, zones, etc)\

When you run `kops create cluster`, we create a state store entry for you based on the command line options you specify. 
For example, when you run with `--node-size=m4.large`, we actually set a line in the configuration
that looks like `NodeMachineType: m4.large`.

The configuration you specify on the command line is actually just a convenient short-cut to
manually editing the configuration.  Options you specify on the command line are merged into the existing
configuration. If you want to configure advanced options, or prefer a text-based configuration, you
may prefer to just edit the config file with `kops edit cluster`.

Because the configuration is merged, this is how you can just specify the changed arguments when
reconfiguring your cluster - for example just `kops create cluster` after a dry-run.

## Moving state between S3 buckets

The state store can easily be moved to a different s3 bucket. The steps for a single cluster are as follows:
1. Recursively copy all files from `${OLD_KOPS_STATE_STORE}/${CLUSTER_NAME}` to `${NEW_KOPS_STATE_STORE}/${CLUSTER_NAME}` with `aws s3 sync` or a similar tool.
2. Update the `KOPS_STATE_STORE` environment variable to use the new S3 bucket.
3. Either run `kops edit cluster ${CLUSTER_NAME}` or edit the cluster manifest yaml file. Update `.spec.configBase` to reference the new s3 bucket.
4. Run `kops update cluster ${CLUSTER_NAME} --yes` to apply the changes to the cluster. Newly launched nodes will now retrieve their dependent files from the new S3 bucket. The files in the old bucket are now safe to be deleted.

Repeat for each cluster needing to be moved.

## State store configuration

There are a few ways to configure your state store.  In priority order:

+ command line argument `--state s3://yourstatestore`
+ environment variable `export KOPS_STATE_STORE=s3://yourstatestore`
+ config file `$HOME/.kops.yaml`
+ config file `$HOME/.kops/config`

### Configuration file example:

`$HOME/.kops/config` might look like this:

```
kops_state_store: s3://yourstatestore
```

## Cross Account State-store (AWS)

There are situations in which the entity executing kops to create the cluster is not in the same account as the owner of the state store bucket. In this case, you must explicitly grant the permission: `s3:getBucketLocation` to the ARN that is running kops.

You can use the following policy to guide your implementation:

```
{
    "Id": "123",
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "123",
            "Action": [
                "s3:GetBucketLocation"
            ],
            "Effect": "Allow",
            "Resource": "arn:aws:s3:::state-store-bucket",
            "Principal": {
                "AWS": [
                    "arn:aws:iam::123456789:user/kopsuser"
                ]
            }
        }
    ]
}
```