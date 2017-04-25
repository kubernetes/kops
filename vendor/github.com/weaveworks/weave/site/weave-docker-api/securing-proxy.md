---
title: Securing the Docker Communications With TLS
menu_order: 20
---

If you are [connecting to the docker daemon with
TLS](https://docs.docker.com/articles/https/), you most likely want
to do the same when connecting to the proxy. The proxy
automatically detects the Docker daemon's TLS configuration, and
attempts to duplicate it. 

In the standard auto-detection case you can launch a TLS-enabled proxy as follows:

    host1$ weave launch-proxy

To disable auto-detection of TLS configuration, you can either pass
the `--no-detect-tls` flag, or you can manually configure the proxy's TLS using
the same TLS-related command-line flags supplied to the Docker
daemon. 

For example, if you generated your certificates and keys
into the Docker host's `/tls` directory, launch the proxy using:

    host1$ weave launch-proxy --tlsverify --tlscacert=/tls/ca.pem \
             --tlscert=/tls/server-cert.pem --tlskey=/tls/server-key.pem

The paths to your certificates and key must be provided as absolute
paths as they exist on the Docker host.

Because the proxy connects to the Docker daemon at
`unix:///var/run/docker.sock`, you must ensure that the daemon is actually
listening there. To do ensure this, pass the `-H unix:///var/run/docker.sock` option when starting the Docker daemon,
in addition to the `-H` options for configuring the TCP listener. See
[the Docker documentation](https://docs.docker.com/articles/basics/#bind-docker-to-another-host-port-or-a-unix-socket)
for an example.

With the proxy running over TLS, you can configure the Docker
client to use TLS on a per-invocation basis by running:

    $ docker --tlsverify --tlscacert=ca.pem --tlscert=cert.pem \
         --tlskey=key.pem -H=tcp://host1:12375 version

or, [by default](https://docs.docker.com/articles/https/#secure-by-default), using:

    $ mkdir -pv ~/.docker
    $ cp -v {ca,cert,key}.pem ~/.docker
    $ eval $(weave env)
    $ export DOCKER_TLS_VERIFY=1
    $ docker version

This is exactly the same configuration used when connecting to the
Docker daemon directly, except that the specified port is the Weave
proxy port.


**See Also**

 * [Setting Up The Weave Docker API Proxy](/site/weave-docker-api.md)
