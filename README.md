# Kubernetes Operations (kops)

[![Build Status](https://travis-ci.org/kubernetes/kops.svg?branch=master)](https://travis-ci.org/kubernetes/kops) [![Go Report Card](https://goreportcard.com/badge/k8s.io/kops)](https://goreportcard.com/report/k8s.io/kops)

The easiest way to get a production grade Kubernetes cluster up and running.

## What is kops?

We like to think of it as `kubectl` for clusters.

`kops` lets you deploy grade Kubernetes clusters from the command line, with
options that support HA Masters. Kubernetes Operations supports deploying
Kubernetes on Amazon Web Services (AWS) and support for more platforms is planned.

## Launching a Kubernetes cluster hosted on AWS

Check out our [tutorial](/docs/aws.md) for launching a Kubernetes cluster hosted
on AWS.

<p align="center">
  <img src="/docs/img/demo.gif" width="885" > </image>
</p>

## Features

* Automated Kubernetes cluster [CRUD](/docs/commands.md) for ([AWS](/docs/aws.md))
* Highly Available (HA) Kubernetes Masters Setup
* Uses a state-sync model for **dry-run** and automatic **idempotency**
* Custom support for `kubectl` [add-ons](/docs/addons.md)
* Kops can generate [Terraform configuration](/docs/terraform.md)
* Based on a simple meta-model defined in a directory tree
* Command line [autocomplete](/docs/cli/kops_completion.md)
* Community support

## Installing

`kubectl` is required, see [here](http://kubernetes.io/docs/user-guide/prereqs/).

<!-- Move this to an install guide -->
    <!-- RE: Move this to an install guide. I like the idea of having the instal on the front page - especially if it's an easy install. It makes the tool look excitng and fun. Also, there is a link below towards the install guide. @kris-nova -->

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

See [building notes](/docs/build.md) for more information.

At this time, Windows is not a supported platform.

## Getting involved!

Want to contribute to kops? We would love the extra help from the community. We
have a quickstart guide on [adding a feature](/docs/development/adding_a_feature.md).

Kops also has time set aside every other week to offer help and guidance to the
community. Kops maintainers have agreed to set aside time specifically dedicated
to working with newcomers, helping with PRs, and discussing new features.

We recommend letting us know **beforehand** if you plan on attending so we can
have time to prepare for the call.

| Maintainer   | Schedule      |  URL |
|--------------|---------------|-------|
| [@justinsb](https://github.com/justinsb)             |  2nd / 4th Friday 9am PDT | [Zoom](https://zoom.us/my/k8ssigaws) |
| [@chrislovecnm](https://github.com/chrislovecnm)     |  2nd / 4th Friday 9am PDT | [Zoom](https://zoom.us/my/k8ssigaws) |
| [@kris-nova](https://github.com/kris-nova)           |  2nd / 4th Friday 9am PDT | [Zoom](https://zoom.us/my/k8ssigaws) |

Reach out to us on [kubernetes slack](https://github.com/kubernetes/community#slack-chat).
A great place to get involved or ask questions is [#sig-cluster-lifecycle](https://kubernetes.slack.com/?redir=%2Fmessages%2Fsig-cluster-lifecycle%2F).

## Other Resources

 - Create [kubecfg settings for kubectl](/docs/tips.md#create-kubecfg-settings-for-kubectl)
 - Set up [add-ons](/docs/addons.md), to add important functionality to Kubernetes
 - Learn about [InstanceGroups](/docs/instance_groups.md); change
 instance types, number of nodes, and other options
 - Read about [networking options](/docs/networking.md)
 - Look at our [other interesting modes](/docs/commands.md#other-interesting-modes)
 - Full command line interface [documentation](/docs/cli/kops.md)

## History

View our [changelog](HISTORY.md)

## GitHub Issues


#### Bugs

If you think you have found a bug please follow the instructions below.

- Please spend a small amount of time giving due diligence to the issue tracker. Your issue might be a duplicate.
- Set `-v 10` command line option and save the log output. Please paste this into your issue.
- Note you version of `kops`, and the command line options you are using
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
