# Documentation Index

## Quick start
* [Getting started on AWS](aws.md)
* [CLI reference](cli/kops.md)


## Overview

* [Getting started on AWS](aws.md)
* [Command-line interface](#commandline-interface)
* [Inspection](#inspection)
* [`kops` design documents](#kops-design-documents)
* [Networking](#networking)
* [Operations](#operations)
* [Security](#security)
* [Workflows](#workflows)


## Command-line interface

* [CLI argument explanations](arguments.md)
* [CLI reference](cli/kops.md)
* [Commands](commands.md)
    * miscellaneous CLI-related remarks
* [Experimental features](experimental.md)
    * list of and how to enable experimental flags in the CLI
* [kubectl](kubectl.md)
    * how to point kubectl to your `kops` cluster


## Operations

* [Cluster addon manager](addon_manager.md)
* [Cluster addons](addons.md)
* [Cluster configuration management](changing_configuration.md)
* [`kops` cluster API definitions](cluster_spec.md)
    * overview of some of the API value to customize a `kops` cluster
* [Cluster upgrades and migrations](cluster_upgrades_and_migrations.md)
* [`etcd` volume encryption setup](etcd_volume_encryption.md)
* [`etcd` backup setup](etcd_backup.md)
* [GPU setup](gpu.md)
* [High Availability](high_availability.md)
* [`kops` instance groups API](instance_groups.md)
    * overview of some of the API value to customize a `kops` groups of ks8 nodes
* [InstanceGroup images](images.md)
    * how to use other image for cluster nodes, and information on available/tested images
* [Using Manifests and Customization via the API](manifests_and_customizing_via_api.md)
    * how to use YAML to manage clusters.
    * how to customize cluster via the `kops` API.
* [`k8s` upgrading](upgrade.md)
* [`kops` updating](update_kops.md)
* [`kube-up` to `kops` upgrade](upgrade_from_kubeup.md)
* [Label management](labels.md)
    * for cluster nodes
* [Secret management](secrets.md)
* [Moving from a Single Master to Multiple HA Masters](single-to-multi-master.md)


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

* [E2E testing with `kops` clusters](development/testing.md)
* [Developing using Docker](development/Docker.md)
* [Updating Go Dependencies](development/dependencies.md)
* [Hack Directory](development/hack.md)
* [Rolling Update Diagrams](development/rolling_update.md)
* [Examlpe on how to add a feature](development/adding_a_feature.md)
* [Documentation Guidelines](development/documentation.md)
* [Releasing with Homebrew](development/homebrew.md)
* [Notes on master instance sizing](development/instancesizes.md)
* [Development with vSphere](development/vsphere-dev.md)
* [How to update `kops` API](development/api_updates.md)
* [Notes on Gossip design](development/gossip.md)
* [Low level description on how kops works](how_it_works.md)
* [Our release process](development/release.md)
