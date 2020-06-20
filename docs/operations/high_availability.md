# High Availability (HA)

## Introduction

For testing purposes, kubernetes works just fine with a single master. However, when the master becomes unavailable, for example due to upgrade or instance failure, the kubernetes API will be unavailable. Pods and services that are running in the cluster continue to operate as long as they do not depend on interacting with the API, but operations such as adding nodes, scaling pods, replacing terminated pods will not work. Running kubectl will also not work. 

kops runs each master in a dedicated autoscaling groups (ASG) and stores data on EBS volumes. That way, if a master node is terminated the ASG will launch a new master instance with the master's volume. Because of the dedicated EBS volumes, each master is bound to a fixed Availability Zone (AZ). If the AZ becomes unavailable, the master instance in that AZ will also become unavailable.

For production use, you therefore want to run kubernetes in a HA setup with multiple masters. With multiple master nodes, you will be able both to do graceful (zero-downtime) upgrades and you will be able to survive AZ failures.

Very few regions offer less than 3 AZs. In this case, running multiple masters in the same AZ is an option. If the AZ with multiple masters becomes unavailable you will still have downtime with this configuration. But regular changes to master nodes such as upgrades will be graceful and without downtime.

If you already have a single-master cluster you would like to convert to a multi-master cluster, read the [single to multi-master](../single-to-multi-master.md) docs.

Note that running clusters spanning several AZs is more expensive than running clusters spanning one or two AZs. This happens not only because of the master EC2 cost, but also because you have to pay for cross-AZ traffic. Depending on your workload you may therefore also want to consider running worker nodes only in two AZs. As long as your application do not rely on quorum, you will still have AZ fault tolerance.

## Creating a HA cluster

### Example 1: public topology

The simplest way to get started with a HA cluster is to run `kops create cluster` as shown below. The `--master-zones` flag lists the zones you want your masters
to run in. By default, kops will create one master per AZ. Since the kubernetes etcd cluster runs on the master nodes, you have to specify an odd number of zones in order to obtain quorum.

```
kops create cluster \
    --node-count 3 \
    --zones us-west-2a,us-west-2b,us-west-2c \
    --master-zones us-west-2a,us-west-2b,us-west-2c \
    hacluster.example.com
```

## Example 2: private topology

Create a cluster using [private network topology](../topology.md):

```
kops create cluster \
    --node-count 3 \
    --zones us-west-2a,us-west-2b,us-west-2c \
    --master-zones us-west-2a,us-west-2b,us-west-2c \
    --topology private \
    --networking <provider> \
    ${NAME}
```

Note that the default networking provider (kubenet) does not support private topology.

## Example 3: multiple masters in the same AZ

If necessary, for example in regions with less than 3 AZs, you can launch multiple masters in the same AZ.

```
kops create cluster \
    --node-count 3 \
    --master-count 3 \
    --zones cn-north-1a,cn-north-1b \
    --master-zones cn-north-1a,cn-north-1b \
    hacluster.k8s.local
```
