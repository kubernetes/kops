# ExternalDNS

ExternalDNS synchronizes exposed Kubernetes Services and Ingresses with DNS providers.

## What it does

Inspired by [Kubernetes DNS](https://github.com/kubernetes/dns), Kubernetes' cluster-internal DNS server, ExternalDNS makes Kubernetes resources discoverable via public DNS servers. Like KubeDNS, it retrieves a list of resources (Services, Ingresses, etc.) from the [Kubernetes API](https://kubernetes.io/docs/api/) to determine a desired list of DNS records. *Unlike* KubeDNS, however, it's not a DNS server itself, but merely configures other DNS providers accordingly—e.g. [AWS Route 53](https://aws.amazon.com/route53/) or [Google CloudDNS](https://cloud.google.com/dns/docs/).

In a broader sense, ExternalDNS allows you to control DNS records dynamically via Kubernetes resources in a DNS provider-agnostic way.

## Deploying to a Cluster

The following tutorials are provided:

* [AWS](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/aws.md)
* [Azure](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/azure.md)
* [Cloudflare](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/cloudflare.md)
* [DigitalOcean](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/digitalocean.md)
* Google Container Engine
	* [Using Google's Default Ingress Controller](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/gke.md)
	* [Using the Nginx Ingress Controller](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/tutorials/nginx-ingress.md)
* [FAQ](https://github.com/kubernetes-incubator/external-dns/blob/master/docs/faq.md)

## Github repository

Source code is managed under kubernetes-incubator at [external-dns](https://github.com/kubernetes-incubator/external-dns).