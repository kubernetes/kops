# Development process and hacks for vSphere

This document contains few details, guidelines and tips about ongoing effort for vSphere support for kops.

## Contact
We are using [#sig-onprem channel](https://kubernetes.slack.com/messages/sig-onprem/) for discussing vSphere support for kops. Please feel free to join and talk to us.

## Process
Here is a [list of requirements and tasks](https://docs.google.com/document/d/10L7I98GuW7o7QuX_1QTouxC0t0aEO_68uHKNc7o4fXY/edit#heading=h.6wyer21z75n9 "Kops-vSphere specification") that we are working on. Once the basic infrastructure for vSphere is ready, we will move these tasks to issues.

## Setting up DNS
Since vSphere doesn't have built-in DNS service, we use CoreDNS to support the DNS requirement in vSphere provider. This requires the users to setup a CoreDNS server before creating a kubernetes cluster. Please follow the following instructions to setup.
**Before the support of CoreDNS becomes stable, use env parameter "VSPHERE_DNS=coredns"** to enable using CoreDNS. Or else AWS Route53 will be the default DNS service. To use Route53, follow instructions on: https://github.com/vmware/kops/blob/vsphere-develop/docs/aws.md

For now we hardcoded DNS zone to skydns.local. So your cluster name should have suffix skydns.local, for example: "mycluster.skydns.local"

### Setup CoreDNS server
1. Login to vSphere Client.
2. Right-Click on ESX host on which you want to deploy the DNS server.
3. Select Deploy OVF template.
4. Copy and paste URL for [OVA](https://storage.googleapis.com/kubernetes-anywhere-for-vsphere-cna-storage/coredns.ova).
5. Follow next steps according to instructions mentioned in wizard.
6. Power on the imported VM.
7. SSH into the VM and execute ./start-dns.sh under /root. Username/Password: root/kubernetes

### Check DNS server is ready
On your local machine, execute the following command:
```bash
dig @[DNS server's IP] -p 53 NS skydns.local
```

Successful answer should look like the following:
```bash
; <<>> DiG 9.8.3-P1 <<>> @10.162.17.161 -p 53 NS skydns.local
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 42011
;; flags: qr aa rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1

;; QUESTION SECTION:
;skydns.local.			IN	NS

;; ANSWER SECTION:
skydns.local.		160	IN	NS	ns1.ns.dns.skydns.local.

;; ADDITIONAL SECTION:
ns1.ns.dns.skydns.local. 160	IN	A	192.168.0.1

;; Query time: 74 msec
;; SERVER: 10.162.17.161#53(10.162.17.161)
;; WHEN: Tue Mar 14 22:40:06 2017
;; MSG SIZE  rcvd: 71
```

### Add DNS server information when create cluster
Add ```--dns=private --vsphere-coredns-server=http://[DNS server's IP]:2379``` into the ```kops create cluster``` command line.

### Use CoreDNS supported DNS Controller
Information about DNS Controller can be found [here](https://github.com/kubernetes/kops/blob/master/dns-controller/README.md)
Currently the DNS Controller is an add-on container and the image is from kope/dns-controller.
Before the vSphere support is officially merged into upstream, we need to set up CoreDNS supported DNS controller manually.
```bash
DOCKER_REGISTRY=[your docker hub repo] make dns-controller-push
export VSPHERE_DNSCONTROLLER_IMAGE=[your docker hub repo]
make
kops create cluster ...
```

## Kops with vSphere
vSphere cloud provider support in kops is a work in progress. To try out deploying kubernetes cluster on vSphere using kops, some extra steps are required.

### Pre-requisites
+ vSphere with at least one ESX, having sufficient free disk space on attached datastore. ESX VM's should have internet connectivity.
+ Setup DNS following steps given in relevant Section above.
+ Create the VM using this template (TBD).
+ Currently vSphere code is using AWS S3 for storing all configurations, specs, addon yamls, etc. You need valid AWS credentials to try out kops on vSphere. s3://your-objectstore/cluster1.skydns.local folder will have all necessary configuration, spec, addons, etc., required to configure kubernetes cluster. (If you don't know how to setup aws, then read more on kops and how to deploy a cluster using kops on aws)
+ Update ```[kops_dir]/hack/vsphere/set_env``` setting up necessary environment variables.

### Building
Execute following command(s) to build all necessary components required to run kops for vSphere-

```bash
source [kops_dir]/hack/vsphere/set_env
make vsphere-version-dist
```

Currently vSphere support is not part of any of the kops releases. Hence, all modified component- kops, nodeup, protokube, need building at least once. ```make vsphere-version-dist``` will do that and copy protokube image and nodeup binary at the target location specified by you in ```[kops_dir]/hack/vsphere/set_env```.

Please note that dns-controller has also been modified to support vSphere. You can continue to use ```export VSPHERE_DNSCONTROLLER_IMAGE=luomiao/dns-controller```. If you have made any local changes to dns-controller and would like to use your custom image you need to build the dns-controller image using ```DOCKER_REGISTRY=[your docker hub repo] make dns-controller-push``` and set ```VSPHERE_DNSCONTROLLER_IMAGE``` accordingly. Please see the relevant Section above, on setting up DNS.

### Launching Cluster
Execute following command to launch cluster.

```bash
.build/dist/darwin/amd64/kops create cluster kubernetes.skydns.local  --cloud=vsphere --zones=vmware-zone --dns-zone=skydns.local --networking=flannel
 --vsphere-server=10.160.97.44 --vsphere-datacenter=VSAN-DC --vsphere-resource-pool=VSAN-Cluster --vsphere-datastore=vsanDatastore --dns private --vsphere-coredns-server=http://10.192.217.24:2379 --image="ubuntu_16_04" 
```

Use .build/dist/linux/amd64/kops if working on a linux machine, instead of mac.

**Notes**

1. ```clustername``` should end with **skydns.local**. Example: ```kubernetes.cluster.skydns.local```.
2. For ```zones``` any string will do, for now. It's only getting used for the construction of names of various entities. But it's a mandatory argument.
3. Make sure following parameters have these values,
    * ```--dns-zone=skydns.local```
    * ```--networking=flannel```
    * ```--dns=private```

### Cleaning up environment
Run following command to cleanup all set environment variables and regenerate all images and binaries without any of the vSphere specific steps.

```bash
source [kops_dir]/hack/vsphere/cleanup_env
make version-dist
```

### Deleting cluster
Cluster deletion hasn't been fully implemented yet. So you will have to delete vSphere VM's manually for now.

Configuration and spec data can be removed from S3 using following command-
```bash
.build/dist/darwin/amd64/kops delete cluster yourcluster.skydns.local --yes
```