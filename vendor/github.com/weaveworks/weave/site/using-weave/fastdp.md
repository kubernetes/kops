---
title: Using Fast Datapath
menu_order: 96
---


The most important thing to know about fast datapath is that you don't need to configure anything before using this feature. If you are using Weave Net 1.2 or greater, fast datapath (`fastdp`) is automatically enabled.

When Weave Net cannot use the fast data path between two hosts, it falls back to a slower packet forwarding approach called `sleeve`. Selecting the fastest forwarding approach is automatic, and is determined on a connection-by-connection basis. For example, a Weave network spanning two data centers might use fast data path within the data centers, but not for the more constrained network link between them. 

See [How Fastdp Works](/site/how-it-works/fastdp-how-it-works.md) for a more in-depth discussion of this feature. 

### Disabling Fast Datapath

You can disable fastdp by enabling the `WEAVE_NO_FASTDP` environment variable at `weave launch`:

    $ WEAVE_NO_FASTDP=true weave launch

### Fast Datapath and Encryption

Fast datapath implements encryption using IPsec which is configured with IP
transformation framework (XFRM) provided by the Linux kernel.

Each encrypted dataplane packet is encapsulated into [ESP](https://tools.ietf.org/html/rfc2406),
thus in some networks a firewall rule for allowing ESP traffic needs to be installed. E.g. Google
Cloud Platform denies ESP packets by default.

See [How Weave Implements Encryption](/site/how-it-works/encryption-implementation.md)
for more details for the fastdp encryption.

### Viewing Connection Mode Fastdp or Sleeve

Weave Net automatically uses the fastest datapath for every connection unless it encounters a situation that prevents it from working. To ensure that Weave Net can use the fast datapath:

 * Avoid Network Address Translation (NAT) devices
 * Open UDP port 6784 (This is the port used by the Weave routers)
 * Ensure that `WEAVE_MTU` fits with the `MTU` of the intermediate network (see below)

The use of fast datapath is an automated connection-by-connection decision made by Weave Net, and because of this, you may end up with a mixture of connection tunnel types. If fast datapath cannot be used for a connection, Weave Net falls back to the `sleeve` "user space" packet path.

Once a Weave network is set up, you can query the connections using the `weave status connections` command:

    $ weave status connections
    <-192.168.122.25:43889  established fastdp a6:66:4f:a5:8a:11(ubuntu1204)

Where fastdp indicates that fast datapath is being used on a connection. If fastdp is not shown, the field displays `sleeve` indicating Weave Net's fall-back encapsulation method:

    $ weave status connections
    <- 192.168.122.25:54782  established sleeve 8a:50:4c:23:11:ae(ubuntu1204)

### <a name="mtu"></a>Packet size (MTU)

The Maximum Transmission Unit, or MTU, is the technical term for the
limit on how big a single packet can be on the network. Weave Net
defaults to 1376 bytes, but you can set a smaller size if your
underlying network has a tighter limit, or set a larger size for
better performance.

The underlying network must be able to deliver packets of the size
specified plus overheads of around 84-87 bytes (the final MTU should be
divisible by four), or else Weave Net will
fall back to Sleeve for that connection.  This requirement applies
to _every path_ between peers. 

To specify a different MTU, before launching Weave Net set the
environment variable `WEAVE_MTU`.  For example, for a typical "jumbo
frame" configuration:

    $ WEAVE_MTU=8916 weave launch host2 host3

**See Also**

 * [Using Weave Net](/site/using-weave.md)
 * [How Fastdp Works](/site/how-it-works/fastdp-how-it-works.md)
 * [Performance measurements](/blog/weave-docker-networking-performance-fast-data-path/)
