# Getting Started with kOps on Hetzner Cloud

**WARNING**: Hetzner Cloud support on kOps is currently in **beta**, which means it is in good shape and could be used for production.
However, it is not as rigorously tested as the stable cloud providers and there are some features that might be missing.

## Requirements
* kOps version >= 1.24
* kubectl version >= 1.23
* Hetzner Cloud [account](https://accounts.hetzner.com/login)
* Hetzner Cloud [token](https://docs.hetzner.cloud/#authentication)
* SSH public and private keys
* S3 compatible object storage (like [MinIO](https://docs.min.io/minio/baremetal/security/minio-identity-management/user-management.html))

## Environment Variables

It is important to set the following environment variables:
```bash
export HCLOUD_TOKEN=<token>
export S3_ENDPOINT=<endpoint>
export S3_ACCESS_KEY_ID=<acces-key>
export S3_SECRET_ACCESS_KEY=<secret-key>
export KOPS_STATE_STORE=s3://<bucket-name>
```

Some S3 compatible stores may also require to set the region:
```bash
export S3_REGION=<region>
```

## Creating a Single Master Cluster

In the following examples, `example.k8s.local` is a [gossip-based DNS ](../gossip.md) cluster name.

```bash
# create a ubuntu 20.04 + calico cluster in fsn1
kops create cluster --name=my-cluster.example.k8s.local \
  --ssh-public-key=~/.ssh/id_rsa.pub --cloud=hetzner --zones=fsn1 \
  --image=ubuntu-20.04 --networking=calico --network-cidr=10.10.0.0/16 
kops update cluster my-cluster.example.k8s.local --yes

# create a ubuntu 20.04 + calico cluster in fsn1 with CPU optimized servers
kops create cluster --name=my-cluster.example.k8s.local \
  --ssh-public-key=~/.ssh/id_rsa.pub --cloud=hetzner --zones=fsn1 \
  --image=ubuntu-20.04 --networking=calico --network-cidr=10.10.0.0/16 \
  --node-size cpx31
kops update cluster --name=my-cluster.example.k8s.local --yes

# update a cluster
kops update cluster --name=my-cluster.example.k8s.local
kops update cluster --name=my-cluster.example.k8s.local --yes
kops rolling-update cluster --name=my-cluster.example.k8s.local
kops rolling-update cluster --name=my-cluster.example.k8s.local --yes

# validate a cluster
kops validate cluster --name=my-cluster.example.k8s.local

# delete a cluster
kops delete cluster --name=my-cluster.example.k8s.local
kops delete cluster --name=my-cluster.example.k8s.local --yes

# export kubecfg
# See https://kops.sigs.k8s.io/cli/kops_export_kubeconfig/#examples. 

# update a cluster
# See https://kops.sigs.k8s.io/operations/updates_and_upgrades/#manual-update.
```

## Features Still in Development

kOps for Hetzner Cloud currently does not support the following features:

* Autoscaling using [Cluster Autoscaler](https://github.com/hetznercloud/autoscaler)
* Terraform support using [terraform-provider-hcloud](https://github.com/hetznercloud/terraform-provider-hcloud) 

## Next steps

Now that you have a working kOps cluster, read through the recommendations for [production setups guide](production.md) to learn more about how to configure kOps for production workloads.
