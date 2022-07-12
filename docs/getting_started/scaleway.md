# Getting Started with kops on Scaleway

**WARNING**: scaleway support on kops is currently **alpha**, which means that scaleway support is in the early stages of development and subject to change, please use with caution.

## Scaleway requirements

* [kops version >= 1.18 installed](../install.md)
* [kubectl installed](../install.md)
* [Scaleway access/secret key](https://www.scaleway.com/en/docs/generate-api-keys/)
* [Setup your SSH key](https://www.scaleway.com/en/docs/configure-new-ssh-key/)

## Environment Variables

It is important to set the following [environment variables](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md):
```bash
# this is required since Scaleway support is currently in alpha so it is feature gated
export KOPS_FEATURE_FLAGS="Scaleway"
export SCW_ACCESS_KEY="my-access-key"
export SCW_SECRET_KEY="my-secret-key"
export SCW_DEFAULT_PROJECT_ID="my-project-id"
# Configure the bucket name to store kops state
export KOPS_STATE_STORE=scw://<bucket-name> # where <bucket-name> is the name of the bucket you set earlier
# Scaleway Object Storage is S3 compatible so we just override some S3 configurations to talk to our bucket
export S3_REGION=fr-par                     # or another scaleway region providing Object Storage
export S3_ENDPOINT=s3.$S3_REGION.scw.cloud  # define provider endpoint
export S3_ACCESS_KEY_ID="my-access-key"     # where <access-key-id> is the Spaces API Access Key for your bucket
export S3_SECRET_ACCESS_KEY="my-secret-key" # where <secret-key> is the Spaces API Secret Key for your bucket
```

## Creating a Single Master Cluster

In the following examples, `example.com` should be replaced with the Scaleway domain you created when going through the [Requirements](#requirements). // TODO(Mia-Cross): fix broken anchor
Note that you kops will only be able to successfully provision clusters in regions that support block storage (AMS3, BLR1, FRA1, LON1, NYC1, NYC3, SFO2, SGP1 and TOR1).

```bash
# debian (the default) + flannel overlay cluster in fr-par-1 using default instance type
kops create cluster --cloud=scaleway --name=mycluster.k8s.local --networking=flannel --zones=fr-par-1 --ssh-public-key=~/.ssh/id_ed25519.pub
kops update cluster my-cluster.example.com --yes
# ubuntu + weave overlay cluster in nl-ams-1 using GP1-S instance type.
kops create cluster --cloud=scaleway --name=mycluster.k8s.local --image=ubuntu_focal --networking=weave --zones=nl-ams-1 --ssh-public-key=~/.ssh/id_ed25519.pub --node-size=gp1-s
kops update cluster my-cluster.example.com --yes
# to delete a cluster
kops delete cluster my-cluster.example.com --yes
```

## Creating a Multi-Master HA Cluster

In the below example, `dev5.k8s.local` should be replaced with any cluster name that ends with `.k8s.local` such that a gossip based cluster is created.
Ensure the master-count is odd-numbered. A load balancer is created dynamically front-facing the master instances.

```bash
# debian (the default) + flannel overlay cluster in tor1 with 3 master setup and a public load balancer.
kops create cluster --cloud=scaleway --name=dev5.k8s.local --networking=cilium --api-loadbalancer-type=public --master-count=3 --zones=fr-par-1 --ssh-public-key=~/.ssh/id_rsa.pub --yes
# to delete a cluster - this will also delete the load balancer associated with the cluster.
kops delete cluster dev5.k8s.local --yes
```

# Next steps

Now that you have a working _kops_ cluster, read through the [recommendations for production setups guide](production.md) to learn more about how to configure _kops_ for production workloads.