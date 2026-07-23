# Command line options

The `dns-controller` executable takes the following command line options:

* `--dns` - DNS provider we should use. Valid options are: `aws-route53`, 
  `google-clouddns`, `openstack-designate`, `scaleway`, and `digitalocean`.
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
