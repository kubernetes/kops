# Documentation Index

## Quick start
* [Getting started on AWS](aws.md)
* [CLI reference](cli/kops.md)


## Overview

* [Getting started on AWS](aws.md)
* [Command-line interface](#command-line-interface)
* [Inspection](#inspection)
* [`kops` design documents](#kops-design-documents)
* [Networking](#networking)
* [Operations](#operations)
* [Security](#security)
* [Development](#development)


## Command-line interface

* [CLI argument explanations](arguments.md)
* [CLI reference](cli/kops.md)
* [Commands](commands.md)
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
* [Cluster addon manager](addon_manager.md)
* [Cluster addons](addons.md)
* [Cluster configuration management](changing_configuration.md)
* [Cluster desired configuration creation from template](cluster_template.md)
* [Cluster upgrades and migrations](cluster_upgrades_and_migrations.md)
* [`etcd` volume encryption setup](etcd_volume_encryption.md)
* [`etcd` backup setup](etcd_backup.md)
* [GPU setup](gpu.md)
* [High Availability](high_availability.md)
* [InstanceGroup images](images.md)
    * how to use other image for cluster nodes, and information on available/tested images
* [`k8s` upgrading](upgrade.md)
* [`kops` updating](update_kops.md)
* [`kube-up` to `kops` upgrade](upgrade_from_kubeup.md)
* [Label management](labels.md)
    * for cluster nodes
* [Secret management](secrets.md)
* [Moving from a Single Master to Multiple HA Masters](single-to-multi-master.md)
* [Developers guide for vSphere support](vsphere-dev.md)
* [vSphere support status](vsphere-development-status.md)

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
* [Get AWS subdomain NS records](ns.md)


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
* [Updating Go Dependencies](development/dependencies.md)
