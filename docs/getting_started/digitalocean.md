# Getting Started with kops on DigitalOcean

**WARNING**: digitalocean support on kops is currently **alpha** meaning it is in the early stages of development and subject to change, please use with caution.

## Requirements

* [kops version >= 1.9 installed](../install.md)
* [kubectl installed](../install.md)
* [DigitalOcean account](https://cloud.digitalocean.com/registrations/new)
* [DigitalOcean access token](https://www.digitalocean.com/community/tutorials/how-to-use-the-digitalocean-api-v2#how-to-generate-a-personal-access-token)
* [DigitalOcean Spaces bucket with API Keys](https://www.digitalocean.com/community/tutorials/how-to-create-a-digitalocean-space-and-api-key)
* [DigitalOcean domain](https://www.digitalocean.com/community/tutorials/an-introduction-to-digitalocean-dns#adding-a-domain). To point your domain registrar to DigitalOcean's nameservers follow [this guide](https://www.digitalocean.com/community/tutorials/how-to-point-to-digitalocean-nameservers-from-common-domain-registrars)
* [Setup your SSH key](https://www.digitalocean.com/community/tutorials/how-to-use-ssh-keys-with-digitalocean-droplets)

## Environment Variables

It is important to set the following environment variables:
```bash
export KOPS_STATE_STORE=do://<bucket-name> # where <bucket-name> is the name of the bucket you set earlier
export DIGITALOCEAN_ACCESS_TOKEN=<access-token>  # where <access-token> is the access token generated earlier to use the V2 API

# DigitalOCcean Spaces is S3 compatible so we just override some S3 configurations to talk to our bucket
export S3_ENDPOINT=nyc3.digitaloceanspaces.com # this can also be ams3.digitaloceanspaces.com or sgp1.digitaloceanspaces.com depending on where you created your Spaces bucket
export S3_ACCESS_KEY_ID=<access-key-id>  # where <access-key-id> is the Spaces API Access Key for your bucket
export S3_SECRET_ACCESS_KEY=<secret-key>  # where <secret-key> is the Spaces API Secret Key for your bucket

# this is required since DigitalOcean support is currently in alpha so it is feature gated
export KOPS_FEATURE_FLAGS="AlphaAllowDO"
```

## Creating a Single Master Cluster

In the following examples, `example.com` should be replaced with the DigitalOcean domain you created when going through the [Requirements](#requirements).
Note that you kops will only be able to successfully provision clusters in regions that support block storage (AMS3, BLR1, FRA1, LON1, NYC1, NYC3, SFO2, SGP1 and TOR1).

```bash
# coreos (the default) + flannel overlay cluster in tor1
kops create cluster --cloud=digitalocean --name=my-cluster.example.com --networking=flannel --zones=tor1 --ssh-public-key=~/.ssh/id_rsa.pub
kops update cluster my-cluster.example.com --yes

# ubuntu + weave overlay cluster in nyc1 using larger droplets
kops create cluster --cloud=digitalocean --name=my-cluster.example.com --image=ubuntu-16-04-x64 --networking=weave --zones=nyc1 --ssh-public-key=~/.ssh/id_rsa.pub --node-size=s-8vcpu-32gb
kops update cluster my-cluster.example.com --yes

# debian + flannel overlay cluster in ams3 using optimized droplets
kops create cluster --cloud=digitalocean --name=my-cluster.example.com --image=debian-9-x64 --networking=flannel --zones=ams3 --ssh-public-key=~/.ssh/id_rsa.pub --node-size=c-4
kops update cluster my-cluster.example.com --yes

# to delete a cluster
kops delete cluster my-cluster.example.com --yes
```

## Creating a Multi-Master HA Cluster

In the below example, `dev5.k8s.local` should be replaced with any cluster name that ends with `.k8s.local` such that a gossip based cluster is created.
Ensure the master-count is odd-numbered. A load balancer is created dynamically front-facing the master instances.

```bash
# coreos (the default) + flannel overlay cluster in tor1 with 3 master setup and a public load balancer.
kops create cluster --cloud=digitalocean --name=dev5.k8s.local --networking=cilium --api-loadbalancer-type=public --master-count=3 --zones=tor1 --ssh-public-key=~/.ssh/id_rsa.pub --yes

# to delete a cluster - this will also delete the load balancer associated with the cluster.
kops delete cluster dev5.k8s.local --yes
```

## Features Still in Development

kops for DigitalOcean currently does not support these features:

* rolling update for instance groups