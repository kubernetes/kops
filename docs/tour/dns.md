# Setting up DNS with kops

kops relies on configuring Route53 (on AWS) for internal discovery and so you can discover the API server
from its name.

Although kops has some support for private hosted zones, this will document setting up with public DNS.

## First, pick your name

First, you need a real domain name, we'll use `example.com`.

We need to configure an AWS Route53 Hosted Zone for `example.com`, or a subdomain of `example.com` (like
`kubernetes.example.com`, or `dev.k8s.example.com`).  The advantage of choosing a subdomain is that you
won't disrupt the existing records (for example, the ones that control delivery of your mail).

Note that a hosted zone `k8s.example.com` can easily contain subdomains like `dev.k8s.example.com`,
or `mycluster.dev.k8s.example.com`.

Sometimes a big organization will want to delegate a subdomain to delegate control, for example `mygroup.example.com`
if they are willing to grant you authority over `mygroup.example.com`, but not `example.com`.

If in doubt, creating a subdomain is probably simpler.

## Creating the hostedzone & setting up your nameservers

Having chosen our hosted zone name, we need to create it and configure nameservers so that
it is part of the DNS system.

### Option 1: If you want to create a subdomain (`k8s.example.com`):

You must create a route53 hosted zone for `k8s.example.com`.  You can do that through the [AWS Route53 Console](https://console.aws.amazon.com/route53/home?region=us-east-1#),
or through the command line:

```
> aws route53 create-hosted-zone --name k8s.example.com --caller-reference 12345
{
    "HostedZone": {
        "ResourceRecordSetCount": 2, 
        "CallerReference": "12345", 
        "Config": {
            "PrivateZone": false
        }, 
        "Id": "/hostedzone/ZXK60VKORO09E", 
        "Name": "k8s.example.com."
    }, 
    "DelegationSet": {
        "NameServers": [
            "ns-1439.awsdns-51.org", 
            "ns-295.awsdns-36.com", 
            "ns-962.awsdns-56.net", 
            "ns-1601.awsdns-08.co.uk"
        ]
    }, 
    "Location": "https://route53.amazonaws.com/2013-04-01/hostedzone/ZXK60VKORO09E", 
    "ChangeInfo": {
        "Status": "PENDING", 
        "SubmittedAt": "2017-01-19T03:19:48.667Z", 
        "Id": "/change/C1V52EBAZT0IMX"
    }
}
```

Whether you create the zone through the console or directly, the important thing is that you have to configure
the 4 `NameServers` records on the parent domain.  That is because you need to tell the DNS system about your subdomain.
To do that, you need to register the subdomain (`k8s.example.com`) in the parent domain (`example.com`).  Then 
DNS clients will recursively resolve names, and follow them down into your subdomain.


In our case, this means creating NS records in example.com, like this:

```
k8s.example.com NS ns-1439.awsdns-51.org
k8s.example.com NS ns-295.awsdns-36.com
k8s.example.com NS ns-962.awsdns-56.net
k8s.example.com NS ns-1601.awsdns-08.co.uk
```

Note that every hostedzone gets 4 "random" names, and this is why we can create duplicate hosted zones
and don't have to prove our ownership of the domain - we prove ownership by configuring NS records.

You will need to do this wherever your domain name `example.com` has its DNS control panel; if in doubt it will
likely be the registrar where you bought your domain name (e.g. GoDaddy, NameCheap etc).  Note that you have to
do this even if `example.com` is also at Route53 - AWS does not automate this (but you could just use `example.com`
as your hosted zone)

Some recommended links:

* [Setting up a subdomain with Godaddy](http://blog.sefindustries.com/redirect-a-subdomain-to-route-53-from-godaddy/)


This might take a few minutes to apply, but then if you `dig NS k8s.example.com`, you should also see your nameservers:

```
> dig NS k8s.example.com

...

;; ANSWER SECTION:
k8s.example.com.            86399   IN      NS      ns-1439.awsdns-51.org.
k8s.example.com.            86399   IN      NS      ns-295.awsdns-36.com.
k8s.example.com.            86399   IN      NS      ns-962.awsdns-56.net.
k8s.example.com.            86399   IN      NS      ns-1601.awsdns-08.co.uk.

...

```


## Option 2: Moving the whole domain (`example.com`):


You must create a route53 hosted zone for `example.com`.  You can do that through the [AWS Route53 Console](https://console.aws.amazon.com/route53/home?region=us-east-1#),
or through the command line:

```
> aws route53 create-hosted-zone --name example.com --caller-reference 12345
{
    "HostedZone": {
        "ResourceRecordSetCount": 2, 
        "CallerReference": "12345", 
        "Config": {
            "PrivateZone": false
        }, 
        "Id": "/hostedzone/ZXK60VKORO09E", 
        "Name": "example.com."
    }, 
    "DelegationSet": {
        "NameServers": [
            "ns-1439.awsdns-51.org", 
            "ns-295.awsdns-36.com", 
            "ns-962.awsdns-56.net", 
            "ns-1601.awsdns-08.co.uk"
        ]
    }, 
    "Location": "https://route53.amazonaws.com/2013-04-01/hostedzone/ZXK60VKORO09E", 
    "ChangeInfo": {
        "Status": "PENDING", 
        "SubmittedAt": "2017-01-19T03:19:48.667Z", 
        "Id": "/change/C1V52EBAZT0IMX"
    }
}
```

You should now copy any important DNS records from your existing hosting, for example if you have
[GMail DNS records](https://support.google.com/quickfixes/answer/6252374?hl=en), you should copy those before repointing
the domain to route53.

Whether you create the zone through the console or directly, you now must configure the NS records with your
registrar, so that `example.com` will be served by route53.  The registrar is normally the place where you
bought your domain name.  That is because you need to tell the DNS system about your route53 hostedzone, and link
it to your domain.  To do that, you configure the nameservers for your domain (`example.com`) at your registrar.
That actually sets up NS records for you automatically under the `.com` nameservers.  Then DNS clients will
recursively resolve names, and follow down into your subdomain.

In our case, this means configuring NS records for example.com, like this:

```
ns-1439.awsdns-51.org
ns-295.awsdns-36.com
ns-962.awsdns-56.net
ns-1601.awsdns-08.co.uk
```

Note that every hostedzone gets 4 "random" names, and this is why we can create duplicate hosted zones
and don't have to prove our ownership of the domain - we prove ownership by configuring NS records.

For GoDadddy:

* Login to your Godaddy account
* Select your domain
* Go to Nameservers
* Click "Set Nameservers"
* paste the 4 nameserver names above.

For NameCheap:

* Login to your NameCheap account
* Select "Domain List" and then click "Manage" on your domain
* In the "Nameservers" section, choose "Custom DNS", and add all 4 records (you will probably have to click "Add Nameserver")
* Click the green "tick" mark to save



It can take a little while for this change to be configured (a few minutes).  But then if you run whois, you should
see your nameservers.

```
> whois example.com
...
Name Server: ns-1439.awsdns-51.org.
Name Server: ns-295.awsdns-36.com.
Name Server: ns-962.awsdns-56.net.
Name Server: ns-1601.awsdns-08.co.uk.
...
```


And if you `dig NS example.com`, you should also see your nameservers:

```
> dig NS example.com

...

;; ANSWER SECTION:
example.com.            86399   IN      NS      ns-1439.awsdns-51.org.
example.com.            86399   IN      NS      ns-295.awsdns-36.com.
example.com.            86399   IN      NS      ns-962.awsdns-56.net.
example.com.            86399   IN      NS      ns-1601.awsdns-08.co.uk.

...

```


Next step: [Create an S3 bucket](create_an_s3_bucket.md)
