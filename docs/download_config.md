# Download kops config spec file
KOPS operates off of a config spec file that is generated during the create phase.  It is uploaded to the amazon s3 bucket that is passed in during create.

If you download the config spec file on a running cluster that is configured the way you like it, you can just pass that config spec file in to the create command and have kops create the cluster for you , `kops create -f spec_file` in a completely unattended manor.

## How to download the config spec file.
Let's say you create your cluster with the following configuration options:

```
declare -x KOPS_STATE_STORE=s3://k8s-us-west
declare -x CLOUD=aws
declare -x ZONE="us-west-1a"
declare -x MASTER_ZONES="us-west-1a"
declare -x NAME=westtest.c.foo.com
declare -x K8S_VERSION=1.4.6
declare -x NETWORKCIDER="10.240.0.0/16"
declare -x MASTER_SIZE="t2.medium"
declare -x WORKER_SIZE="t2.large"
```
Next you call the kops command to create the cluster in your terminal:

```
  kops create cluster $NAME       \
    --cloud=$CLOUD                      \
    --zones="$ZONE"                     \
    --kubernetes-version=$K8S_VERSION   \
    --master-zones="$MASTER_ZONES"       \
    --node-count=3                      \
    --node-size="$WORKER_SIZE"          \
    --master-size="$MASTER_SIZE"        \
    --network-cidr=${NETWORKCIDER}      \
    --dns-zone=ZVO7KL181S5AP \
    --ssh-public-key=/Users/foo/.ssh/lab_no_password.pub
```

Your spec file will be located in your s3 bucket, in the location `s3://k8s-us-west/$NAME/config`.  Using the above as an example the config file is located at `s3://k8s-us-west/westtest.c.foo.com/config`

To download this file via the aws cli you can use the aws s3 copy command.  Using the above cluster as an example the command would be

`aws s3 cp $KOPS_STATE_STORE/$NAME/config ~/a_fun_name_you_will_remember.yml`  

