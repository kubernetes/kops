# Detailed description of arguments

## admin-access

`admin-access` controls the CIDR which can access the admin endpoints (SSH to each node, HTTPS to the master).

If not specified, no IP level restrictions will apply (though there are still restrictions, for example you need
a permitted SSH key to access the SSH service!).

Currently this can only be a single CIDR.

Examples:

**CLI:**

`--admin-access=18.0.0.0/8` to restrict to IPs in the 18.0.0.0/8 CIDR

**YAML:**

See the docs in [cluster_spec.md#adminaccess](cluster_spec.md#adminaccess)

## dns-zone

`dns-zone` controls the Route53 hosted zone in which DNS records will be created.  It can either by the name
of the hosted zone (`example.com`), or it can be the ID of the hosted zone (`Z1GABCD1ABC2DEF`)

Suppose you're creating a cluster named "dev.kubernetes.example.com`:

* You can specify a `--dns-zone=example.com` (you can have subdomains in a hosted zone)
* You could also use `--dns-zone=kubernetes.example.com`

You do have to set up the DNS nameservers so your hosted zone resolves.  kops used to create the hosted
zone for you, but now (as you have to set up the nameservers anyway), there doesn't seem much reason to do so!

If you don't specify a dns-zone, kops will list all your hosted zones, and choose the longest that
is a a suffix of your cluster name.  So for `dev.kubernetes.example.com`, if you have `kubernetes.example.com`,
`example.com` and `somethingelse.example.com`, it would choose `kubernetes.example.com`.  `example.com` matches
but is shorter; `somethingelse.example.com` is not a suffix-match.

Examples:

`--dns-zone=example.com` to use the hosted zone with a name of example.com

## UpdatePolicy

Cluster.Spec.UpdatePolicy

Values:

* `external` updates are performed by an external system (or manually), should not be automatically applied

* unset means to use the default policy, which is currently to apply OS security updates unless they require a reboot

## out

`out` determines the directory into which kubectl will write the target output.  It defaults to `out/terraform`
