# NLB required for ACM Certificates and Kubernetes 1.19+

Historically kOps has supported attaching an ACM certificate on the Listener of the API ELB via the Cluster's `spec.api.loadBalancer.sslCertificate`.
This changes the ELB listener from TCP to TLS, performing TLS termination between the client and ELB, and establishing its own TLS connections between the ELB and kube-apiserver pods.

kOps maintains a set of credentials per cluster that are used in kubeconfig files via `kops export kubecfg` and have `cluster-admin` RBAC permissions.
Historically this included creating a kubeconfig user with multiple authentication methods:
* Client certificate auth (`client-certificate-data` and `client-key-data` fields)
* Basic auth via HTTP request header (`username` and `password` fields)

Because the ELB was establishing its own TLS session with the client rather than passing it through to the kube-apiserver pods, the client certificate authentication would fail. `kubectl` and `client-go` would silently fallback to the basic auth credentials and succeed.

Kubernetes 1.16 deprecated the basic auth support by [deprecating the --basic-auth-file kube-apiserver flag](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.16.md#deprecations-and-removals).
Kubernetes 1.19 [removed basic auth support entirely](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.19.md#urgent-upgrade-notes).

This means that clients using the `kops export kubecfg` credentials can no longer fallback to basic auth, causing API requests to fail.

## How do I know if I'm affected?

Any AWS cluster with `spec.api.loadBalancer.sslCertificate` set and running Kubernetes 1.19+ is affected.

It is encouraged to perform the below migration on existing clusters prior to upgrading to Kubernetes 1.19.

## Solution

kOps 1.19.0 supports using a network load balancer (NLB) rather than a classic load balancer (CLB) for Kubernetes API access.
The NLB can be configured with a second listener configured for TCP, allowing the client certificate authentication to succeed by passing the TLS session through to the kube-apiserver pods.

Migrate from CLB to NLB with the following steps:
1. Set the Cluster's `spec.api.loadBalancer.class: Network`.
2. Run `kops update cluster --yes`.
   The second TCP listener will be configured automatically if `sslCertificate` is set.
   This will provision a new NLB and update the API DNS record to point to the new NLB.
3. Any clients using kOps' admin credentials will need to run `kops export kubecfg --admin`.
   This will use the secondary NLB port with its client cert auth user.
   The primary listener still uses the ACM certificate, preserving any external authentication mechanisms.
4. Manually delete the old API CLB through the AWS Management Console or CLI.
   The CLB will have a name of `api-<hyphenated-cluster-name>-<suffix>` and will not have any instances attached.
   A future version of kOps may do this automatically.

Terraform users can follow their normal workflow, confirming that `terraform plan` reports deletion of the ELB and creation of the NLB and its target groups.

API access through the load balancer will not be available for a few minutes while the NLB is being provisioned.
This should only affect external clients unless `spec.api.loadBalancer.useForInternalApi: true` is set.

### Gossip Clusters

Gossip clusters will require additional steps:
* Any kubeconfig `server`s created from `kops export kubecfg` will need to be regenerated due to the load balancer DNS name changing.
* The kube-apiserver server certificate needs to be reissued by the cluster CA to include the new load balancer DNS name:
  1. Delete the old server certificate with `kops delete secret master`
  2. Provision the new certificate with `kops update cluster --yes`
  3. Replace the master instances such that they use the new certificate with `kops rolling-update cluster --instance-group-roles Master --cloudonly`

Cluster validation and API access through the load balancer will not succeed until the masters have been replaced.

## More information

See [#9756](https://github.com/kubernetes/kops/issues/9756) for more information.