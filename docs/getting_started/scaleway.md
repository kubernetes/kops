# Getting Started with kops on Scaleway

**WARNING**: Scaleway support on kOps is currently in **alpha**, which means that it is in the early stages of development and subject to change, please use with caution.

## Features

* Create, update and delete clusters
* Create, edit and delete instance groups 
* Migrating from single to multi-master

### Coming soon

* Load-balancers
* Scaleway DNS (to create clusters with a custom domain name)
* Private network

### Next features to implement

* `kops rolling-update`
* BareMetal servers
* [Terraform](https://github.com/scaleway/terraform-provider-scaleway) support
* [Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/scaleway) support

## Requirements

* [kops version >= 1.26 installed](../install.md)
* [kubectl installed](../install.md)
* [Scaleway credentials](https://www.scaleway.com/en/docs/generate-api-keys/) : you will need at least an access key, a secret key and a project ID.
* [S3 bucket and its credentials](https://www.scaleway.com/en/docs/storage/object/quickstart/) : the bucket's credentials may differ from the one used for provisioning the resources needed by the cluster. If you use a Scaleway bucket, you will need to prefix the bucket's name with `scw://` in the `KOPS_STATE_STORE` environment variable. For more information about buckets, see [here](../state.md)

### Optional

* [SSH key](https://www.scaleway.com/en/docs/configure-new-ssh-key/) : creating a cluster can be done without an SSH key, but it is required to update it. `id_rsa` and `id_ed25519` keys are supported


## Environment Variables

It is important to set the following [environment variables](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md):
```bash
export SCW_ACCESS_KEY="my-access-key"
export SCW_SECRET_KEY="my-secret-key"
export SCW_DEFAULT_PROJECT_ID="my-project-id"
export SCW_DEFAULT_REGION="fr-par"
export SCW_DEFAULT_ZONE="fr-par-1"
# Configure the bucket name to store kops state
export KOPS_STATE_STORE=scw://<bucket-name> # where <bucket-name> is the name of the bucket you set earlier
# Scaleway Object Storage is S3 compatible so we just override some S3 configurations to talk to our bucket
export S3_REGION=fr-par                     # or another scaleway region providing Object Storage
export S3_ENDPOINT=s3.$S3_REGION.scw.cloud  # define provider endpoint
export S3_ACCESS_KEY_ID="my-access-key"     # where <my-access-key> is the S3 API Access Key for your bucket
export S3_SECRET_ACCESS_KEY="my-secret-key" # where <my-secret-key> is the S3 API Secret Key for your bucket
# this is required since Scaleway support is currently in alpha so it is feature gated
export KOPS_FEATURE_FLAGS="Scaleway"
```

## Creating a Single Master Cluster

Note that for now you can only create a kops cluster in a single availability zone (fr-par-1, fr-par-2, fr-par-3, nl-ams-1, nl-ams-2, pl-waw-1, pl-waw-2).

```bash
# The default cluster uses ubuntu images on DEV1-M machines with cilium as Container Network Interface
  # This creates a cluster with the gossip DNS in zone fr-par-1
kops create cluster --cloud=scaleway --name=mycluster.k8s.local --zones=fr-par-1 --yes
```

### Editing your cluster
```bash
# Update a cluster
kops update cluster mycluster.k8s.local --yes
# Delete a cluster
kops delete cluster mycluster.k8s.local --yes
```

# Next steps

Now that you have a working _kops_ cluster, read through the [recommendations for production setups guide](production.md) to learn more about how to configure _kops_ for production workloads.