# Getting Started with kops on Spot Ocean

[Ocean](https://spot.io/products/ocean/) by [Spot](https://spot.io/) simplifies infrastructure management for Kubernetes.  With robust, container-driven infrastructure auto-scaling and intelligent right-sizing for container resource requirements, operations can literally "set and forget" the underlying cluster.

Ocean seamlessly integrates with your existing instance groups, as a drop-in replacement for AWS Auto Scaling groups, and allows you to streamline and optimize the entire workflow, from initially creating your cluster to managing and optimizing it on an ongoing basis.

## Features

- **Simplify Cluster Management** —
Ocean's Virtual Node Groups make it easy to run different infrastructure in a single cluster, which can span multiple AWS VPC availability zones and subnets for high-availability.

- **Container-Driven Autoscaling and Vertical Rightsizing** —
Auto-detect your container infrastructure requirements so the appropriate instance size or type will always be available. Measure real-time CPU/Memory consumption of your Pods for ongoing resource optimization.

- **Cloud-Native Showback** —
Gain a granular view of your cluster's cost breakdown (compute and storage) for each and every one of the cluster's resources such as Namespaces, Deployments, Daemon Sets, Jobs, and Pods.

- **Optimized Pricing and Utilization** —
Ocean not only intelligently leverages Spot Instances and reserved capacity to reduce costs, but also eliminates underutilized instances with container-driven autoscaling and advanced bin-packing.

## Prerequisites

Make sure you have [installed kops](../install.md) and [installed kubectl](../install.md#installing-other-dependencies).

## Setup your environment

### Spot
Generate your credentials [here](https://console.spotinst.com/spt/settings/tokens/permanent). If you are not a Spot Ocean user, sign up for free [here](https://console.spotinst.com/spt/auth/signUp). For further information, please checkout our [Spot API](https://help.spot.io/spotinst-api/) guide, available on the [Spot Help Center](https://help.spot.io/) website.

To use environment variables, run:
```bash
export SPOTINST_TOKEN=<spotinst_token>
export SPOTINST_ACCOUNT=<spotinst_account>
```

To use credentials file, run the [spotctl configure](https://github.com/spotinst/spotctl#getting-started) command:
```bash
spotctl configure
? Enter your access token [? for help] **********************************
? Select your default account  [Use arrows to move, ? for more help]
> act-01234567 (prod)
  act-0abcdefg (dev)
```

Or, manually create an INI formatted file like this:
```ini
[default]
token   = <spotinst_token>
account = <spotinst_account>
```

and place it in:

- Unix/Linux/macOS:
```bash
~/.spotinst/credentials
```
- Windows:
```bash
%UserProfile%\.spotinst\credentials
```

### AWS

Make sure to set up [a dedicated IAM user](./aws.md#setup-iam-user), [DNS records](./aws.md#configure-dns) and [cluster state storage](./aws.md#cluster-state-storage). Please refer to [setup your environment](./aws.md#setup-your-environment) for further details.

## Feature Flags

| Flag | Description |
|---|---|
| `+Spotinst` | Enables the use of the Spot integration. |
| `+SpotinstOcean` | Enables the use of the Spot Ocean integration. |
| `+SpotinstHybrid` | Toggles between hybrid and full instance group implementations. Allows you to gradually integrate with Spot Ocean by continuing to use instance groups through AWS Auto Scaling groups, except for specific instance groups labeled with a predefined [metadata label](#metadata-labels). |
| `-SpotinstController` | Toggles the installation of the Spot controller addon off. Please note that the feature flag must be prefixed with a minus (`-`) sign to set its value to `false`, which results in disabling the controller. |

## Creating a Cluster

You can add an Ocean instance group to new or existing clusters. To create a new cluster with a Ocean instance groups, run:

```bash
# configure the feature flags
export KOPS_FEATURE_FLAGS="Spotinst,SpotinstOcean"

# create the cluster
kops create cluster --zones=us-west-2a example
```

!!!note
    It's possible to have a cluster with both Ocean-managed and unmanaged instance groups.

```bash
# configure the feature flags
export KOPS_FEATURE_FLAGS="Spotinst,SpotinstOcean,SpotinstHybrid"

# create the instance groups
kops create -f instancegroups.yaml
```

```yaml
# instancegroups.yaml
# A cluster with both Ocean-managed and unmanaged instance groups.
---
# Use Ocean in hybrid mode.
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: "example"
    spotinst.io/hybrid: "true"
  ...
---
# Use AWS Auto Scaling group.
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: "example"
  ...
```

## Creating an Instance Group

To create a new instance group, run:

```bash
# configure the feature flags
export KOPS_FEATURE_FLAGS="Spotinst,SpotinstOcean"

# create the instance group
kops create instancegroup --role=node --name=example
```

To create a new instance group and have more control over the configuration, a config file can be used.

```yaml
# instancegroup.yaml
# An instance group with Ocean configuration.
---
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: "example"
    spotinst.io/spot-percentage: "90"
  ...
```

## Metadata Labels

| Label | Description | Default |
|---|---|---|
| `spotinst.io/spot-percentage` | Specify the percentage of Spot instances that should spin up from the target capacity. | `100` |
| `spotinst.io/utilize-reserved-instances` | Specify whether reserved instances should be utilized. | `true` |
| `spotinst.io/fallback-to-ondemand` | Specify whether fallback to on-demand instances should be enabled. | `true` |
| `spotinst.io/grace-period` | Specify a period of time, in seconds, that Ocean should wait before applying instance health checks. | none |
| `spotinst.io/ocean-default-launchspec` | Specify whether to use the InstanceGroup's spec as the default Launch Spec for the Ocean cluster. | none |
| `spotinst.io/ocean-instance-types-whitelist` | Specify whether to whitelist specific instance types. | none |
| `spotinst.io/ocean-instance-types-blacklist` | Specify whether to blacklist specific instance types. | none |
| `spotinst.io/autoscaler-disabled` | Specify whether the auto scaler should be disabled. | `false` |
| `spotinst.io/autoscaler-default-node-labels` | Specify whether default node labels should be set for the auto scaler. | `false` |
| `spotinst.io/autoscaler-headroom-cpu-per-unit` | Specify the number of CPUs to allocate for headroom. CPUs are denoted in millicores, where 1000 millicores = 1 vCPU. | none |
| `spotinst.io/autoscaler-headroom-gpu-per-unit` | Specify the number of GPUs to allocate for headroom. | none |
| `spotinst.io/autoscaler-headroom-mem-per-unit` | Specify the amount of memory (MB) to allocate for headroom. | none |
| `spotinst.io/autoscaler-headroom-num-of-units` | Specify the number of units to retain as headroom, where each unit has the defined CPU and memory. | none |
| `spotinst.io/autoscaler-cooldown` | Specify a period of time, in seconds, that Ocean should wait between scaling actions. | `300` |
| `spotinst.io/autoscaler-scale-down-max-percentage` | Specify the maximum scale down percentage. | none |
| `spotinst.io/autoscaler-scale-down-evaluation-periods` | Specify the number of evaluation periods that should accumulate before a scale down action takes place. | `5` |

## Documentation

If you're new to [Spot](https://spot.io/) and want to get started, please checkout our [Getting Started](https://help.spot.io/getting-started-with-spotinst/) guide, available on the [Spot Help Center](https://help.spot.io/) website.

## Getting Help

Please use these community resources for getting help:

- Join our [Spot](https://spot.io/) community on [Slack](http://slack.spot.io/).
- Open a GitHub [issue](https://github.com/kubernetes/kops/issues/new/choose/).
- Ask a question on [Stack Overflow](https://stackoverflow.com/) and tag it with `spot-ocean`.
