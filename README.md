# Kubernetes Operations (kops)

[![Build Status](https://travis-ci.org/kubernetes/kops.svg?branch=master)](https://travis-ci.org/kubernetes/kops) [![Go Report Card](https://goreportcard.com/badge/k8s.io/kops)](https://goreportcard.com/report/k8s.io/kops)

The easiest way to get a production Kubernetes cluster up and running.  

# What is kops?

We like to think of it as `kubectl` for clusters. 

kops lets you deploy production grade (and HA) Kubernetes clusters in the cloud from the command line.

#### Quickstart

Launching a Kubernetes cluster on [AWS](/docs/aws.md).

### Example on AWS

<p align="center">
  <img src="/docs/img/demo.gif"> </image>
</p>


### Features

* Automated Kubernetes cluster [CRUD](/docs/commands.md) for the cloud ([AWS](/docs/aws.md))
* HA (Highly Available) Kubernetes clusters
* Uses a state-sync model for **dry-run** and automatic **idempotency**
* Custom support for `kubectl` [add-ons](docs/addons.md)
* Kops can generate [Terraform configuration](/docs/terraform.md)
* Based on a simple meta-model defined in a directory tree
* Easy command line syntax
* Community support

# Installation

### Recommended
 
Download the [latest release](https://github.com/kubernetes/kops/releases/latest)

### History

View our [changelog](HISTORY.md)

### From Source

```
go get -d k8s.io/kops
cd ${GOPATH}/src/k8s.io/kops/
git checkout release
make
```

See [building notes](/docs/build.md) for more information.

# Getting involved!

Want to contribute to kops? We would love the extra help from the community. We have a quickstart guide on [adding a feature](/docs/adding_a_feature.md).

Kops also has time set aside every other week to offer help and guidance to the community. Kops maintainers have agreed to set aside time specifically dedicated to working with newcomers, helping with PRs, and discussing new features.

We recommend letting us know **beforehand** if you plan on attending so we can have time to prepare for the call.

| Maintainer   | Schedule      |  URL |
|--------------|---------------|-------|
| [@justinsb](https://github.com/justinsb)             |  2nd / 4th Friday 9am PDT | [Zoom](https://zoom.us/my/k8ssigaws) |
| [@chrislovecnm](https://github.com/chrislovecnm)     |  2nd / 4th Friday 9am PDT | [Zoom](https://zoom.us/my/k8ssigaws) |
| [@kris-nova](https://github.com/kris-nova)           |  2nd / 4th Friday 9am PDT | [Zoom](https://zoom.us/my/k8ssigaws) |

Reach out to us on [kubernetes slack](https://github.com/kubernetes/community#slack-chat). A great place to get involved or ask questions is [#sig-cluster-lifecycle](https://kubernetes.slack.com/?redir=%2Fmessages%2Fsig-cluster-lifecycle%2F)

# Other Resources 

 - Create [kubecfg settings for kubectl](/docs/tips.md#create-kubecfg-settings-for-kubectl)
 - Set up [add-ons](docs/addons.md), to add important functionality to Kubernetes
 - Learn about [InstanceGroups](docs/instance_groups.md), which let you change instance types, cluster sizes etc.. 
 - Read about [networking options](docs/networking.md), including a 50 node limit in the default configuration.
 - Look at our [other interesting modes](/docs/commands.md#other-interesting-modes).

# Bugs

If you think you have found a bug : 

- Set `--v=8` and save the log output 
- Open a [new issue](https://github.com/kubernetes/kops/issues/new)
- Feel free to reach out to the kops community on [kubernetes slack](https://github.com/kubernetes/community#slack-chat)
