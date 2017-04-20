---
title: How the Weave Net Docker Network Plugin Works
menu_order: 20
---


The Weave Net plugin actually provides *two* network drivers to Docker - one named `weavemesh` that can operate without a cluster store and another one named `weave` that can only work with one (like Docker's overlay driver).

### `weavemesh` driver

* Weave Net handles all co-ordination between hosts (referred to by Docker as a "local scope" driver)
* Supports a single network only. A network named `weave` is automatically created for you.
* Uses Weave Net's partition tolerant IPAM

If you do create additional networks using the `weavemesh` driver, containers attached to them will be able to communicate with containers attached to `weave`. There is no isolation between those networks.

### `weave` driver

* This runs in what Docker calls "global scope", which requires an external cluster store
* Supports multiple networks that must be created using `docker network create --driver weave ...`
* Used with Docker's cluster-store-based IPAM

There's no specific documentation from Docker on using a cluster
store, but the first part of
[Getting Started with Docker Multi-host Networking](https://github.com/docker/docker/blob/master/docs/userguide/networking/get-started-overlay.md) is a good place to start.

>**Note:** In the case of multiple networks using the `weave` driver, all containers are on the same virtual network but Docker allocates their addresses on different subnets so they cannot talk to each other directly.

The plugin accepts the following options via `docker network create ... --opt`:

 * `works.weave.multicast` -- tells weave to add a static IP
   route for multicast traffic onto its interface.

>**Note:** If you connect a container to multiple Weave networks, at
   most one of them can have the multicast route enabled.  The `weave`
   network created when the plugin is first launched has the multicast
   option turned on, but for any networks you create it defaults to off.

**See Also**

 * [Integrating Docker via the Network Plugin](/site/plugin.md)

