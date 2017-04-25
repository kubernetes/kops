---
title: Integrating Docker via the API Proxy
menu_order: 50
---

The Docker API proxy automatically attaches containers to the Weave
network when they are started using the ordinary Docker
[command-line interface](https://docs.docker.com/reference/commandline/cli/)
or the [remote API](https://docs.docker.com/reference/api/docker_remote_api/).

### <a name="attaching-containers"></a>Attaching Containers to a Weave Network

There are three ways to attach containers to a Weave network (which method to use is 
entirely up to you):

**1.** The Weave Net Docker API Proxy. See [Setting Up the Weave Net Docker API Proxy](#weave-api-proxy).  

**2.**  The Docker Network Plugin framework. The Docker Network Plugin is used when 
Docker containers are started with the --net flag, for example: 

`docker run --net <docker-run-options>`

**Where,** 

 * `<docker-run-options>` are the [docker run options](https://docs.docker.com/engine/reference/run/) 
 you give to your container on start 

Note that if a Docker container is started with the --net flag, then the Weave Docker API Proxy
is automatically disabled and is not used to attach containers. 
See [Integrating Docker via the Network Plugin](/site/plugin.md).

**3.** Containers can also be attached to the Weave network with `weave attach` commands. This method also
does not use the Weave Docker API Proxy. 
See [Dynamically Attaching and Detaching Containers](/site/using-weave/dynamically-attach-containers.md).

### <a name="weave-api-proxy"></a>Setting Up The Weave Net Docker API Proxy

The proxy sits between the Docker client (command line or API) and the
Docker daemon, and intercepts the communication between the two. You can
start it simultaneously with the router and weaveDNS via `launch`:

    host1$ weave launch

or independently via `launch-proxy`:

    host1$ weave launch-router && weave launch-proxy

The first form is more convenient. But only `launch-proxy` can be passed configuration arguments.
Therefore if you need to modify the default behaviour of the proxy, you must use `launch-proxy`.

By default, the proxy decides where to listen based on how the
launching client connects to Docker. If the launching client connected
over a UNIX socket, the proxy listens on `/var/run/weave/weave.sock`. If
the launching client connects over TCP, the proxy listens on port
12375, on all network interfaces. This can be adjusted using the `-H`
argument, for example:

    host1$ weave launch-proxy -H tcp://127.0.0.1:9999

If no TLS or listening interfaces are set, TLS is auto-configured
based on the Docker daemon's settings, and the listening interfaces are
auto-configured based on your Docker client's settings.

Multiple `-H` arguments can be specified. If you are working with a
remote docker daemon, then any firewalls in between need to be
configured to permit access to the proxy port.

All docker commands can be run via the proxy, so it is safe to adjust
your `DOCKER_HOST` to point at the proxy. Weave Net provides a convenient
command for this:

    host1$ eval $(weave env)
    host1$ docker ps

The prior settings can be restored with

    host1$ eval $(weave env --restore)

Alternatively, the proxy host can be set on a per-command basis with

    host1$ docker $(weave config) ps

The proxy can be stopped independently with

    host1$ weave stop-proxy

or in conjunction with the router and weaveDNS via `stop`.

If you set your `DOCKER_HOST` to point at the proxy, you should revert
to the original settings prior to stopping the proxy.


**See Also**

 * [Using The Weave Docker API Proxy](/site/weave-docker-api/using-proxy.md)
 * [Securing Docker Communications With TLS](/site/weave-docker-api/securing-proxy.md)
 * [Launching Containers With Weave Run (without the Proxy)](/site/weave-docker-api/launching-without-proxy.md)


