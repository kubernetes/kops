# Kubernetes Operations (kops)

[![Build Status](https://travis-ci.org/kubernetes/kops.svg?branch=master)](https://travis-ci.org/kubernetes/kops) [![Go Report Card](https://goreportcard.com/badge/k8s.io/kops)](https://goreportcard.com/report/k8s.io/kops)  [![GoDoc Widget]][GoDoc]

[GoDoc]: https://godoc.org/k8s.io/kops
[GoDoc Widget]: https://godoc.org/k8s.io/kops?status.svg


The easiest way to get a production grade Kubernetes cluster up and running.

## What is kops?

We like to think of it as `kubectl` for clusters.

`kops` helps you create, destroy, upgrade and maintain production-grade, highly available, Kubernetes clusters from the command line.  AWS (Amazon Web Services) is currently officially supported, with GCE and VMware vSphere in alpha and other platforms planned.

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

* Automates the provisioning of Kubernetes clusters in ([AWS](/docs/aws.md))
* Deploys Highly Available (HA) Kubernetes Masters
* Supports upgrading from [kube-up](/docs/upgrade_from_kubeup.md)
* Built on a state-sync model for **dry-runs** and automatic **idempotency**
* Ability to output to Terraform [Terraform configuration](/docs/terraform.md)
* Supports custom Kubernetes [add-ons](/docs/addons.md)
* Command line [autocompletion](/docs/cli/kops_completion.md)
* Community supported!

## Installing

`kubectl` is required, see [here](http://kubernetes.io/docs/user-guide/prereqs/).

### OSX From Homebrew (Latest Stable Release)

```console
$ brew update && brew install kops
```
### Linux

Download the [latest release](https://github.com/kubernetes/kops/releases/latest), then:

```console
$ chmod +x kops-linux-amd64                 # Add execution permissions
$ mv kops-linux-amd64 /usr/local/bin/kops   # Move the kops to /usr/local/bin
```

### Developer From Source


Go 1.8+ and make are required. You may need to do a full build including pushing protokube, nodeup, and kops to s3.

See the [install notes](/docs/install.md) for more information.

```console
$ go get -d k8s.io/kops
$ cd ${GOPATH}/src/k8s.io/kops/
$ git checkout release
$ make
```

At this time, Windows is not a supported platform.

## History

See the [releases](https://github.com/kubernetes/kops/releases) for more
information on changes between releases.

## Getting involved and contributing!

Are you interested in contributing to kops? We, the maintainers and community, would love your suggestions, contributions, and help! We have a quick-start guide on [adding a feature](/docs/development/adding_a_feature.md).  Also, the maintainers can be contacted at any time to learn more about how to get involved.

In the interest of getting more new folks involved with kops, we are starting to tag issues with `good-starter-issue`. These are typically issues that have smaller scope but are good ways to start to get acquainted with the codebase.

We also encourage ALL active community participants to act as if they are maintainers, even if you don't have "official" write permissions. This is a community effort, we are here to serve the Kubernetes community. If you have an active interest and you want to get involved, you have real power! Don't assume that the only people who can get things done around here are the "maintainers".

We also would love to add more "official" maintainers, so show us what you can do!

What this means:

__Issues__
* Help read and triage issues, assist when possible.
* Point out issues that are duplicates, out of date, etc.
  - Even if you don't have tagging permissions, make a note and tag maintainers. (`/close`,`/dupe #127`)

__Pull Requests__
* Read and review the code. Leave comments, questions, and critiques (`/lgtm` )
* Download, compile, and run the code and make sure it does what the PR says and doesn't break something else
  - If you're ready to run someone else's code that interacts with your cloud provider account, you should probably actually understand what it does, first!


### Maintainers

* [@justinsb](https://github.com/justinsb)
* [@chrislovecnm](https://github.com/chrislovecnm)
* [@kris-nova](https://github.com/kris-nova)
* [@geojaz](https://github.com/geojaz)
* [@yissachar](https://github.com/yissachar)

## Office Hours

Kops maintainers set aside 1 hour every other week for **public** office hours for a video chat in which we strive to get to know the other developers either working on the project or interested in getting to know more about it. We tend to use this hour to discuss the current state of the kops project, to strategize about how to move it forward, ti discuss open and upcoming PRs, to demo, and to offer help and guidance to the community. Generally this time is focused on developers, although we will never turn a courteous participant away. Even if you've never actually installed kops, we're interested to have you stop by our office hours.

We encourage you to reach out **beforehand** if you plan on attending. Reaching out can be as simple as adding an item or your name to the [agenda](https://docs.google.com/document/d/12QkyL0FkNbWPcLFxxRGSPt_tNPBHbmni3YLY-lHny7E/edit) where we track notes from office hours.

Office hours, on [Zoom](https://zoom.us/my/k8ssigaws) video conference are on Fridays at [5pm UTC/12 noon ET/9 am US Pacific](http://www.worldtimebuddy.com/?pl=1&lid=100,5,8,12) every other week, on odd week numbered weeks.

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
