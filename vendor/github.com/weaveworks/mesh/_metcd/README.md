# metcd

metcd implements the [etcd](https://github.com/coreos/etcd)
 [V3 API](https://github.com/coreos/etcd/blob/master/Documentation/rfc/v3api.md)
 on top of Weave Mesh.

**Note** that this package no longer compiles due to changes in etcd upstream.
The code remains for historical purposes.

# Caveats

- We only partially implement the etcd V3 API. See [etcd_store.go](https://github.com/weaveworks/mesh/blob/master/metcd/etcd_store.go) for details.
- Snapshotting and compaction are not yet implemented.

## Usage

```go
ln, err := net.Listen("tcp", ":8080")
if err != nil {
	panic(err)
}

minPeerCount := 3
logger := log.New(os.Stderr, "", log.Lstdflags)
server := metcd.NewDefaultServer(minPeerCount, logger)

server.Serve(ln)
```

To have finer-grained control over the mesh, use [metcd.NewServer](http://godoc.org/github.com/weaveworks/mesh/metcd#NewServer).
See [metcdsrv](https://github.com/weaveworks/mesh/tree/master/metcd/metcdsrv/main.go) for a complete example.
