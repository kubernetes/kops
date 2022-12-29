# Recommendations for production setups

The getting started-documentation is a fast way of spinning up a Kubernetes cluster, but there are some aspects of kOps that require extra consideration. This document will highlight the most important things you should know about before deploying your production workload.

## High availability

Running only a single control-plane node can be error-prone and disruptive. 

Read through the [high availability documentation](../operations/high_availability.md) to learn how to set up a cluster with redundant control plane.

## Networking

The default networking of kOps, Cilium, is suitable for production.

Read through the [networking page](../networking.md) to see what the other CNI choices are.

## Private topology

By default, kOps will create IPv4 clusters using public topology, where all nodes and the Kubernetes API are exposed on public Internet.

Read through the [topology page](../topology.md) to understand the options you have running nodes in internal IP addresses and using a [bastion](../bastion.md) for SSH access.

## Cluster spec

The `kops` command allows you to configure some aspects of your cluster, but for almost any production cluster, you will want to change settings that are not accessible through the CLI. The cluster spec can be exported as a yaml file and checked into version control.

Read through the [cluster spec page](../cluster_spec.md) and familiarize yourself with the key options that kOps offers.

## Templating

If your cluster contains multiple Instance Groups, or if you manage multiple clusters, you want to use generate the cluster spec using templates.

Read through the [templating documentation](../operations/cluster_template.md) to learn how to make use of templates.
