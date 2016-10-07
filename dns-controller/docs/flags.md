## zone

Pass a list of zones to determine which names can be updated.  Zones not permitted will be ignored
(but the default is to allow all zones).

The following syntax options are recognized:

`*` or `*/*` wildcard allowing all zones to be updated.  The default if no zones are specified, but
can also be used if some zones must be explicitly mapped.

`example.com` to permit updates in this zone, specified by name.  Use the ID syntax
if there are multiple zones with the same name.

`*/id` to permit updates in a zone, by id.

`example.com/id` to permit updates in the zone named example.com, by id.
