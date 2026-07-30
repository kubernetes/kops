# Gossip DNS

!!! warning
    Gossip DNS support was removed in kOps 1.37. Existing gossip clusters must [migrate to None-DNS](#migrating-from-gossip-to-none-dns) using kOps 1.36 before upgrading to kOps 1.37. See [issue #18240](https://github.com/kubernetes/kops/issues/18240) for background and discussion.

Gossip-based clusters use a peer-to-peer network, instead of externally hosted DNS, to propagate the Kubernetes API address. Before None-DNS, this was the only way to run a cluster without an externally hosted DNS zone, for example in AWS regions without Route 53, such as the China and GovCloud ones.

Its successor is None-DNS (`--dns=none`), introduced in kOps 1.26. None-DNS covers the same use cases without running a gossip mesh on every node. Worker nodes bootstrap directly against the API load balancer, and the cluster does not publish any DNS records.

## Deprecation and removal

Gossip DNS has been deprecated since kOps 1.29, when newly created clusters started defaulting to None-DNS. Since kOps 1.35, `kops create cluster` no longer creates gossip clusters at all: a cluster name ending in `.k8s.local` now creates a None-DNS cluster. Only clusters created as gossip by older kOps versions still use gossip.

The main reasons for the removal were:

* Both gossip implementations depended on unmaintained upstream libraries (`weaveworks/mesh` and a fork of `hashicorp/memberlist`). They ran in the privileged `protokube` daemon on every node.
* Gossip peer discovery required broad cloud permissions on every node, including workers. These permissions exposed the cluster topology to anyone who could read a node's instance role.

The removal timeline was:

* **kOps 1.36**: gossip clusters switched to hybrid bootstrap (see below), which removed gossip from worker nodes.
* **kOps 1.37**: gossip support was removed entirely. New gossip clusters are rejected. Existing gossip clusters must migrate to None-DNS using kOps 1.36 before they can upgrade.

As part of the removal, nodes no longer install or run the `protokube` component. Its only remaining responsibility was gossip DNS. The `gossipConfig` and `dnsControllerGossipConfig` cluster spec fields are deprecated and have no effect. Validation rejects any cluster spec that sets either field, so remove both fields before upgrading.

## Changes in kOps 1.36

kOps 1.36 made the migration to None-DNS as small a step as possible:

* Gossip clusters use a hybrid bootstrap path: control-plane nodes keep using gossip, while worker nodes bootstrap directly against the API load balancer. Workers no longer run `protokube`, no longer join the gossip mesh, and no longer need cloud permissions for peer discovery.
* Gossip seed discovery is restricted to control-plane nodes, and cloud credentials are no longer exported to worker nodes.
* When a cluster stops using gossip, kOps removes the now-unused `dns-controller` deployment automatically.

## Migrating from gossip to None-DNS

Migrate with kOps 1.36, before you upgrade to kOps 1.37. The migration keeps the cluster name: the certificates, the state store location, and the kubeconfig context do not change. Access to the cluster does not change either: the kubeconfig keeps using the API load balancer address. Once the DNS topology is `None`, the `.k8s.local` suffix no longer implies gossip.

1. Upgrade the cluster to kOps 1.36 and apply the changes:

    ```
    kops reconcile cluster --yes
    ```

    This moves the cluster to the hybrid bootstrap path and removes gossip from the worker nodes.

2. Disable gossip by setting the DNS topology to `None`. Run `kops edit cluster` and set:

    ```yaml
    spec:
      topology:
        dns:
          type: None
    ```

    On AWS, None-DNS requires a Network Load Balancer for the API. If the cluster still uses a Classic Load Balancer, also set `spec.api.loadBalancer.class: Network`.

3. Apply the changes:

    ```
    kops reconcile cluster --yes
    ```

    This replaces the control-plane nodes, which stop running gossip. kOps removes the unused `dns-controller` deployment automatically. On kOps versions before 1.36, it had to be deleted manually with `kubectl -n kube-system delete deployment dns-controller`.

4. Validate the cluster:

    ```
    kops validate cluster --wait 10m
    ```
