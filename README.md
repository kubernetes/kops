# Kubernetes Operations (kops)

[![Build Status](https://travis-ci.org/kubernetes/kops.svg?branch=master)](https://travis-ci.org/kubernetes/kops) [![Go Report Card](https://goreportcard.com/badge/k8s.io/kops)](https://goreportcard.com/report/k8s.io/kops)  [![GoDoc Widget]][GoDoc]

[GoDoc]: https://godoc.org/k8s.io/kops
[GoDoc Widget]: https://godoc.org/k8s.io/kops?status.svg


The easiest way to get a production grade Kubernetes cluster up and running.

## What is kops?

We like to think of it as `kubectl` for clusters.

`kops` lets you deploy production-grade, highly available, Kubernetes clusters
from the command line.  Deployment is currently supported on Amazon Web
Services (AWS), with more platforms planned.

## Can I see it in action?

<p align="center">
  <a href="https://asciinema.org/a/97298">
  <img src="https://asciinema.org/a/97298.png" width="885"></image>
  </a>
</p>

## Launching a Kubernetes cluster hosted on AWS

To replicate the above demo, check out our [tutorial](/docs/aws.md) for
launching a Kubernetes cluster hosted on AWS.

## Features

* Automate the provisioning of Kubernetes clusters in ([AWS](/docs/aws.md))
* Deploy Highly Available (HA) Kubernetes Masters
* Supports upgrading from [kube-up](/docs/upgrade_from_kubeup.md)
* Built on a state-sync model for **dry-runs** and automatic **idempotency**
* Ability to generate [Terraform configuration](/docs/terraform.md)
* Supports custom [add-ons](/docs/addons.md) for `kubectl`
* Command line [autocompletion](/docs/cli/kops_completion.md)
* Community supported!

## Installing

`kubectl` is required, see [here](http://kubernetes.io/docs/user-guide/prereqs/).

### OSX From Homebrew (Latest Stable Release)

```console
$ brew update && brew install kops
```

### OSX From Homebrew (HEAD of master)

```console
$ brew update && brew install --HEAD kops
```

### Linux

Download the [latest release](https://github.com/kubernetes/kops/releases/latest), then:

```console
$ chmod +x kops-linux-amd64                 # Add execution permissions
$ mv kops-linux-amd64 /usr/local/bin/kops   # Move the kops to /usr/local/bin
```

### From Source

Go 1.7+ and make are required.

```console
$ go get -d k8s.io/kops
$ cd ${GOPATH}/src/k8s.io/kops/
$ git checkout release
$ make
```

See the [install notes](/docs/install.md) for more information.

At this time, Windows is not a supported platform.

## History

See the [releases](https://github.com/kubernetes/kops/releases) for more
information on changes between releases.

## Getting involved and contributing!

Are you interested in contributing to kops? We the maintainers and community would love your suggestions, contributions, and help! We have a quickstart guide on [adding a feature](/docs/development/adding_a_feature.md) and the maintainers can be contacted at any time to learn more about how to get involved. 

### Maintainers

* [@justinsb](https://github.com/justinsb)     
* [@chrislovecnm](https://github.com/chrislovecnm)
* [@kris-nova](https://github.com/kris-nova)
* [@geojaz](https://github.com/geojaz)
* [@yissachar](https://github.com/yissachar)

## Office Hours

Kops maintainers also have time (1 hour) set aside every other week to offer help and guidance to the
community. This is the maintainer's **office hours**. During this time we might work with newcomers, help with PRs, and discuss new features. Anything goes.

We encourage you to reach out **beforehand** if you plan on attending. We have an [agenda](https://docs.google.com/document/d/12QkyL0FkNbWPcLFxxRGSPt_tNPBHbmni3YLY-lHny7E/edit) where we track notes from office hours.

Office hours, on [Zoom](https://zoom.us/my/k8ssigaws) video
conference are on Fridays at [5pm UTC/9 am US Pacific](http://www.worldtimebuddy.com/?pl=1&lid=100,5,8,12) every other week, on odd week numbers.

You can check your week number using:

```bash
date +%V
```

The maintainers and other community members are generally available on the [kubernetes slack](https://github.com/kubernetes/community#slack-chat) in [#kops](https://kubernetes.slack.com/messages/kops/), so come find and chat with us about how kops can be better for you! 

## Other Resources

 - Create [kubecfg settings for kubectl](/docs/tips.md#create-kubecfg-settings-for-kubectl)
 - Set up [add-ons](/docs/addons.md), to add important functionality to Kubernetes
 - Learn about [InstanceGroups](/docs/instance_groups.md); change
 instance types, number of nodes, and other options
 - Read about [networking options](/docs/networking.md)
 - Look at our [other interesting modes](/docs/commands.md#other-interesting-modes)
 - Full command line interface [documentation](/docs/cli/kops.md)

## GitHub Issues

#### Bugs

If you think you have found a bug please follow the instructions below.

- Please spend a small amount of time giving due diligence to the issue tracker. Your issue might be a duplicate.
- Set `-v 10` command line option and save the log output. Please paste this into your issue.
- Note the version of kops you are running (from `kops version`), and the command line options you are using
- Open a [new issue](https://github.com/kubernetes/kops/issues/new)
- Remember users might be searching for your issue in the future, so please give it a meaningful title to helps others.
- Feel free to reach out to the kops community on [kubernetes slack](https://github.com/kubernetes/community#slack-chat)

#### Features

We also use the issue tracker to track features. If you have an idea for a feature, or think you can help kops become even more awesome follow the steps below.

- Open a [new issue](https://github.com/kubernetes/kops/issues/new)
- Remember users might be searching for your issue in the future, so please give it a meaningful title to helps others.
- Clearly define the use case, using concrete examples. EG: I type `this` and kops does `that`.
- Some of our larger features will require some design. If you would like to include a technical design for your feature please include it in the issue.
- After the new feature is well understood, and the design agreed upon we can start coding the feature. We would love for you to code it. So please open up a **WIP** *(work in progress)* pull request, and happy coding.