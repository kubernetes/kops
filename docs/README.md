# Documentation Index

## Quick start
* [Getting started on AWS](getting_started/aws.md)
* [Getting started on GCE](getting_started/gce.md)
* [CLI reference](cli/kops.md)


## Overview

- [Documentation Index](#documentation-index)
  - [Quick start](#quick-start)
  - [Overview](#overview)
  - [Command-line interface](#command-line-interface)
  - [Advanced / Detailed List of Configurations](#advanced--detailed-list-of-configurations)
    - [API / Configuration References](#api--configuration-references)
    - [API Usage Guides](#api-usage-guides)
  - [Operations](#operations)
  - [Networking](#networking)
  - [`kops` design documents](#kops-design-documents)
  - [Security](#security)
  - [Inspection](#inspection)
  - [Development](#development)


## Command-line interface

* [CLI argument explanations](arguments.md)
* [CLI reference](cli/kops.md)
* [Commands](usage/commands.md)
    * miscellaneous CLI-related remarks
* [Experimental features](experimental.md)
    * list of and how to enable experimental flags in the CLI
* [kubectl](kubectl.md)
    * how to point kubectl to your `kops` cluster

## Advanced / Detailed List of Configurations

### API / Configuration References
* [Godocs for Cluster - `ClusterSpec`](https://godoc.org/k8s.io/kops/pkg/apis/kops#ClusterSpec).
* [Godocs for Instance Group - `InstanceGroupSpec`](https://godoc.org/k8s.io/kops/pkg/apis/kops#InstanceGroupSpec).

### API Usage Guides
* [`kops` cluster API definitions](cluster_spec.md)
    * overview of some of the API value to customize a `kops` cluster
* [`kops` instance groups API](instance_groups.md)
    * overview of some of the API value to customize a `kops` groups of k8s nodes
* [Using Manifests and Customizing via the API](manifests_and_customizing_via_api.md)

## Operations
* [Cluster addon manager](operations/addons.md#addon_management)
* [Cluster addons](operations/addons.md)
* [Cluster configuration management](changing_configuration.md)
* [Cluster desired configuration creation from template](operations/cluster_template.md)
* [Cluster upgrades and migrations](operations/cluster_upgrades_and_migrations.md)
* [`etcd` volume encryption setup](operations/etcd_backup_restore_encryption.md#etcd-volume-encryption)
* [`etcd` backup/restore](operations/etcd_backup_restore_encryption.md#backing-up-etcd)
* [GPU setup](gpu.md)
* [High Availability](operations/high_availability.md)
* [InstanceGroup images](operations/images.md)
    * how to use other image for cluster nodes, and information on available/tested images
* [`k8s` upgrading](operations/updates_and_upgrades.md#upgrading-kubernetes)
* [`kops` updating](operations/updates_and_upgrades.md#updating-kops)
* [`kube-up` to `kops` upgrade](upgrade_from_kubeup.md)
* [Label management](labels.md)
    * for cluster nodes
* [Secret management](secrets.md)
* [Moving from a Single Master to Multiple HA Masters](single-to-multi-master.md)
* [Upgrading Kubernetes](tutorial/upgrading-kubernetes.md)
* [Working with Instance Groups](tutorial/working-with-instancegroups.md)
* [Developers guide for vSphere support](vsphere-dev.md)
* [vSphere support status](vsphere-development-status.md)
* [Running `kops` in a CI environment](continuous_integration.md)

## Networking

* [Networking Overview including CNI](networking.md)
* [Run `kops` in an existing VPC](run_in_existing_vpc.md)
* [Supported network topologies](topology.md)
* [Subdomain setup](creating_subdomain.md)


## `kops` design documents

* [`kops` cluster boot sequence](boot-sequence.md)
* [`kops` philosophy](philosophy.md)
* [`kops` state store](state.md)


## Security

* [Bastion setup](bastion.md)
* [IAM roles](iam_roles.md)
* [MFA setup](mfa.md)
    * how to set up MFA for `kops`
* [Security](security.md)
    * overview of secret storage, SSH credentials etc.


## Inspection

* [Download `kops` configuration](download_config.md)
    * methods to download the current generated `kops` configuration
* [Get AWS subdomain NS records](advanced/ns.md)


## Development

* [Developing using Docker](development/Docker.md)
* [Development with vSphere](vsphere-dev.md)
* [Documentation Guidelines](development/documentation.md)
* [E2E testing with `kops` clusters](development/testing.md)
* [Example on how to add a feature](development/adding_a_feature.md)
* [Hack Directory](development/hack.md)
* [How to update `kops` API](development/api_updates.md)
* [Low level description on how kops works](development/how_it_works.md)
* [Notes on Gossip design](development/gossip.md)
* [Notes on master instance sizing](development/instancesizes.md)
* [Our release process](development/release.md)
* [Releasing with Homebrew](development/homebrew.md)
* [Rolling Update Diagrams](development/rolling_update.md)
