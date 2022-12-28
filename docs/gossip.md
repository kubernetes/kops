# Gossip DNS

Gossip-based clusters use a peer-to-peer network instead of externally hosted DNS for propagating the K8s API address.
This means that an externally hosted DNS service is not needed.

Gossip does not suffer potential disruptions due to out of date records in DNS caches as the propagation is almost instant.

Gossip is also the only option if you want to deploy a cluster in any of the AWS regions without Route 53, such as the China and GovCloud ones.

There are two supported gossip protocols: weaveworks mesh and memberlist. Memberlist is suggested on larger clusters due to known high resource usage of the mesh protocol on larger clusters.
## Configuring a cluster to use Gossip

In order to use gossip-based DNS, configure the cluster domain name to end with `.k8s.local`.

## Accessing the cluster

### Kubernetes API

When using gossip mode, you have to expose the kubernetes API using a loadbalancer. Since there is no hosted zone for gossip-based clusters, you simply use the load balancer address directly. The user experience is identical to standard clusters. kOps will add the ELB DNS name to the kops-generated kubernetes configuration.

### Bastion

If you are using [bastion hosts](bastion.md), it is a bit tricky to find the bastion address name. On AWS, you can run the following command:

```
kops toolbox dump -ojson | grep 'bastion.*elb.amazonaws.com'
```

## Switching to memberlist from weaveworks mesh

Weaveworks mesh is known to spike exponentially with CPU, memory and network bandwidth usage on cluster scale up. [issue 7429](https://github.com/kubernetes/kops/issues/7427), [issue 7436](https://github.com/kubernetes/kops/issues/7436), [issue 13974](https://github.com/kubernetes/kops/issues/13974)

### 1 Turn on memberlist as secondary protocol

Add `gossipConfig` to your cluster configuration:

```bash
kops edit cluster
```

```yaml
  gossipConfig:
    protocol: mesh
    listen: 0.0.0.0:3999
    secondary:
      listen: 0.0.0.0:4000
      protocol: memberlist

  dnsControllerGossipConfig:
    protocol: mesh
    listen: 0.0.0.0:3998
    seed: 127.0.0.1:3999
    secondary:
      protocol: memberlist
      listen: 0.0.0.0:3993
```

Note that this will not flag nodes as *Needs update*; you will have to do *rolling-update* with `--force` flag.

Perform a rolling-update of control-plane nodes only.

Check that dns-controller is restarted and now uses the new gossip protocol (you can see that under args).

```bash
kubectl get pod -n kube-system | grep dns-controller
kubectl get pod -n kube-system | grep dns-controller | awk '{print $1}' | xargs kubectl get pod -nkube-system -oyaml
```

```
  - args:
    - --watch-ingress=false
    - --dns=gossip
    - --gossip-protocol=mesh
    - --gossip-listen=0.0.0.0:3998
    - --gossip-seed=127.0.0.1:3999
    - --gossip-protocol-secondary=memberlist
    - --gossip-seed-secondary=127.0.0.1:3993
    - --zone=staging.stats.superbet.com
    - --internal-ipv4
    - --zone=*/*
    - -v=2
```

### 2 Switch memberlist to primary and set secondary protocol to empty string

```yaml
  gossipConfig:
    protocol: memberlist
    listen: 0.0.0.0:4000
    secondary:
      protocol: ""

  dnsControllerGossipConfig:
    protocol: memberlist
    listen: 0.0.0.0:3993
    seed: 127.0.0.1:4000
    secondary:
      protocol: ""
```

Note that this will not flag nodes as *Needs update*; you will have to do *rolling-update* with `--force` flag.

Perform a rolling-update of worker nodes and then a rolling-update of control-plane nodes

**!Perform a rolling-update of control-plane nodes last, othwerwise all nodes that still have mesh as primary will get into NotReady state!**

**Do not remove secondary protocol completly, otherwise new nodes won't join** [Duplicate metrics issue, Protokube with memberlist gossip DNS doesnt startup](https://github.com/kubernetes/kops/issues/9006)
