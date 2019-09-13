# Download kops config spec file

KOPS operates off of a config spec file that is generated during the create phase.  It is uploaded to the amazon s3 bucket that is passed in during create.

If you download the config spec file on a running cluster that is configured the way you like it, you can just pass that config spec file in to the create command and have kops create the cluster for you , `kops create -f spec_file` in a completely unattended manner.

Let us say you create your cluster with the following configuration options:

```
export KOPS_STATE_STORE=s3://k8s-us-west
export CLOUD=aws
export ZONE="us-west-1a"
export MASTER_ZONES="us-west-1a"
export NAME=k8s.example.com
export K8S_VERSION=1.6.4
export NETWORKCIDR="10.240.0.0/16"
export MASTER_SIZE="m3.large"
export WORKER_SIZE="m4.large"
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
   --network-cidr=${NETWORKCIDR}       \
   --dns-zone=ZVO7KL181S5AP            \
   --ssh-public-key=$HOME/.ssh/lab_no_password.pub
```

## kops command

You can simply use the kops command `kops get --name $NAME -o yaml > a_fun_name_you_will_remember.yml`

Note: for the above command to work the cluster NAME and the KOPS_STATE_STORE will have to be exported in your environment.

For more information on how to use and modify the configurations see [here](../manifests_and_customizing_via_api.md).

## Managing instance groups

You can also manage instance groups in separate YAML files as well.  The command `kops get --name $NAME -o yaml > $NAME.yml` exports the entire cluster.  An option is to have a YAML file for the cluster, and individual YAML files for the instance groups.  This allows you to do stuff like:

```shell
if ! kops get cluster --name "$NAME"; then
    kops create -f "kops/$CLUSTER/$REGION.yaml"
else
 kops replace -f "kops/$CLUSTER/$REGION.yaml"
fi

for ig in kops/$CLUSTER/instancegroup/*; do
  if ! kops get ig --name "$NAME" "$(basename "$ig")"; then
    kops create -f "$ig"
  else
    kops replace -f "$ig"
  fi
done
```
