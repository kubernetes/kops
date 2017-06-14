Documentation Index
===================

This page serves as a categorized index for all `kops` documentation.


Quick start
-----------
* [Getting started on AWS](aws.md)
* [CLI reference](cli/kops.md)


Overview
--------

* [CLI](#cli)
* [Introspection](#introspection)
* [`kops` design documents](#kops-design-documents)
* [Networking](#networking)
* [Operations](#operations)
* [Security](#security)
* [Workflows](#workflows)



### CLI ###

* [CLI argument explanations](arguments.md)
* [CLI reference](cli/kops.md)
* [Commands](commands.md)
    * miscellaneous CLI-related remarks
* [Experimental features](experimental.md)
    * list of and how to enable experimental flags in the CLI
* [kubectl](kubectl.md)
    * how to point kubectl to your `kops` cluster


### Introspection ###

* [Download `kops` configuration](download_config.md)
    * methods to download the current generated `kops` configuration
* [Get AWS subdomain NS records](ns.md)


### `kops` design documents ###

* [`kops` cluster boot sequence](boot-sequence.md)
* [`kops` cluster spec](cluster_spec.md)
* [`kops` instance groups](instance_groups.md)
* [`kops` philosophy](philosophy.md)
* [`kops` state store](state.md)


### Networking ###

* [Networking overview](networking.md)
* [Run `kops` in an existing VPC](run_in_existing_vpc.md)
* [Supported network topologies](topology.md)
* [Subdomain setup](creating_subdomain.md)


### Operations ###

* [Cluster addon manager](addon_manager.md)
* [Cluster addons](addons.md)
* [Cluster configuration management](changing_configuration.md)
* [Cluster with HA creation example](advanced_create.md)
* [Cluster upgrades and migrations](cluster_upgrades_and_migrations.md)
* [`etcd` volume encryption setup](etcd_volume_encryption.md)
* [`etcd` backup setup](etcd_backup.md)
* [GPU setup](gpu.md)
* [High Availability](high_availability.md)
* [InstanceGroup images](images.md)
    * how to modify IG images and information on available/tested images
* [Label management](labels.md)
    * for AWS instance tags and `k8s` node labels
* [`kube-up` to `kops` migration](upgrade_from_kubeup.md)
* [Secret management](secrets.md)
* [Single-master to multi-master migration](single-to-multi-master.md)
* [`k8s` updating](upgrade.md)
* [`kops` updating](update_kops.md)


### Security ###

* [Bastion setup](bastion.md)
* [IAM roles](iam_roles.md)
* [MFA setup](mfa.md)
    * how to set up MFA for `kops`
* [Security](security.md)
    * overview of secret storage, SSH credentials etc.


### Workflows ###

* [E2E testing with `kops` clusters](testing.md)
* [Getting started on AWS](aws.md)
