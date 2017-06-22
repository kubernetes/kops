# Download kops config spec file

KOPS operates off of a config spec file that is generated during the create phase.  It is uploaded to the amazon s3 bucket that is passed in during create.

If you download the config spec file on a running cluster that is configured the way you like it, you can just pass that config spec file in to the create command and have kops create the cluster for you , `kops create -f spec_file` in a completely unattended manor.

Let us say you create your cluster with the following configuration options:

```
declare -x KOPS_STATE_STORE=s3://k8s-us-west
declare -x CLOUD=aws
declare -x ZONE="us-west-1a"
declare -x MASTER_ZONES="us-west-1a"
declare -x NAME=k8s.example.com
declare -x K8S_VERSION=1.6.4
declare -x NETWORKCIDER="10.240.0.0/16"
declare -x MASTER_SIZE="t2.medium"
declare -x WORKER_SIZE="t2.large"
```
Next you call the kops command to create the cluster in your terminal:

```
kops create cluster $NAME              \
   --cloud=$CLOUD                      \
   --zones="$ZONE"                     \
   --kubernetes-version=$K8S_VERSION   \
   --master-zones="$MASTER_ZONES"      \
   --node-count=3                      \
   --node-size="$WORKER_SIZE"          \
   --master-size="$MASTER_SIZE"        \
   --network-cidr=${NETWORKCIDER}      \
   --dns-zone=ZVO7KL181S5AP            \
   --ssh-public-key=$HOME/.ssh/lab_no_password.pub
```

## kops command

You can simply use the kops command `kops get --name $NAME -o yaml > a_fun_name_you_will_remember.yml`

Note: for the above command to work the cluster NAME and the KOPS_STATE_STORE will have to be exported in your environment.  

For more information on how to use and modify the configurations see (here)[manifests_and_customizing_via_api.md].
