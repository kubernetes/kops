# Cluster API
## What is the Cluster API?

The Cluster API is a Kubernetes project to bring declarative, Kubernetes-style
APIs to cluster creation, configuration, and management. It provides optional,
additive functionality on top of core Kubernetes.

Note that Cluster API effort is still in the prototype stage while we get
feedback on the API types themselves. All of the code here is to experiment with
the API and demo its abilities, in order to drive more technical feedback to the
API design. Because of this, all of the prototype code is rapidly changing.

To learn more, see the full [Cluster API proposal][proposal].

## Get involved!

* Join our Cluster API working group sessions
  * Weekly on Wednesdays @ 11:00 PT (19:00 UTC) on [Zoom][zoomMeeting]
  * Previous meetings: \[ [notes][notes] | [recordings][recordings] \]
* Chat with us on [Slack](http://slack.k8s.io/): #sig-cluster-lifecycle

## Getting Started
### Prerequisites
* `kubectl` is required, see [here](http://kubernetes.io/docs/user-guide/prereqs/).

### Prototype implementations
* [gcp machine controller](https://github.com/kubernetes/kube-deploy/blob/master/cluster-api-gcp/README.md)

## How to use the API

To see how to build tooling on top of the Cluster API, please check out a few examples below:

* [upgrader](tools/upgrader/README.md): a cluster upgrade tool.
* [repair](tools/repair/README.md): detect problematic nodes and fix them.
* [machineset](tools/machineset/README.md): a client-side implementation of MachineSets for declaratively scaling Machines.

[proposal]: https://docs.google.com/document/d/1G2sqUQlOYsYX6w1qj0RReffaGXH4ig2rl3zsIzEFCGY/edit#
[notes]: https://docs.google.com/document/d/16ils69KImmE94RlmzjWDrkmFZysgB2J4lGnYMRN89WM/edit#heading=h.xqb69epnpv
[recordings]: https://www.youtube.com/watch?v=I9764DRBKLI&list=PL69nYSiGNLP29D0nYgAGWt1ZFqS9Z7lw4
[gcpSDK]: https://cloud.google.com/sdk/downloads
[zoomMeeting]: https://zoom.us/j/166836624
