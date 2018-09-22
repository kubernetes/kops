# Getting Started with kops on OpenStack

**WARNING**: OpenStack support on kops is currently **alpha** meaning it is in the early stages of development and subject to change, please use with caution.

## Create config file
The config file contains the OpenStack credentials required to create a cluster. The config file has the following format:
```ini
[Default]
identity=<OS_AUTH_URL>
user=mk8s=<OS_USERNAME>
password=<OS_PASSWORD>
domain_name=<OS_USER_DOMAIN_NAME>
tenant_id=<OS_PROJECT_ID>

[Swift]
service_type=object-store
region=<OS_REGION_NAME>

[Cinder]
service_type=volumev3
region=<OS_REGION_NAME>

[Neutron]
service_type=network
region=<OS_REGION_NAME>

[Nova]
service_type=compute
region=<OS_REGION_NAME>
```

## Environment Variables

It is important to set the following environment variables:
```bash
export OPENSTACK_CREDENTIAL_FILE=<config-file> # where <config-file> is the path of the config file
export KOPS_STATE_STORE=swift://<bucket-name> # where <bucket-name> is the name of the Swift container to use for kops state

# TODO(lmb): Add a feature gate for OpenStack
# this is required since OpenStack support is currently in alpha so it is feature gated
# export KOPS_FEATURE_FLAGS="AlphaAllowOpenStack"
```

## Creating a Cluster

```bash
# coreos (the default) + flannel overlay cluster in Default
kops create cluster --cloud=openstack --name=my-cluster.k8s.local --networking=flannel --zones=Default --network-cidr=192.168.0.0/16
# Not implemented yet...
# kops update cluster my-cluster.k8s.local --yes

# to delete a cluster
# Not implemented yet...
# kops delete cluster my-cluster.k8s.local --yes
```

## Features Still in Development

kops for OpenStack currently does not support these features:
* cluster create (servers, servergroups, load balancers, and DNS are not implemented yet)
* cluster delete
* state delete (fails due to unimplemented methods)

