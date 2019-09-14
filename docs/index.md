# kops - Kubernetes Operations

[GoDoc]: https://godoc.org/k8s.io/kops
[GoDoc Widget]: https://godoc.org/k8s.io/kops?status.svg

The easiest way to get a production grade Kubernetes cluster up and running.

## What is kops?

We like to think of it as `kubectl` for clusters.

`kops` helps you create, destroy, upgrade and maintain production-grade, highly
available, Kubernetes clusters from the command line. AWS (Amazon Web Services)
is currently officially supported, with GCE and OpenStack in beta support, and VMware vSphere
in alpha, and other platforms planned.

## Can I see it in action?

<p align="center">
  <a href="https://asciinema.org/a/97298">
  <img src="https://asciinema.org/a/97298.png" width="885"></image>
  </a>
</p>


## Features

* Automates the provisioning of Kubernetes clusters in [AWS](getting_started/aws.md) and [GCE](getting_started/gce.md)
* Deploys Highly Available (HA) Kubernetes Masters
* Built on a state-sync model for **dry-runs** and automatic **idempotency**
* Ability to generate [Terraform](terraform.md)
* Supports custom Kubernetes [add-ons](operations/addons.md)
* Command line [autocompletion](cli/kops_completion.md)
* YAML Manifest Based API [Configuration](manifests_and_customizing_via_api.md)
* [Templating](operations/cluster_template.md) and dry-run modes for creating
 Manifests
* Choose from eight different CNI [Networking](networking.md) providers out-of-the-box
* Supports upgrading from [kube-up](upgrade_from_kubeup.md)
* Capability to add containers, as hooks, and files to nodes via a [cluster manifest](cluster_spec.md)


## Documentation

[To check out Live documentation](https://kubernetes.github.io/kops/)


## Kubernetes Release Compatibility


### Kubernetes Version Support

kops is intended to be backward compatible.  It is always recommended to use the
latest version of kops with whatever version of Kubernetes you are using.  We suggest
kops users run one of the [3 minor versions](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/release/versioning.md#supported-releases-and-component-skew) Kubernetes is supporting however we
do our best to support previous releases for a period of time.

One exception, in regard to compatibility, kops supports the equivalent
Kubernetes minor release number.  A minor version is the second digit in the
release number.  kops version 1.13.0 has a minor version of 13. The numbering
follows the semantic versioning specification, MAJOR.MINOR.PATCH.

For example, kops 1.12.0 does not support Kubernetes 1.13.0, but kops 1.13.0
supports Kubernetes 1.12.2 and previous Kubernetes versions. Only when the kops minor
version matches the Kubernetes minor version does kops officially support the
Kubernetes release.  kops does not stop a user from installing mismatching
versions of K8s, but Kubernetes releases always require kops to install specific
versions of components like docker, that tested against the particular
Kubernetes version.

| kops version  | k8s 1.9.x | k8s 1.10.x | k8s 1.11.x | k8s 1.12.x | k8s 1.13.x | k8s 1.14.x |
|---------------|-----------|------------|------------|------------|------------|------------|
| 1.14.x - Beta | ✔         | ✔          | ✔          | ✔          | ✔          | ✔          |
| 1.13.x        | ✔         | ✔          | ✔          | ✔          | ✔          | ❌         |
| 1.12.x        | ✔         | ✔          | ✔          | ✔          | ❌         | ❌         |
| 1.11.x        | ✔         | ✔          | ✔          | ❌         | ❌         | ❌         |
| ~~1.10.x~~    | ✔         | ✔          | ❌         | ❌         | ❌         | ❌         |
| ~~1.9.x~~     | ✔         | ❌         | ❌         | ❌         | ❌         | ❌         |

Use the latest version of kops for all releases of Kubernetes, with the caveat
that higher versions of Kubernetes are not _officially_ supported by kops. Releases who are ~~crossed out~~ _should_ work but we suggest should be upgraded soon.

### kops Release Schedule

This project does not follow the Kubernetes release schedule.  `kops` aims to
provide a reliable installation experience for kubernetes, and typically
releases about a month after the corresponding Kubernetes release. This time
allows for the Kubernetes project to resolve any issues introduced by the new
version and ensures that we can support the latest features. kops will release
alpha and beta pre-releases for people that are eager to try the latest
Kubernetes release.  Please only use pre-GA kops releases in environments that
can tolerate the quirks of new releases, and please do report any issues
encountered.
