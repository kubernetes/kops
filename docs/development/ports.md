# Port usage

This document includes the port used by system components,
so we can avoid port collisions.

| Port | Description                            |
|------|----------------------------------------|
| 22   | SSH                                    |
| 443  | Kubernetes API                         |
| 179  | Calico                                 |
| 2380 | etcd main peering                      |
| 2381 | etcd events peering                    |
| 3998 | dns gossip - protokube                 |
| 3999 | dns gossip - dns-controller            |
| 4001 | etcd main client                       |
| 4002 | etcd events client                     |
| 4789 | VXLAN                                  |
