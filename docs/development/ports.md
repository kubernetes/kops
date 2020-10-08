# Port usage

This document includes the port used by system components,
so we can avoid port collisions.

See also pkg/wellknownports/wellknownports.go


| Port | Description                              |
|------|------------------------------------------|
| 22   | SSH                                      |
| 443  | Kubernetes API                           |
| 179  | Calico                                   |
| 2380 | etcd main peering                        |
| 2381 | etcd events peering                      |
| 3988 | kops controller serving port             |
| 3989 | node local dns health check              |
| 3990 | Kube API health check                    |
| 3991 | etcd-manager - cilium - grpc             |
| 3992 | etcd-manager - cilium - quarantined      |
| 3993 | dns gossip - dns-controller - memberlist |
| 3994 | etcd-manager - main - quarantined        |
| 3995 | etcd-manager - events - quarantined      |
| 3996 | etcd-manager - main - grpc               |
| 3997 | etcd-manager - events - grpc             |
| 3998 | dns gossip - protokube - weave mesh      |
| 3999 | dns gossip - dns-controller - weave mesh |
| 4000 | protokube gossip member list             |
| 4001 | etcd main client                         |
| 4002 | etcd events client                       |
| 4789 | VXLAN                                    |
| 6942 | Cilium operator prometheus port          |
| 9090 | Cilium prometheus port                   |
| 9091 | Cilium hubble prometheus port            |
