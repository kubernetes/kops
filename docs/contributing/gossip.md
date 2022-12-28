# Notes on gossip support


* kOps supports weave mesh's and mememberlist gossip protocol
* weave mesh's gossip is default
* Switching to memberlist protocol is recommended for clusters of 150+ nodes. [issue 7429](https://github.com/kubernetes/kops/issues/7427), [issue 7436](https://github.com/kubernetes/kops/issues/7436), [issue 13974](https://github.com/kubernetes/kops/issues/13974)

# Notes on mesh's gossip

* We use weave mesh's gossip functionality, which supports gossip of a CRDT
* protokube listens on 0.0.0.0:3999
* dns-controller listens on 0.0.0.0:3998
* The seed for dns-controller is protokube, discovered on 127.0.0.1:3999
* The real seeding is done by protokube, which currently finds peers by querying the cloud provider

## DNS

* We implement a dnsprovider backed by our local gossip state
* We write to `/etc/hosts`; this is sort of hacky but avoids the need for a custom local resolver