# Serving Certificates, Authentication, and Authorization

The apiserver-builder creates API servers that are set up to use a variety
of authentication and authorization options designed for use with an
existing Kubernetes cluster.  Together, these options form the *delegated
authentication and authorization* pattern.

In this document, we'll refer to API servers generated using
apiserver-builder as *addon API servers*.

## Certificates Overview

Several of the authentication methods that make up delegated
authentication make use of client and CA certificates.

CA (Certificate Authority) certificates are used to delegate trust.
Whenever something trusts the CA, it can trust any certificates *signed*
by the CA private key by verifying the signature using the CA public
certificate.

If a certificate is not signed by a separate CA, it is instead
*self-signed*. A self-signed certificate must either be trusted directly
(instead of being trusted indirectly by trusting a CA), or not trusted at
all.  Generally, our client CA certificates will be self-signed, since
they represent the "root" of our trust relationship: clients must
inherently trust the CA.

For the API servers created by apiserver-builder, there are three
*different* important CAs (and these really should be different):

1. a serving CA: this CA signs "serving" certificates, which are used to
   encrypt communication over HTTPS.  The same CA used to sign the main
   Kubernetes API server serving certificate pair may also be used to sign
   the addon API server serving certificates, but a different CA
   may also be used.

   By default, addon API servers automatically generate self-signed
   certificates if no serving certificates are passed in, making this CA
   optional. However, in a real setup, you'll need this CA so that clients
   can easily trust the identity of the addon API server.

2. a client CA: this CA signs client certificates, and is used by the
   addon API server to authenticate users based on the client certificates
   they submit.  The same client CA may be used for both the main
   Kubernetes API server as well as addon API servers, but a different CA
   may also be used. Using the same CA ensures that identity trust works
   the same way between the main Kubernetes API server and the addon API
   servers.

   As an example, the default cluster admin user generated in many
   Kubernetes setups uses client certificate authentication. Additionally,
   controllers or non-human clients running outside the cluster often use
   certificate-based authentication.

3. a RequestHeader client CA: this special CA signs proxy client
   certificates.  Clients presenting these certificates are effectively
   trusted to masquerade as any other identity.  When running behind the
   API aggregator, this *must* be the same CA used to sign the
   aggregator's proxy client certificate.  When not running with an
   aggregator (e.g. pre-Kubernetes-1.7, without a separate aggregator
   pod), this simply needs to exist.

### Generating certificates

The Kubernetes documentation has a [detailed
section](https://kubernetes.io/docs/admin/authentication/#creating-certificates)
on how to create certificates several different ways.  For convenience,
we'll reproduce the basics using the `openssl` and `cfssl` commands below
(you can install `cfssl` using `go get -u
github.com/cloudflare/cfssl/cmd/...`).

In the common case, all three CA certificates referenced above already
exist as part of the main Kubernetes cluster setup.

In case you need to generate any of the CA certificate pairs mentioned
above yourself, you can do so using the following command (see below for
appropriate values of `$PURPOSE`):

```shell
export PURPOSE=<purpose>
openssl req -x509 -sha256 -new -nodes -days 365 -newkey rsa:2048 -keyout ${PURPOSE}-ca.key -out ${PURPOSE}-ca.crt -subj "/CN=ca"
echo '{"signing":{"default":{"expiry":"43800h","usages":["signing","key encipherment","'${PURPOSE}'"]}}}' > "${PURPOSE}-ca-config.json"
```

This generates a certificate and private key for the CA, as well as
a signing configuration used by `cfssl` below. `$PURPOSE` should be set to
one of `serving`/`server`, `client`, or `requestheader-client`, as
detailed above in the [certificates overview](#certificates-overview).

These CA certificates are self-signed; no "higher-level" CAs are signing
these CA certificates, so they represent the "roots" of your trust
relationship.

To generate a serving certificate keypair (see the [serving
certificates](#serving-certificates) section for more details), you can
use the following commands:

```shell
export SERVICE_NAME=<service>
export ALT_NAMES='"<service>.<namespace>","<service>.<namespace>.svc"'
echo '{"CN":"'${SERVICE_NAME}'","hosts":['${ALT_NAMES}'],"key":{"algo":"rsa","size":2048}}' | cfssl gencert -ca=server-ca.crt -ca-key=server-ca.key -config=server-ca-config.json - | cfssljson -bare apiserver
```

`<service>` should be the name of the Service for the addon API server,
and `<namespace>` is the name of the namespace in which the server will
run.

This will create a pair of files named `apiserver-key.pem` and
`apiserver.pem`.  These are the private key and public certificate,
respectively.  The private key and certificate are commonly referred to
with `.key ` and `.crt` extensions, respectively: `apiserver.key` and
`apiserver.crt`.

### Serving Certificates

In order to securely serve your APIs over HTTPS, you'll need serving
certificates. By default, a set of self-signed certificates are generated
by addon API servers. However, clients have no way to trust these (since
they are self-signed, there is no separate CA), so in production
deployments, or deployments running behind an API server aggregator, you
should use manually generated CA certificates.

By default, addon API servers server looks for these certificates in the
`/var/run/kubernetes` directory, although this may be overridden using the
`--cert-dir` option.  The files must be named `apiserver.crt` and
`apiserver.key`.

## Authentication

There are three components to the delegated authentication setup,
described below:

- [Client Certificate Authentication](#client-certificate-authentication)
- [Delegated Token Authentication](#delegated-token-authentication)
- [RequestHeader Authentication](#requestheader-authentication)

### Client Certificate Authentication

Client certificate authentication authenticates clients who connect using
certificates signed by a given CA (as specified by the [*client CA
certificate*](#certificates-overview)).  This same mechanism is also often
used by the main Kubernetes API server.

Generally, the default admin user in a cluster connects with client
certificate authentication.  Additionally, off-cluster non-human clients
often use client certificate authentication.

By default, a main Kubernetes API server configured with the
`--client-ca-file` option automatically creates a ConfigMap called
`extension-apiserver-authentication` in the `kube-system` namespace,
populated with the client CA file.  Addon API servers use this CA
certificate as the CA used to verify client authentication. This way,
client certificate users who can authenticate with the main Kubernetes
system can also authenticate with addon API servers.

See the [delegated token authentication](#delegated-token-authentication)
section for more information about how addon API servers contact the main
Kubernetes API server to access this ConfigMap.

If you wish to use a different client CA certificate to verify client
certificate authentication, you can manually pass the `--client-ca-file`
option to your addon API server.

See the [x509 client
certificates](https://kubernetes.io/docs/admin/authentication/#x509-client-certs)
section of the Kubernetes documentation for more information.

### Delegated Token Authentication

Delegated token authentication authenticates clients who pass in a token
using the `Authorization: Bearer $TOKEN` HTTP header.  This is the common
authentication method used by most human Kubernetes clients, as well as
in-cluster non-human clients.

In this case, addon API servers extract the token from the HTTP request,
and verify it against another API server using a `TokenReview`. In common
cases, this is the main Kubernetes API server.  This allows users who are
can authentication with the main Kubernetes system to also authenticate
with addon API servers.

By default, the addon API servers search for the connection information
and credentials that are automatically injected into every pod running on
a Kubernetes cluster in order to connect to the main Kubernetes API
server.

If you do not wish to have your addon API server authenticate against the
same cluster that it is running on, or if it is running outside of
a cluster, you can pass the `--authentication-kubeconfig` option to the
addon API server to specify a different Kubeconfig file to use to connect.

The [Webhook token
authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication)
method described in the Kubernetes authentication documentation works
similarly in principal to delegated token authentication, except that we
use an existing Kubernetes cluster instead of an external webhook.

### RequestHeader Authentication

RequestHeader authentication authenticates connections from API server
proxies, which themselves have already authenticated the client.  It works
similarly to [client certificate
authentication](#client-certificate-authentication): it validates the
certificate of the proxy using a CA certificate.  However, it then allows
the proxy to masquerade as any other user, by reading a series of headers
set by the proxy. This allows addon API servers to run behind the API
server aggregator.

By default, addon API servers attempt to pull the requestheader client CA
certificate and appropriate header names from the
`extension-apiserver-authentication` ConfigMap mentioned above in
[client-certificate-authentication](#client-certificate-authentication).
The main Kubernetes API server populates this if it was configured with
the `--requestheader-client-ca-file` option (and optionally associated
`--requestheader-` options).

However, some API servers are not configured with the
`--requestheader-client-ca-file` option.  In these cases, you must pass
the `--requestheader-client-ca-file` option directly to the addon API
server. Any API server proxies (such as the API server aggregator) need to
have client certificates signed by this CA certificate in order to
properly pass their authentication information through to addon API
servers.

Alternatively, you can pass the `--authentication-skip-lookup` flag to
addon API servers.  However, this will *also* disable client certificate
authentication unless you manually pass the corresponding
`--client-ca-file` flag.

In addition to the CA certificate, you can also configure a number of
additional options.  See the [authenticating
proxy](https://kubernetes.io/docs/admin/authentication/#authenticating-proxy)
section of the Kubernetes documentation for more information.

### Authorization

Addon API servers use delegated authorization.  This means that they query
for authorization against the main Kubernetes API server using
a `SubjectAccessReview`, allowing cluster admins to store policy for addon
API servers in the same place as the policy used for the main Kubernetes
API server, and in the same format (e.g. Kubernetes RBAC).

By default, the addon API servers search for the connection information
and credentials that are automatically injected into every pod running on
a Kubernetes cluster in order to connect to the main Kubernetes API
server.

If you do not wish to have your addon API server authenticate against the
same cluster that it is running on, or if it is running outside of
a cluster, you can pass the `--authorization-kubeconfig` option to the
addon API server to specify a different Kubeconfig file to use to connect.

### RBAC Rules

By default, Kubernetes ships with RBAC, (**R**esource **B**ased **A**ccess
**C**ontrol) enabled by default, with some standard policy.  This means
that in order for your addon API server to be able to delegate
authentication (for [Delegated Token
Authentication](#delegated-token-authentication) and authorization, you'll
need to create several role bindings.

First, to allow the addon API server to delegate authentication and
authorization requests to the main Kubernetes API server, you'll need to
add a cluster role binding for the cluster role `system:auth-delegator`.

Then, you'll need to create a role binding for the
`extension-apiserver-authentication-reader` in the `kube-system` namespace.
This allows the addon API server to read the client CA file and
RequestHeader client CA file from the `extension-apiserver-authentication`
ConfigMap in the `kube-system` namespace.
