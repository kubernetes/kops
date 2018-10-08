# kops - Kubernetes Operations

[GoDoc]: https://godoc.org/k8s.io/kops
[GoDoc Widget]: https://godoc.org/k8s.io/kops?status.svg

The easiest way to get a production grade Kubernetes cluster up and running.

## What is kops?

We like to think of it as `kubectl` for clusters.

`kops` helps you create, destroy, upgrade and maintain production-grade, highly
available, Kubernetes clusters from the command line. AWS (Amazon Web Services)
is currently officially supported, with GCE in beta support , and VMware vSphere
in alpha, and other platforms planned.


## Can I see it in action?

<p align="center">
  <a href="https://asciinema.org/a/97298">
  <img src="https://asciinema.org/a/97298.png" width="885"></image>
  </a>
</p>


## Features

* Automates the provisioning of Kubernetes clusters in [AWS](aws.md) and [GCE](tutorial/gce.md)
* Deploys Highly Available (HA) Kubernetes Masters
* Built on a state-sync model for **dry-runs** and automatic **idempotency**
* Ability to generate [Terraform](terraform.md)
* Supports custom Kubernetes [add-ons](addons.md)
* Command line [autocompletion](cli/kops_completion.md)
* YAML Manifest Based API [Configuration](manifests_and_customizing_via_api.md)
* [Templating](cluster_template.md) and dry-run modes for creating
 Manifests
* Choose from eight different CNI [Networking](networking.md) providers out-of-the-box
* Supports upgrading from [kube-up](upgrade_from_kubeup.md)
* Capability to add containers, as hooks, and files to nodes via a [cluster manifest](cluster_spec.md)


## Documentation

[To check out Live documentation](https://kubernetes.github.io/kops/)


## Kubernetes Release Compatibility


### Kubernetes Version Support

kops is intended to be backward compatible.  It is always recommended to use the
latest version of kops with whatever version of Kubernetes you are using.  Always
use the latest version of kops.

One exception, in regards to compatibility, kops supports the equivalent
Kubernetes minor release number.  A minor version is the second digit in the
release number.  kops version 1.8.0 has a minor version of 8. The numbering
follows the semantic versioning specification, MAJOR.MINOR.PATCH.

For example, kops 1.8.0 does not support Kubernetes 1.9.2, but kops 1.9.0
supports Kubernetes 1.9.2 and previous Kubernetes versions. Only when kops minor
version matches, the Kubernetes minor version does kops officially support the
Kubernetes release.  kops does not stop a user from installing mismatching
versions of K8s, but Kubernetes releases always require kops to install specific
versions of components like docker, that tested against the particular
Kubernetes version.

#### Compatibility Matrix

| kops version | k8s 1.5.x | k8s 1.6.x | k8s 1.7.x | k8s 1.8.x | k8s 1.9.x |
|--------------|-----------|-----------|-----------|-----------|-----------|
| 1.9.x        | Y         | Y         | Y         | Y         | Y         |
| 1.8.x        | Y         | Y         | Y         | Y         | N         |
| 1.7.x        | Y         | Y         | Y         | N         | N         |
| 1.6.x        | Y         | Y         | N         | N         | N         |

Use the latest version of kops for all releases of Kubernetes, with the caveat
that higher versions of Kubernetes are not _officially_ supported by kops.
