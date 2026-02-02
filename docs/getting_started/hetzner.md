# Getting Started with kOps on Hetzner Cloud

**WARNING**: Hetzner Cloud support on kOps is currently in **BETA**, which means it is in good shape and could be used for production.
However, it is not as rigorously tested as the stable cloud providers, and there are some features that might be missing.

## Requirements
* kOps version >= 1.24
* kubectl version >= 1.23
* Hetzner Cloud [account](https://accounts.hetzner.com/login)
* Hetzner Cloud [API token](https://docs.hetzner.cloud/reference/cloud#description/authentication)
* Hetzner Cloud [S3 credentials](https://docs.hetzner.com/storage/object-storage/faq/s3-credentials/)
* SSH public and private keys

## Environment Variables

It is important to set the following environment variables:
```bash
export HCLOUD_TOKEN=<token>
export S3_ACCESS_KEY_ID=<acces-key>
export S3_SECRET_ACCESS_KEY=<secret-key>
export S3_ENDPOINT=https://fsn1.your-objectstorage.com
export KOPS_STATE_STORE=hos://<bucket-name>
```

Some S3 compatible stores may also require setting the region:
```bash
export S3_REGION=<region>
```

## Creating a Single Master Cluster

```bash
# create a ubuntu 24.04 + Cilium cluster in fsn1
kops create cluster --name=my.k8s \
  --ssh-public-key=~/.ssh/id_ed25519.pub --cloud=hetzner --zones=fsn1 \
  --image=ubuntu-24.04 --networking=calico --network-cidr=10.10.0.0/16 \
  --control-plane-size cx23 --node-size cx23
kops update cluster my.k8s --yes

# update a cluster
kops update cluster --name=my.k8s
kops update cluster --name=my.k8s --yes
kops rolling-update cluster --name=my.k8s
kops rolling-update cluster --name=my.k8s --yes

# validate a cluster
kops validate cluster --name=my.k8s

# delete a cluster
kops delete cluster --name=my.k8s
kops delete cluster --name=my.k8s --yes

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
