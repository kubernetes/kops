# Overview

# ipsets

The policy controller maintains a number of ipsets which are
subsequently referred to by the iptables rules used to effect network
policy specifications. These ipsets are created, modified and
destroyed automatically in response to Pod, Namespace and
NetworkPolicy object updates from the k8s API server:

* A `hash:ip` set per namespace, containing the IP addresses of all
  pods in that namespace
* A `list:set` per distinct (across all network policies in all
  namespaces) namespace selector mentioned in a network policy,
  containing the names of any of the above hash:ip sets whose
  corresponding namespace labels match the selector
* A `hash:ip` set for each distinct (within the scope of the
  containing network policy's namespace) pod selector mentioned in a
  network policy, containing the IP addresses of all pods in the
  namespace whose labels match that selector

IPsets are implemented by the kernel module `xt_set`, without which
weave-npc will not work.

ipset names are generated deterministically from a string
representation of the corresponding label selector. Because ipset
names are limited to 31 characters in length, this is done by taking a
SHA hash of the selector string and then printing that out as a base
85 string with a "weave-" prefix e.g.:

    weave-k?Z;25^M}|1s7P3|H9i;*;MhG

Because pod selectors are scoped to a namespace, we need to make sure
that if the same selector definition is used in different namespaces
that we maintain distinct ipsets. Consequently, for such selectors the
namespace name is prepended to the label selector string before
hashing to avoid clashes.

# iptables chains

The policy controller maintains two iptables chains in response to
changes to pods, namespaces and network policies. One chain contains
the ingress rules that implement the network policy specifications,
and the other is used to bypass the ingress rules for namespaces which
have an ingress isolation policy of `DefaultAllow`.

## Dynamically maintained `WEAVE-NPC-DEFAULT` chain

The policy controller maintains a rule in this chain for every
namespace whose ingress isolation policy is `DefaultAllow`. The
purpose of this rule is simply to ACCEPT any traffic destined for such
namespaces before it reaches the ingress chain.

```
iptables -A WEAVE-NPC-DEFAULT -m set --match-set $NSIPSET dst -j ACCEPT
```

## Dynamically maintained `WEAVE-NPC-INGRESS` chain

For each namespace network policy ingress rule peer/port combination:

```
iptables -A WEAVE-NPC-INGRESS -p $PROTO [-m set --match-set $SRCSET] -m set --match-set $DSTSET --dport $DPORT -j ACCEPT
```

## Static `WEAVE-NPC` chain

Static configuration:

```
iptables -A WEAVE-NPC -m state --state RELATED,ESTABLISHED -j ACCEPT
iptables -A WEAVE-NPC -m state --state NEW -j WEAVE-NPC-DEFAULT
iptables -A WEAVE-NPC -m state --state NEW -j WEAVE-NPC-INGRESS
```

# Steering traffic into the policy engine

To direct traffic into the policy engine:

```
iptables -A FORWARD -o weave -m physdev ! --physdev-out vethwe-bridge -j WEAVE-NPC
iptables -A FORWARD -o weave -m physdev ! --physdev-out vethwe-bridge -j DROP
```

Note this only affects traffic which egresses the bridge on a physical
port which is not the Weave Net router - in other words, it is
destined for an application container veth. The following traffic is
affected:

* Traffic bridged between local application containers
* Traffic bridged from the router to a local application container
* Traffic originating from the internet destined for nodeports - this
  is routed via the FORWARD chain to a container pod IP after DNAT

The following traffic is NOT affected:

* Traffic bridged from a local application container to the router
* Traffic originating from processes in the host network namespace
  (e.g. kubelet health checks)
* Traffic routed from an application container to the internet

The above mechanism relies on the kernel module `br_netfilter` being
loaded and enabled via `/proc/sys/net/bridge/bridge-nf-call-iptables`.

See these resources for helpful context:

* http://ebtables.netfilter.org/br_fw_ia/br_fw_ia.html
* https://commons.wikimedia.org/wiki/File:Netfilter-packet-flow.svg
