# Aggregation

While the API servers created by the apiserver-builder can be accessed
directly, it is most convinient to use them with the Kubernetes API server
aggregator.  This allows multiple APIs in the cluster to appear as if they
were being served by a single API server, so that cluster components, and
kubectl, can continue communicating as normal without special logic to
discover the locations of the different APIs.

In this document, we'll refer to API servers generated with
apiserver-builder as *addon API servers*.  We'll speak in terms of
a sample addon API server which serves an API group called `wardle`, with
an single namespaced resource `flunders`, abbreviated `fl`.

Understanding how the aggregator works with addon API servers requires
understanding how [authentication and authorization](./auth.md) works.  If
you have not yet read that section, please do.

## Enabling the Aggregator

The API aggregator is integrated into the main Kubernetes API server in
Kubernetes 1.7+, but must be run as a separate pod in Kubernetes 1.6.  In
Kubernetes 1.7+, it is only served on the secure port of the API server.

In Kubernetes 1.7, in order for the aggregator to work properly, it needs
to have the appropriate certificates.  First, ensure that you have
RequestHeader CA certificates, and client certificates signed by that CA,
as discussed in the [authentication and authorization](./auth.md) section.

These certificates must be passed to the main Kubernetes API server using
the flags `--proxy-client-cert-file` and `--proxy-client-key-file`.  This
allows the aggregator to identify itself when making requests, so that
addon API servers can use its delegated RequestHeader authentication.

Enabling the aggregator in 1.6 is outside the scope of this document.  The
aggregator directory itself provides examples on how to do so in the
Kubernetes 1.6 release.

## kubectl and Discovery

When `kubectl get fl` is called, `kubectl` does not initially know what an
`fl` is, or how to get one.  In order to determine this information, it
uses a mechanism called *discovery*.

All Kubernetes API servers serve discovery information.  To get this
information, `kubectl` first queries `https://$SERVER/apis`, which lists
all available API groups and versions, as well as the *preferred version*.
Then, for each of these API groups and versions, it queries
`https://$SERVER/apis/$GROUP/$VERSION`.  This returns a list of resources,
along with whether the resource is *namespaced*, which operations (e.g.
`get` or `list`) it supports, as well as any short names (like `fl` for
`flunders`).  See the section on [viewing discovery
information](#viewing-discovery-information) below for more information on
how this works.

Once `kubectl` has retrieved the entire set of available resources for
a particular cluster, it can then determine how to operate on the exposed
resources.  For example, in our case above, `kubectl` determines that `fl`
is a shortname for `flunders`, and that `flunders` are a namespaced
resource that's part of the API group `wardle`, which has a preferred
version of `v1alpha`.  Thus, `kubectl` will attempt a query of
`https://$SERVER/apis/wardle/v1alpha1/namespaces/$NS/flunders/`.

Controller managers also commonly use discovery information to determine
whether or not they should run: if the resources that they require are not
present of the cluster, controller managers can just fail to start.

You can see the list of API group versions returned by discover using the
`kubectl api-versions` command.

You can see discovery in action by running `kubectl` with the
`--v=6` flag to see requests and responses, or `--v=8` to see the full
bodies.

## Registering APIs

In order for your API to appear in the discovery information served by the
aggregator, it must be registered with the aggregator.  In order to do
this, the aggregator exposes a normal Kubernetes API group called
`apiregistration.k8s.io`, with a single resource, APIService.

Each APIService corresponds to a single group-version, and different
versions of a single API can be backed by different APIService objects.

Let's take a look at the APIService for `wardle/v1alpha1`, using `kubectl
get apiservice v1alpha1.wardle -o yaml`:

```yaml
apiVersion: apiregistration.k8s.io/v1beta1
kind: APIService
metadata:
  name: v1alpha1.wardle
spec:
  caBundle: <base64-encoded-serving-ca-certificate>
  group: wardle
  version: v1alpha1
  groupPriorityMinimum: 1000
  versionPriority: 15
  service:
    name: wardle-server
    namespace: wardle-namespace
status:
  ...
```

Notice that this is a Kubernetes API object like any other: it has a name,
a spec, a status, etc.  There are several important fields in the spec.

The first important field is `caBundle`.  This is the base64-encoded
version of a CA certificate that can be used to verify the serving
certificates of the API server.  The aggregator will check that the
serving certificates are for a hostname of `<service>.<namespace>.svc` (in
the case of the APIService above, that's
`wardle-server.wardle-namespace.svc`).

Next are the `group` and `version` fields.  These determine which
group-version the APIService describes.

The `groupPriorityMinimum` and `versionPriority` fields communicate how
the aggregator orders API groups and versions for discovery.  The higher
the priority, the earlier in the discovery list the group-version will
appear.  API groups are sorted according to the highest value of
`groupPriorityMinimum` across each APIService for their versions.  Then,
each version within that group is sorted according to the
`versionPriority`.

For instance, suppose we also had a `wardle/v1` API version with
`groupPriorityMinimum: 2000` and `versionPriority: 20`, and an API
group-version `bloops/v1` at `groupPriorityMinimum: 1500`.  Then, both
entries for the `wardle` group would appear before the `bloops` group, and
`wardle/v1` would appear before `wardle/v1alpha1`.

Finally, the `service` field determines how the aggregator actually
connects to the addon API server.  It will look up the service IP of the
service described, and connect to that.  As mentioned above, however, it
validates certificates based on hostname.  Since the aggregator validates
certificates on hostnames, and service IPs may be re-used, your
certificates should not contain the service IP.

## Proxying

In addition to serving discovery information and registering API groups
and servers, the aggregator acts as proxy.

The aggregator "natively" serves discovery information that lets us list
the available API groups, and the corresponding versions.  For other
requests, such as fetching the available resources in a group-version, or
making a request against an API, the aggregator contacts a registered API
server.

When the aggregator receives a request that it needs to proxy, it first
performs authentication using the authentication methods configured for
the main Kubernetes API server, as well as authorization.  Once it has
completed authentication, it records the authentication information in
headers, and forwards the request to the appropriate addon API server.

For instance, suppose we make a request

```
GET /apis/wardle/v1alpha1/namespaces/somens/flunders/foo
```

using the admin client certificates.  The aggregator will verify the
certificates, strip them from the request, and add the `X-Remote-User:
system:admin` header.

The aggregator will then connect to the wardle server, verifying the wardle
server's service certificates using the CA certificate from the `caBundle`
field of the APIService object for `wardle/v1alpha1`, and submitting it's
own proxy client certificates to identify itself to the wardle server.

The wardle server will receive the modified request, verify the proxy
client certificates against it's RequestHeader CA certificate, and treat
the request as if it had come from the `system:admin` user, as marked in
the `X-Remote-User` header.  The wardle server can then proceed along with
its normal serving logic, validating authorization and returning a result
to the aggregator.  The aggregator then returns the result back to us.

It's important to note that while the aggregator performs authentication
and authorization, this does not mean that addon API servers can skip
performing authentication and authorization themselves.  The API servers
are exposed behind regular Kubernetes services, so clients are free to
access the API servers directly.

## Troubleshooting Tips

### RBAC

Kubernetes enables RBAC (**R**esource **B**ased **A**ccess **C**ontrol) by
default.  Aggregated API servers commonly need serveral RBAC roles
assinged to them in order to function properly.  See the [RBAC
section](./auth.md#rbac) for more details.

### Viewing Discovery Information

It can occasionally be useful to directly look at the discovery
information being published by the aggregator and your API server. To do
this, we can use `kubectl get --raw` to perform an HTTP request with our
authentication information attached.  We can use `jq` to pretty-print the
resulting JSON.

First, we can see the list of API groups and versions at the `/apis/`
endpoint with `kubectl get --raw /apis`:

```json
{
  "kind": "APIGroupList",
  "apiVersion": "v1",
  "groups": [
    {
      "name": "wardle",
      "versions": [
        {
            "groupVersion": "wardle/v1alpha1",
            "version": "v1alpha1"
        }
      ],
      "preferredVersion": {
        "groupVersion": "wardle/v1alpha1",
        "version": "v1alpha1"
      }
    },
    ...
  ]
}
```

Notice each group is listed in order of priority as discussed above, and
each group contains a list of versions, as well as the preferred version.
`kubectl` uses the preferred version if none is manually specified.  If we
wish to see just the group-versions for the `wardle` API group, we can use
the URL `/apis/wardle`.

Next, we can investigate the details of the `wardle/v1alpha1` API
group-version with `kubectl get --raw /apis/wardle/v1alpha1`:

```json
{
  "kind": "APIResourceList",
  "apiVersion": "v1",
  "groupVersion": "wardle/v1alpha1",
  "resources": [
    {
      "name": "flunders",
      "singularName": "",
      "namespaced": true,
      "kind": "Flunder",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "fl"
      ],
      "categories": [
        "all"
      ]
    }
  ]
}
```

This provides us with a list of resources.  `name` and `singularName`
provide cues as to the plural and singular resource names, while
`shortNames` provides a list of abbreviations.  Together, these form the
names that `kubectl` will accept when operating on this resource.

The `categories` field provides an indication as to which bulk operations,
like `kubectl get all`, the resource should be included in.  Currently,
this is not honored by `kubectl`, but will be in the future.

The `kind` field indicates the Kind of the object involved with this
resource.   Finally, `verbs` describes the Kubernetes operations available
on the resource.  For instance, read-only resources might only have `get`,
`list`, and `watch`.

### kubectl Discovery Caching

Since querying all of the discovery endpoints for every kubectl request
would be expensive, `kubectl` caches discovery information in the same
folder as where your kubeconfig lives, based on the name of the cluster to
which you are connecting.   If you encounter a case where an API service
exists, and curling to the API works, but `kubectl` does not see the
resource as existing, clearing this cache may resolve the issue.

You can check which discovery information `kubectl` is using by passing it
the `--v=8` flag.

### Certificate Issues

If the aggregator returns x509 validation errors, or your requests from
the aggregator are coming through as `system:unauthenticated`, it can be
useful to troubleshoot the certificate issues using command line
utilities.

A good first step is trying the requests yourself with `curl`.  First,
fetch the APIService object and decode the caBundle into a file:

```shell
$ kubectl get apiservice -o jsonpath='{ .spec.caBundle }' | tee /tmp/the-ca.crt
-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
```

Then, find the aggregator's proxy client certificates, as discussed above.
We'll call them `/path/to/proxy-client.crt` and
`/path/to/proxy-client.key`.

Finally, find the IP address of your API server service:

```shell
$ kubectl get service <apiserver> -o jsonpath='{ .spec.clusterIP }'
<clusterIP>
```

Make your request, resolving the service hostname to the service IP, using
the appropriate certificates:

```shell
$ curl --cacert /tmp/the-ca.crt --cert /path/to/proxy-client.crt --key /path/to/proxy-client.key \
  --resolve <service>.<namespace>.svc:443:<clusterIP> -v \
  https://<service>.<namespace>.svc/apis/wardle/v1alpha1/namespaces/default/flunders/theflundinator
{
  ...
}
```

If this doesn't work, start substituting out parts.  For instance, if it
works with `-k` instead of `--cacert /tmp/the-ca.crt`, your CA or serving
certificates are incorrect.  At that point, you can use `openssl` to
investigate certificate issues.  `openssl` is a complicated tool, but the
most useful commands for these kind of issues are:

- Show the details of a certificate in text form: `openssl x509 -noout
  -text -in /path/to/serving-ca.crt`.  This is useful for making sure the
  certificate has the correct DNS names encoded in it.  Make sure that
  either the "Common Name", or one of the "Alternative Names", is set to
  `<service>.<namespace>.svc`.  The alternative names should each be
  prefixed with `DNS:` for DNS names.

- Verify that a certificate is signed by a CA: `openssl verify -CAfile
  /path/to/the-ca.crt /path/to/the-certificate.crt`.  This is useful for
  making sure your certificates and CAs work properly.

- Show the certificate actually being served by an addon API server:
  `openssl s_client -connect <service-cluster-ip>:443`.  This will print
  out the certificate served by the API server, plus some additional
  information.  If you copy the certificate (from `-----BEGIN
  CERTIFICATE-----` to `-----END CERTIFICATE-----`) and place it in
  a file, you can use the `openssl x509` command above to inspect the
  details and confirm that it's the correct certificate.
