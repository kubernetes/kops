<div class="hidden">
<hr>
<strong>For a better viewing experience please check out our live documentation site at <a href="https://kops.sigs.k8s.io/">kops.sigs.k8s.io</a>.</strong>
<hr>
</div>

# kops - Kubernetes Operations

[GoDoc]: https://pkg.go.dev/k8s.io/kops
[GoDoc Widget]: https://godoc.org/k8s.io/kops?status.svg

The easiest way to get a production grade Kubernetes cluster up and running.

## 2020-05-06 etcd-manager Certificate Expiration Advisory

kops versions released today contain a **critical fix** to etcd-manager: 1 year after creation (or first adopting etcd-manager), clusters will stop responding due to expiration of a TLS certificate. Upgrading kops to 1.15.3, 1.16.2, 1.17.0-beta.2, or 1.18.0-alpha.3 is highly recommended. Please see the [advisory](./advisories/etcd-manager-certificate-expiration.md) for the full details.

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