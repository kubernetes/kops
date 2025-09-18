# High Availability (HA)

## Introduction

For testing purposes, Kubernetes works just fine with a single control-plane node. However, when the control-plane node becomes unavailable, for example due to upgrade or instance failure, the Kubernetes API will be unavailable. Pods and services that are running in the cluster continue to operate as long as they do not depend on interacting with the API, but operations such as adding nodes, scaling pods, replacing terminated pods will not work. Running `kubectl` will also not work. 

kOps runs each control-plane node in a dedicated autoscaling groups (ASG) and stores data on EBS volumes. That way, if a control-plane node is terminated the ASG will launch a new control-plane instance with the control-plane node's volume. Because of the dedicated EBS volumes, each control-plane node is bound to a fixed Availability Zone (AZ). If the AZ becomes unavailable, the control-plane instance in that AZ will also become unavailable.

For production use, you therefore want to run Kubernetes in a HA setup with multiple control-plane nodes. With multiple control-plane nodes, you will be able both to do graceful (zero-downtime) upgrades and you will be able to survive AZ failures.

Very few regions offer less than 3 AZs. In this case, running multiple control-plane nodes in the same AZ is an option. If the AZ with multiple control-plane nodes becomes unavailable you will still have downtime with this configuration. But regular changes to control-plane nodes such as upgrades will be graceful and without downtime.

If you already have a single control-plane node cluster you would like to convert to a multi control-plane node cluster, read the [single to multi-master](../single-to-multi-master.md) docs.

Note that running clusters spanning several AZs is more expensive than running clusters spanning one or two AZs. This happens not only because of the master EC2 cost, but also because you have to pay for cross-AZ traffic. Depending on your workload you may therefore also want to consider running worker nodes only in two AZs. As long as your application do not rely on quorum, you will still have AZ fault tolerance.

## Creating a HA cluster

### Example 1: public topology

The simplest way to get started with a HA cluster is to run `kops create cluster` as shown below. The `--control-plane-zones` flag lists the zones you want your control-plane nodes
to run in. By default, kOps will create one control-plane node per AZ. Since the Kubernetes etcd cluster runs on the control-plane nodes, you have to specify an odd number of zones in order to obtain quorum.

```
kops create cluster \
    --node-count 3 \
    --zones us-west-2a,us-west-2b,us-west-2c \
    --control-plane-zones us-west-2a,us-west-2b,us-west-2c \
    hacluster.example.com
```

## Example 2: private topology

Create a cluster using [private network topology](../topology.md):

```
kops create cluster \
    --node-count 3 \
    --zones us-west-2a,us-west-2b,us-west-2c \
    --control-plane-zones us-west-2a,us-west-2b,us-west-2c \
    --topology private \
    --networking <provider> \
    ${NAME}
```

Note that the default networking provider (kubenet) does not support private topology.

## Example 3: multiple control-plane nodes in the same AZ

If necessary, for example in regions with less than 3 AZs, you can launch multiple control-plan nodes in the same AZ.

```
kops create cluster \
    --node-count 3 \
    --control-plane-count 3 \
    --zones cn-north-1a,cn-north-1b \
    --control-plane-zones cn-north-1a,cn-north-1b \
    hacluster.k8s.local
```
