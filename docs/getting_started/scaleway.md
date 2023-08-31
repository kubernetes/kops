# Getting Started with kops on Scaleway

**WARNING**: Scaleway support on kOps is currently in **alpha**, which means that it is in the early stages of development and subject to change, please use with caution.

## Features

* Create, update and delete clusters
  * [Rolling-update](../operations/rolling-update.md)
* Create, edit and delete instance groups --> Editable fields include but are not limited to:
  * Instance image
  * Instance size (also called commercial type)
* Migrating from single to multi-master

### Coming soon

* [Terraform](https://github.com/scaleway/terraform-provider-scaleway) support
* Private network

### Next features to implement

* [Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/scaleway) support
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
