## Creating a Subdomain That Uses Amazon Route 53 as the DNS Service without Migrating the Parent Domain
<hr>

You can create a subdomain that uses Amazon Route 53 as the DNS service without migrating the parent domain from another DNS service.

The procedure shall involve following steps:

- Create subdomain hosted zone
- Create NS record on the parent domain hosted zone

In this example, we use `example.com` as parent hosted zone.

## Create Subdomain
You want to keep those parent domain hosted zone records, so now lets create the subdomain.

On your `route 53` create the subdomain :

`Create Hosted zone`

Fill up the box `Domain Name:` with your subdomain : k8s.example.com

`Route 53` should generate your NS server like below in subdomain management console:

```
ns-613.awsdns-13.net.
ns-75.awsdns-04.com.
ns-1022.awsdns-35.co.uk.
ns-1149.awsdns-27.org.
```

Take note on these records.

## Create NS record on Parent domain hosted zone

Add / Create a NS record on the parent domain hosted zone with previous noted subdomain NS server records via parent domain management console.

After done, the result should like this from cli:

>dig ns k8s.example.com

```
;; ANSWER SECTION:
k8s.example.com.		172800	IN	NS	ns-613.awsdns-13.net.
k8s.example.com.		172800	IN	NS	ns-75.awsdns-04.org.
k8s.example.com.		172800	IN	NS	ns-1022.awsdns-35.com.
k8s.example.com.		172800	IN	NS	ns-1149.awsdns-27.co.uk.
```

Wait until the NS replication is ok

