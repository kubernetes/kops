# Gossip DNS

Gossip-based clusters uses a peer-to-peer network for propagating the K8s API address instead of normal DNS.
This means that no hosted zone is needed for the cluster.

Gossip does not suffer potential disruptions due to the DNS TTL as the propagation is almost instant.

Gossip is also the only option if you want to deploy a cluster in any of the China of GovCloud AWS regions as Route 53 is not available there.

## Configuring a cluster to use Gossip

The only thing you need to do in order to use gossip-based DNS is to use the `k8s.local` suffix for the cluster domain name.

## Accessing the cluster

### Kubernetes API

When using gossip mode, you have to expose the kubernetes API using a loadbalancer. Since there is no hosted zone for gossip-based clusters, you simply use the load balancer address directly. The user experience is identical to standard clusters. Kops will add the ELB DNS name to the kops-generated kubernetes configuration.

### Bastion

If you are using [bastion hosts](bastion.md), it is a bit tricky to find the bastion address name. On AWS, you can run the following command:

```
kops toolbox dump -ojson | grep 'bastion.*elb.amazonaws.com'
```



