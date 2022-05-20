<div class="hidden">
<hr>
<strong>For a better viewing experience please check out our live documentation site at <a href="https://kops.sigs.k8s.io/">kops.sigs.k8s.io</a>.</strong>
<hr>
</div>

# kOps - Kubernetes Operations

[GoDoc]: https://pkg.go.dev/k8s.io/kOps
[GoDoc Widget]: https://godoc.org/k8s.io/kOps?status.svg

The easiest way to get a production grade Kubernetes cluster up and running.

## What is kOps?

We like to think of it as `kubectl` for clusters.

`kops` will not only help you create, destroy, upgrade and maintain production-grade, highly
available, Kubernetes cluster, but it will also provision the necessary cloud infrastructure.

[AWS](getting_started/aws.md) (Amazon Web Services) is currently officially supported, with [DigitalOcean](getting_started/digitalocean.md), [GCE](getting_started/gce.md) and [OpenStack](getting_started/openstack.md) in beta support, and [Azure](getting_started/azure.md), and AliCloud in alpha.

## Can I see it in action?

<p align="center">
  <a href="https://asciinema.org/a/97298">
  <img src="https://asciinema.org/a/97298.png" width="885"></image>
  </a>
</p>


## Features

* Automates the provisioning of Highly Available Kubernetes clusters
* Built on a state-sync model for **dry-runs** and automatic **idempotency**
* Ability to generate [Terraform](terraform.md)
* Supports **zero-config** managed kubernetes [add-ons](addons.md)
* Command line [autocompletion](cli/kops_completion.md)
* YAML Manifest Based API [Configuration](manifests_and_customizing_via_api.md)
* [Templating](operations/cluster_template.md) and dry-run modes for creating Manifests
* Choose from most popular CNI [Networking](networking.md) providers out-of-the-box
* Multi-architecture ready with ARM64 support
* Capability to add containers, as hooks, and files to nodes via a [cluster manifest](cluster_spec.md)