---
title: Troubleshooting Weave Net
menu_order: 110
---


 * [Basic Diagnostics](#diagnostics)
 * [Status Reporting](#weave-status)
   - [List connections](#weave-status-connections)
   - [List peers](#weave-status-peers)
   - [List DNS entries](#weave-status-dns)
   - [JSON report](#weave-report)
   - [List attached containers](#list-attached-containers)
 * [Stopping Weave](#stop)
 * [Reboots](#reboots)
 * [Snapshot Releases](#snapshots)

## <a name="diagnostics"></a>Basic Diagnostics

Check the version of Weave Net you are running using:

    weave version

If it is not the latest version, as shown in the list of
[releases](https://github.com/weaveworks/weave/releases), then it is
recommended you upgrade using the
[installation instructions](https://github.com/weaveworks/weave#installation).

To check the Weave Net container logs:

    docker logs weave

A reasonable amount of information, and all errors, get logged there.

The log verbosity may be increased by using the
`--log-level=debug` option during `weave launch`. To log information on
a per-packet basis use `--pktdebug` - but be warned, as this can produce a
lot of output.

Another useful debugging technique is to attach standard packet
capture and analysis tools, such as tcpdump and wireshark, to the
`weave` network bridge on the host.

## <a name="weave-status"></a>Status Reporting

A status summary can be obtained using `weave status`:

```
$ weave status

        Version: 1.1.0 (up to date; next check at 2016/04/06 12:30:00)

        Service: router
       Protocol: weave 1..2
           Name: 4a:0f:f6:ec:1c:93(host1)
     Encryption: disabled
  PeerDiscovery: enabled
        Targets: [192.168.48.14 192.168.48.15]
    Connections: 5 (1 established, 1 pending, 1 retrying, 1 failed, 1 connecting)
          Peers: 3 (with 5 established, 1 pending connections)
 TrustedSubnets: none

        Service: ipam
         Status: ready
          Range: 10.32.0.0-10.47.255.255
  DefaultSubnet: 10.32.0.0/12

        Service: dns
         Domain: weave.local.
            TTL: 1
        Entries: 9

        Service: proxy
        Address: tcp://127.0.0.1:12375

       Service: plugin
    DriverName: weave
```

The terms used here are explained further at
[How Weave Net Works](/site/how-it-works.md).

 * **Version** - shows the Weave Net version. If checkpoint is enabled (i.e.
`CHECKPOINT_DISABLE` is not set), information about existence of a new version
will be shown.

 * **Protocol**- indicates the Weave Router inter-peer
communication protocol name and supported versions (min..max).

 * **Name** - identifies the local Weave Router as a peer on the
Weave network. The nickname shown in parentheses defaults to the name
of the host on which the Weave container was launched. It
can be overridden by using the `--nickname` argument at `weave
launch`.

 * **Encryption** - indicates whether
[encryption](/site/how-it-works/encryption.md) is in use for communication
between peers.

 * **PeerDiscovery** - indicates whether
[automatic peer discovery](/site/ipam/allocation-multi-ipam.md) is
enabled (which is the default).

 * **Targets** - are the number of hosts that the local Weave Router has been
asked to connect to at `weave launch` and `weave connect`. The
complete list can be obtained using `weave status targets`.

 * **Connections** - show the total number connections between the local Weave
Router and other peers, and a break down of that figure by connection
state. Further details are available with
[`weave status connections`](#weave-status-connections).

 * **Peers** - show the total number of peers in the network, and the total
number of connections peers have to other peers. Further details are
available with [`weave status peers`](#weave-status-peers).

 * **TrustedSubnets** - show subnets which the router trusts as specified by the `--trusted-subnets` option at `weave launch`.



### <a name="weave-status-connections"></a>List Connections

Connections between Weave Net peers carry control traffic over TCP and
data traffic over UDP. For a connection to be fully established, the
TCP connection and UDP datapath must be able to transmit information
in both directions. Weave Net routers check this regularly with
heartbeats. Failed connections are automatically retried, with an
exponential back-off.

To view detailed information on the local Weave Net router's type `weave status connections`:

```
$ weave status connections
<- 192.168.48.12:33866   established unencrypted fastdp 7e:21:4a:70:2f:45(host2) mtu=1410
<- 192.168.48.13:60773   pending     encrypted   fastdp 7e:ae:cd:d5:23:8d(host3)
-> 192.168.48.14:6783    retrying    dial tcp4 192.168.48.14:6783: no route to host
-> 192.168.48.15:6783    failed      dial tcp4 192.168.48.15:6783: no route to host, retry: 2015-08-06 18:55:38.246910357 +0000 UTC
-> 192.168.48.16:6783    connecting
```

The columns are as follows:

 * Connection origination direction (`->` for outbound, `<-` for
   inbound)
 * Remote TCP address
 * Status
    * `connecting` - first connection attempt in progress
    * `failed` - TCP connection or UDP heartbeat failed
    * `retrying` - retry of a previously failed connection attempt in
      progress; reason for previous failure follows
    * `pending` - TCP connection up, waiting for confirmation of UDP
      heartbeat
    * `established` - TCP connection and corresponding UDP path are up
 * Info - the failure reason for failed and retrying connections, or
   the encryption mode, data transport method, remote peer name and
   nickname for pending and established connections, mtu if known

### <a name="weave-status-peers"></a>List Peers

Detailed information on peers can be obtained with `weave status
peers`:

```
$ weave status peers
ce:31:e0:06:45:1a(host1)
   <- 192.168.48.12:39634   ea:2d:b2:e6:e4:f5(host2)         established
   <- 192.168.48.13:49619   ee:38:33:a7:d9:71(host3)         established
ea:2d:b2:e6:e4:f5(host2)
   -> 192.168.48.11:6783    ce:31:e0:06:45:1a(host1)         established
   <- 192.168.48.13:58181   ee:38:33:a7:d9:71(host3)         established
ee:38:33:a7:d9:71(host3)
   -> 192.168.48.12:6783    ea:2d:b2:e6:e4:f5(host2)         established
   -> 192.168.48.11:6783    ce:31:e0:06:45:1a(host1)         established
```

This lists all peers known to this router, including itself.  Each
peer is shown with its name and nickname, then each line thereafter
shows another peer that it is connected to, with the direction, IP
address and port number of the connection.  In the above example,
`host3` has connected to `host1` at `192.168.48.11:6783`; `host1` sees
the `host3` end of the same connection as `192.168.48.13:49619`.

### <a name="weave-status-dns"></a>Listing DNS Entries

Detailed information on DNS registrations can be obtained with `weave
status dns`:

```
$ weave status dns
one          10.32.0.1       eebd81120ee4 4a:0f:f6:ec:1c:93
one          10.43.255.255   4fcec78d2a9b 66:c4:47:c6:65:bf
one          10.40.0.0       bab69d305cba ba:98:d0:37:4f:1c
three        10.32.0.3       7615b6537f74 4a:0f:f6:ec:1c:93
three        10.44.0.1       c0b39dc52f8d 66:c4:47:c6:65:bf
three        10.40.0.2       8a9c2e2ef00f ba:98:d0:37:4f:1c
two          10.32.0.2       83689b8f34e0 4a:0f:f6:ec:1c:93
two          10.44.0.0       7edc306cb668 66:c4:47:c6:65:bf
two          10.40.0.1       68a5e9c2641b ba:98:d0:37:4f:1c
```

The columns are as follows:

 * Unqualified hostname
 * IPv4 address
 * Registering entity identifier (typically a container ID)
 * Name of peer from which the registration originates

### <a name="weave-report"></a>Producing a JSON Report

    weave report

Produces a comprehensive dump of the internal state of the router,
IPAM and DNS services in JSON format, including all the information
available from the `weave status` commands. You can also supply a
Golang text template to `weave report` in a similar fashion to `docker
inspect`:

    $ weave report -f '{{.DNS.Domain}}' weave.local.

Weave Net adds a template function, `json`, which can be applied to get
results in JSON format.

    $ weave report -f '{{json .DNS}}'
    {"Domain":"weave.local.","Upstream":["8.8.8.8","8.8.4.4"],"Address":"172.17.0.1:53","TTL":1,"Entries":null}

### <a name="list-attached-containers"></a>Listing Attached Containers

    weave ps

Produces a list of all containers running on this host that are
connected to the Weave network, like this:

    weave:expose 7a:c4:8b:a1:e6:ad 10.2.5.2/24
    b07565b06c53 ae:e3:07:9c:8c:d4
    5245643870f1 ce:15:34:a9:b5:6d 10.2.5.1/24
    e32a7d37a93a 7a:61:a2:49:4b:91 10.2.8.3/24
    caa1d4ee2570 ba:8c:b9:dc:e1:c9 10.2.1.1/24 10.2.2.1/24

On each line are the container ID, its MAC address, then the list of
IP address/routing prefix length ([CIDR
notation](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing))
assigned on the Weave network. The special container name `weave:expose`
displays the Weave bridge MAC and any IP addresses added to it via the
`weave expose` command.

You can also supply a list of container IDs/names to `weave ps`, like this:

    $ weave ps able baker
    able ce:15:34:a9:b5:6d 10.2.5.1/24
    baker 7a:61:a2:49:4b:91 10.2.8.3/24

## <a name="stop"></a>Stopping Weave Net

To stop Weave Net, if you have configured your environment to use the
Weave Docker API Proxy, e.g. by running `eval $(weave env)` in your
shell, you must first restore the environment using:

    eval $(weave env --restore)

Then run:

    weave stop

Note that this leaves the local application container network intact.
Containers on the local host can continue to communicate, whereas
communication with containers on different hosts, as well as service
export/import, is disrupted but resumes once Weave is relaunched.

To stop Weave Net and to completely remove all traces of the Weave network on
the local host, run:

    weave reset

Any running application containers permanently lose connectivity
with the Weave network and will have to be restarted in order to
re-connect.

## <a name="reboots"></a>Reboots

All the containers started by `weave launch` are configured with the
Docker restart policy `--restart=always`, so they will come back again
on reboot. This can be disabled via:

    weave launch --no-restart

Note that the
[Weave Net Docker API Proxy](/site/weave-docker-api.md)
is responsible for reconfiguring the Weave router and re-attaching
application containers to the Weave network at startup, so if you
choose not to run it you must make arrangements for this
reconfiguration to take place. In this scenario, set up your favourite
process manager to run `weave launch-router` every time the machine
reboots.

## <a name="snapshots"></a>Snapshot Releases

Snapshot releases are published at times to provide previews of new
features, assist in the validation of bug fixes, etc. One can install the
latest snapshot release using:

    sudo curl -L git.io/weave-snapshot -o /usr/local/bin/weave
    sudo chmod a+x /usr/local/bin/weave
    weave setup

Snapshot releases report the script version as "unreleased",
and the container image versions as git hashes.



**See Also**

 * [Troubleshooting IPAM](/site/ipam/troubleshooting-ipam.md)
 * [Troubleshooting the Proxy](/site/weave-docker-api/using-proxy.md)
