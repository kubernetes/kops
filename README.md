<img src="/docs/img/logo.jpg" width="500px" alt="kops logo">

# kops - Kubernetes Operations

[![Build Status](https://travis-ci.org/kubernetes/kops.svg?branch=master)](https://travis-ci.org/kubernetes/kops) [![Go Report Card](https://goreportcard.com/badge/k8s.io/kops)](https://goreportcard.com/report/k8s.io/kops)  [![GoDoc Widget]][GoDoc]

[GoDoc]: https://pkg.go.dev/k8s.io/kops
[GoDoc Widget]: https://godoc.org/k8s.io/kops?status.svg


The easiest way to get a production grade Kubernetes cluster up and running.

## 2020-05-06 etcd-manager Certificate Expiration Advisory

kops versions released today contain a **critical fix** to etcd-manager: 1 year after creation (or first adopting etcd-manager), clusters will stop responding due to expiration of a TLS certificate. Upgrading kops to 1.15.3, 1.16.2, 1.17.0-beta.2, or 1.18.0-alpha.3 is highly recommended. Please see the [advisory](./docs/advisories/etcd-manager-certificate-expiration.md) for the full details.

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


## Installing and launching a Kubernetes cluster hosted on AWS, GCE, DigitalOcean or OpenStack

See [Getting Started](https://kops.sigs.k8s.io/getting_started/install/)


## Documentation

Documentation is in the `/docs` directory, and can be seen at [kops.sigs.k8s.io](https://kops.sigs.k8s.io/).


## Releases and kubernetes Release Compatibility

See [Releases and versioning](https://kops.sigs.k8s.io/welcome/releases/)


## Getting Involved and Contributing

See [Contributing](https://kops.sigs.k8s.io/welcome/contributing/)

### Office Hours

Kops maintainers set aside one hour every other week for **public** office hours. This time is used to gather with community members interested in kops. This session is open to both developers and users.

We do maintain an [agenda](https://docs.google.com/document/d/12QkyL0FkNbWPcLFxxRGSPt_tNPBHbmni3YLY-lHny7E/edit) and stick to it as much as possible. If you want to hold the floor, put your item in this doc. Bullet/note form is fine. Even if your topic gets in late, we do our best to cover it.

For more information about the office hours and how to join, see [Office Hours](https://kops.sigs.k8s.io/welcome/office_hours/)
