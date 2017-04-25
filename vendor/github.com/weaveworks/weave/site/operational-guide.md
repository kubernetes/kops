---
title: Operational Guide
menu_order: 45
---
This operational guide is intended to give you an overview of how to
operate and manage a Weave Network in production. It consists of three
main parts:

* A [glossary of concepts](/site/operational-guide/concepts.md) with
  which you will need to be familiar
* Detailed instructions for safely bootstrapping, growing and
  shrinking Weave networks in a number of different deployment
  scenarios:
    * An [interactive
      installation](/site/operational-guide/interactive.md), suitable
      for evaluation and development
    * A [uniformly configured
      cluster](/site/operational-guide/uniform-fixed-cluster.md) with
      a fixed number of initial nodes, suitable for automated
      provisioning but requiring manual intervention for resizing
    * A [heterogenous cluster](/site/operational-guide/autoscaling.md)
      comprising fixed and autoscaling components, suitable for a base
      load with automated scale-out/scale-in
    * A [uniformly configured
      cluster](/site/operational-guide/uniform-dynamic-cluster.md)
      with dynamic nodes, suitable for automated provisioning and
      resizing.
* A list of [common administrative
  tasks](/site/operational-guide/tasks.md), such as configuring Weave
  Net to start on boot, upgrading clusters, cleaning up peers and
  reclaiming IP address space

