# Notes on gossip support

* We use weave mesh's gossip functionality, which supports gossip of a CRDT
* protokube listens on 0.0.0.0:3999
* dns-controller listens on 0.0.0.0:3998
* The seed for dns-controller is protokube, discovered on 127.0.0.1:3999
* The real seeding is done by protokube, which currently finds peers by querying the cloud provider

## DNS

* We implement a dnsprovider backed by our local gossip state
* We write to `/etc/hosts`; this is sort of hacky but avoids the need for a custom local resolver