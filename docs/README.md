Documentation Index
===================

Quick start
-----------
* [Getting started on AWS](aws.md)
* [CLI reference](cli/kops.md)


Overview
--------

* [Command-line interface](#commandline-interface)
* [Inspection](#inspection)
* [`kops` design documents](#kops-design-documents)
* [Networking](#networking)
* [Operations](#operations)
* [Security](#security)
* [Workflows](#workflows)



### Command-line interface ###

* [CLI argument explanations](arguments.md)
* [CLI reference](cli/kops.md)
* [Commands](commands.md)
    * miscellaneous CLI-related remarks
* [Experimental features](experimental.md)
    * list of and how to enable experimental flags in the CLI
* [kubectl](kubectl.md)
    * how to point kubectl to your `kops` cluster


### Inspection ###

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
* [Single-master to multi-master update](single-to-multi-master.md)
* [Developers guide for vSphere support](vsphere-dev.md)
* [vSphere support status](vsphere-development-status.md)

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
