# Command line options

The `dns-controller` executable takes the following command line options:

* `--dns` - DNS provider we should use. Valid options are: `aws-route53`, 
  `google-clouddns` or `coredns`.
* `--gossip-listen` - The address on which to listen if gossip is enabled.
* `--gossip-seed` - If set, will enable gossip zones and seed using the 
  provided address.
* `--gossip-secret` - Secret to use to secure the gossip protocol.
* `--zone` - Configure permitted zones and their mappings. See further notes 
  below.
* `--watch-ingress` - Watch for DNS records in `ingress` resources in addition 
  to `service` resources.

## zone

Pass a list of zones to determine which names can be updated.  Zones not 
permitted will be ignored (but the default is to allow all zones).

The following syntax options are recognized:

`*` or `*/*` wildcard allowing all zones to be updated.  The default if no
zones are specified, but can also be used if some zones must be explicitly
mapped.

`example.com` to permit updates in this zone, specified by name.  Use the ID 
syntax if there are multiple zones with the same name.

`*/id` to permit updates in a zone, by id.

`example.com/id` to permit updates in the zone named example.com, by id.
