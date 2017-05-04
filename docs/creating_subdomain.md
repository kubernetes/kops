## Creating a Subdomain That Uses Amazon Route 53 as the DNS Service without Migrating the Parent Domain
<hr>

You can create a subdomain that uses Amazon Route 53 as the DNS service without migrating the parent domain from another DNS service.

example :

Your main domain is `example.com` but you want to create a subdomain NameServer.

Stat of your domain.

> dig ns example.com

``` 
;; QUESTION SECTION:
;example.com.			IN	NS
 
;; ANSWER SECTION:
example.com.		3600	IN	NS	ns3.somensserver.com.
```

## Create Subdomain
You want to keep those records, now lets create the subdomain.

On your `route 53` create the subdomain :

`Create Hosted zone`

Fill up the box `Domain Name:` with your subdomain : k8s.example.com

`Route 53` should generate your NS server like :

```
;; ANSWER SECTION:
ns-613.awsdns-13.net.
ns-75.awsdns-04.com.
ns-1022.awsdns-35.co.uk.
ns-1149.awsdns-27.org.
```

With your registrar add those NS server to your subdomain.

The result should be.

>dig ns k8s.example.com

```
;; ANSWER SECTION:
k8s.example.com.		172800	IN	NS	ns-613.awsdns-13.net.
k8s.example.com.		172800	IN	NS	ns-75.awsdns-04.org.
k8s.example.com.		172800	IN	NS	ns-1022.awsdns-35.com.
k8s.example.com.		172800	IN	NS	ns-1149.awsdns-27.co.uk.
```

Wait until the NS replication is ok

