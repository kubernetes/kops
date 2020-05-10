# Gossip DNS

Gossip-based clusters use a peer-to-peer network instead of externally hosted DNS for propagating the K8s API address.
This means that an externally hosted DNS service is not needed.

Gossip does not suffer potential disruptions due to out of date records in DNS caches as the propagation is almost instant.

Gossip is also the only option if you want to deploy a cluster in any of the AWS regions without Route 53, such as the China and GovCloud ones.

## Configuring a cluster to use Gossip

In order to use gossip-based DNS,  configure the cluster domain name to end with `.k8s.local`.

## Accessing the cluster

### Kubernetes API

When using gossip mode, you have to expose the kubernetes API using a loadbalancer. Since there is no hosted zone for gossip-based clusters, you simply use the load balancer address directly. The user experience is identical to standard clusters. Kops will add the ELB DNS name to the kops-generated kubernetes configuration.

### Bastion

If you are using [bastion hosts](bastion.md), it is a bit tricky to find the bastion address name. On AWS, you can run the following command:

```
kops toolbox dump -ojson | grep 'bastion.*elb.amazonaws.com'
```