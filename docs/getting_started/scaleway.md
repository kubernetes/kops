# Getting Started with kops on Scaleway

**WARNING**: Scaleway support on kOps is currently in **alpha**, which means that it is in the early stages of development and subject to change, please use with caution.

## Features

* Create, update and delete clusters
  * [Rolling-update](../operations/rolling-update.md)
* Create, edit and delete instance groups --> Editable fields include but are not limited to:
  * Instance image
  * Instance size (also called commercial type)
* Migrating from single to multi-master

### Next features to implement

* [Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/scaleway) support
* Private network
* BareMetal servers

## Requirements

* [kops version >= 1.26 installed](../install.md)
* [kubectl installed](../install.md)
* [Scaleway credentials](https://www.scaleway.com/en/docs/generate-api-keys/) : you will need at least an access key, a secret key and a project ID.
* [S3 bucket and its credentials](https://www.scaleway.com/en/docs/storage/object/quickstart/) : the bucket's credentials may differ from the one used for provisioning the resources needed by the cluster. If you use a Scaleway bucket, you will need to prefix the bucket's name with `scw://` in the `KOPS_STATE_STORE` environment variable. For more information about buckets, see [here](../state.md)

### Optional

* [SSH key](https://www.scaleway.com/en/docs/configure-new-ssh-key/) : creating a cluster can be done without an SSH key, but it is required to update it. `id_rsa` and `id_ed25519` keys are supported
* [Domain name](https://www.scaleway.com/en/docs/network/domains-and-dns/quickstart/) : if you want to host your cluster on your own domain, you will have to register it with Scaleway.

## Environment Variables

### Enable Scaleway

Since Scaleway support is currently in alpha, it is feature gated and you will need to set this variable:
```bash
export KOPS_FEATURE_FLAGS="Scaleway"
```

### Scaleway Credentials

To be able to use Scaleway APIs, it is required to set up your credentials in the [environment](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md).
You have two ways to pass your credentials:

1. If you are already familiar with Scaleway's DevTools, then you probably have a config file (its default location is `$HOME/.config/scw/config.yaml`).
If so, you can use the profile of your choice by setting:
```bash
export SCW_PROFILE="my-profile"
```
2. If not, you can directly set the credentials in your environment:

```bash
export SCW_ACCESS_KEY="my-access-key"
export SCW_SECRET_KEY="my-secret-key"
export SCW_DEFAULT_PROJECT_ID="my-project-id"
```

**NB:** Keep in mind that the profile is checked first and the environment second, so if you set both, the environment variables will override the information in the config file (profile).

### S3 Bucket credentials

For kOps to be able to read and write configuration to the state-store bucket, you'll need to set up the following environment variables. The credentials can be the same as in the previous section, but they don't have to be.
```bash
export KOPS_STATE_STORE=scw://<bucket-name> # where <bucket-name> is the name of the bucket you set earlier
# Scaleway Object Storage is S3 compatible so we just override some S3 configurations to talk to our bucket
export S3_REGION=fr-par                     # or another scaleway region providing Object Storage
export S3_ENDPOINT=s3.$S3_REGION.scw.cloud  # define provider endpoint
export S3_ACCESS_KEY_ID="my-access-key"     # where <my-access-key> is the S3 API Access Key for your bucket
export S3_SECRET_ACCESS_KEY="my-secret-key" # where <my-secret-key> is the S3 API Secret Key for your bucket
```

## Creating a Single Master Cluster

```bash
# This creates a cluster with no DNS in zone fr-par-1
kops create cluster --cloud=scaleway --name=my.cluster --zones=fr-par-1 --dns=none --yes
# This creates a cluster with the Scaleway DNS (on a domain name that you own and have registered with Scaleway) in zone pl-waw-1
kops create cluster --cloud=scaleway --name=mycluster.mydomain.com --zones=pl-waw-1 --yes 
# This creates a cluster with the gossip DNS in zone nl-ams-2. This is not recommended since the no-DNS option is available because it is more secure.
kops create cluster --cloud=scaleway --name=mycluster.k8s.local --zones=nl-ams-2 --yes
```
These basic commands create a cluster with default parameters:
- Container Network Interface = `cilium`. To change it, set the flag `--networking=calico`. To see the list of supported CNIs, check the [networking page](../networking.md)
- Instance type = `DEV1-M`. To change it, set the flag `--node-size=PRO2-XS` and/or `--control-plane-size=PRO2-XS`
- Instance image = `ubuntu_jammy`. To change it, set the flag `--node-image=ubuntu_focal` and/or `--control-plane-image=ubuntu_focal`

**NB:** For now, you can only create a kops cluster in a single availability zone (fr-par-1, fr-par-2, fr-par-3, nl-ams-1, nl-ams-2, nl-ams-3, pl-waw-1, pl-waw-2).


# Next steps

Now that you have a working _kops_ cluster, read through the [recommendations for production setups guide](production.md) to learn more about how to configure _kops_ for production workloads.
For example, you can migrate your cluster to [high-availability](../operations/high_availability.md).

### Editing your cluster

```bash
# This opens the cluster's configuration file in a text editor for you to make the desired changes
kops edit cluster mycluster.k8s.local --state=scw://my-state-store
# This applies the changes
kops update cluster mycluster.k8s.local --yes
```

### Deleting your cluster

```bash
kops delete cluster mycluster.k8s.local --yes
```

# Terraform support

kOps offers the possibility to generate a Terraform configuration corresponding to the cluster that would have been created directly otherwise.

You can find more information on the dedicated page on [kOps Terraform support](../terraform.md) or [Scaleway's Terraform provider's documentation](https://github.com/scaleway/terraform-provider-scaleway).

## For clusters without load-balancers

This concerns clusters using Scaleway DNS. For this type of clusters, things are pretty simple.

```bash
kops create cluster --cloud=scaleway --name=mycluster.mydomain.com --zones=fr-par-1 --target=terraform --out=$OUTPUT_DIR
cd $OUTPUT_DIR
terraform init
terraform apply
```
kOps will generate a `kubernetes.tf` file in the output directory of your choice, you just have to initialize Terraform and apply the configuration.
NB: keep in mind that every new call to kOps with the flags `--target=terraform --out=$OUTPUT_DIR` will overwrite `kubernetes.tf` so any changes that you made to it will be lost.

## For clusters with load-balancers

This concerns clusters using no DNS and gossip DNS. For these types of cluster, a small trick is needed because kOps doesn't know the IPs of the load-balancer at the time of writing the instances' cloud-init configuration, so we will have to run an update, then a rolling-update.

### Creating a valid cluster

```bash
kops create cluster --cloud=scaleway --name=my.cluster --zones=fr-par-1 --target=terraform --out=$OUTPUT_DIR
cd $OUTPUT_DIR
terraform init
terraform apply
# Now that the load-balancer is up, we update the cluster to integrate its IP to the instances' configuration
kops update cluster my.cluster --target=terraform --out=$OUTPUT_DIR
# Then we replace the instances's for them to reboot with the new configuration (the --cloudonly flag is needed because the cluster can't be validated at this point)
kops rolling-update cluster my.cluster --cloudonly --yes
```

### Keeping the Terraform state consistent after a rolling-update

Now that the instances have been replaced by the rolling-update, your cluster can now be validated.
However, since resources have changed outside of Terraform, the state is now invalid. If you need to keep the state consistent with the cluster, you should import the new instances. This can be achieved with this script:

```bash
# First we need to retrieve the names of the instances
cd "$OUTPUT_DIR" || exit
TF_SERVERS=($(grep 'resource "scaleway_instance_server"' < kubernetes.tf | awk '{print $3}' | cut -d'"' -f 2))
# Then we get the zone for the import
ZONE=$(terraform output zone | cut -d '"' -f2)
# And for each instance:
for SERVER in "${TF_SERVERS[@]}"; do
  # We remove the stale instance from the state
  terraform state rm scaleway_instance_server.$SERVER
  # We fetch its new ID
  NEW_SERVER_ID=$(scw instance server list zone=$ZONE name=$SERVER -o template="{{ .ID }}")
  if [ "$NEW_SERVER_ID" == "" ]; then
    echo "could not find new ID of the server $SERVER"
  fi
  # We import the new instance in the state
  terraform import scaleway_instance_server.$SERVER $ZONE/$NEW_SERVER_ID
done
```

NB: for the script to run, you will need to have the [Scaleway CLI](https://github.com/scaleway/scaleway-cli) installed. You can also fetch the IDs of the new instances manually in the [Scaleway Console](https://console.scaleway.com) but if you have a lot of them this may not be practical.
If you need help with the CLI, these resources might help:
* [Installing the CLI](https://github.com/scaleway/scaleway-cli#readme)
* [Tutorial for setting up the CLI and managing instances with it](https://www.scaleway.com/en/docs/compute/instances/api-cli/creating-managing-instances-with-cliv2/) 
