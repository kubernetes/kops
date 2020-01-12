
HTTP Forward Proxy Support
==========================

It is possible to launch a Kubernetes cluster from behind an http forward proxy ("corporate proxy").  To do so, you will need to configure the `egressProxy` for the cluster.

It is assumed the proxy is already existing.  If you want a private topology on AWS, for example, with a proxy instead of a NAT instance, you'll need to create the proxy yourself.  See [Running in a shared VPC](run_in_existing_vpc.md).

This configuration only manages proxy configurations for Kops and the Kubernetes cluster.  We can not handle proxy configuration for application containers and pods.

## Configuration

Add `spec.egressProxy` port and url as follows

``` yaml
spec:
  egressProxy:
    httpProxy:
      host: proxy.corp.local
      port: 3128
```

Currently we assume the same configuration for http and https traffic.

## Proxy Excludes

Most clients will blindly try to use the proxy to make all calls, even to localhost and the local subnet, unless configured otherwise.  Some basic exclusions necessary for successful launch and operation are added for you at initial cluster creation.  If you wish to add additional exclusions, add or edit `egressProxy.excludes` with a comma separated list of hostnames.  Matching is based on suffix, ie, `corp.local` will match `images.corp.local`, and `.corp.local` will match `corp.local` and `images.corp.local`, following typical `no_proxy` environment variable conventions.

``` yaml
spec:
  egressProxy:
    httpProxy:
      host: proxy.corp.local
      port: 3128
    excludes: corp.local,internal.corp.com
```

## AWS VPC Endpoints and S3 access

If you are hosting on AWS have configured VPC "Endpoints" for S3 or other services, you may want to add these to the `spec.egressProxy.excludes`.  Keep in mind that the S3 bucket must be in the same region as the VPC for it to be accessible via the endpoint.
