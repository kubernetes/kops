# meshconn

meshconn implements [net.PacketConn](https://golang.org/pkg/net/#PacketConn) on top of mesh.
Think of it as UDP with benefits:
 NAT and bastion host (DMZ) traversal,
 broadcast/multicast in networks where this is normally not possible e.g. EC2,
 and an up-to-date, queryable memberlist.

meshconn supports [net.Addr](https://golang.org/pkg/net/#Addr) of the form `weavemesh://<PeerName>`.
By default, `<PeerName>` is a hardware address of the form `01:02:03:FD:FE:FF`.
Other forms of PeerName e.g. hashes are supported.

meshconn itself is largely stateless and has best-effort delivery semantics.
As a future experiment, it could easily be amended to have basic resiliency guarantees.
Also, at the moment, PacketConn read and write deadlines are not supported.
