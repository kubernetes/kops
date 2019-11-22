# Development process and hacks for vSphere

This document contains details about ongoing effort for vSphere support in kops- how to use kops with vSphere, contact information and current status. vSphere support in kops is an experimental feature, under `KOPS_FEATURE_FLAGS=+VSphereCloudProvider` feature flag and is not production ready yet.

## Contact
We are using [#kops channel](https://kubernetes.slack.com/messages/C3QUFP0QM) for discussing vSphere support for kops. Please feel free to join and talk to us.

## Current status
Here is the [current status](vsphere-development-status.md) of vSphere support in kops.

## Setting up DNS
Since vSphere doesn't have built-in DNS service, we use CoreDNS to support the DNS requirement in vSphere provider. This requires the users to setup a CoreDNS server before creating a kubernetes cluster. Please follow the following instructions to setup.

For now we hardcoded DNS zone to skydns.local. So your cluster name should have suffix skydns.local, for example: "mycluster.skydns.local"

### Setup CoreDNS server
1. Login to vSphere Client.
2. Right-Click on ESX host on which you want to deploy the DNS server.
3. Select Deploy OVF template.
4. Copy and paste URL for [OVA](https://storage.googleapis.com/kops-vsphere/DNSStorage.ova) (uploaded 04/18/2017).
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
Information about DNS Controller can be found [here](https://github.com/kubernetes/kops/blob/master/dns-controller/README.md).
Currently the DNS Controller is an add-on container and the image is from kope/dns-controller.
Before the vSphere support is officially merged into upstream, please use the following CoreDNS supported DNS controller.
```bash
export DNSCONTROLLER_IMAGE=cnastorage/dns-controller
```
(The above environment variable is already set in [kops_dir]/hack/vsphere/set_env)

## Setting up cluster state storage
Kops requires the state of clusters to be stored inside certain storage service. AWS S3 is the default option.
More about using AWS S3 for cluster state store can be found at "Cluster State storage" on this [page](getting_started/aws.md).

Users can also setup their own S3 server and use the following instructions to use user-defined S3-compatible applications for cluster state storage.
This is recommended if you don't have AWS account or you don't want to store the status of your clusters on public cloud storage.

Minio is a S3-compatible object storage application. We have included Minio components inside the same OVA template for CoreDNS service.
If you haven't setup CoreDNS according to section "Setup CoreDNS server" of this document, please follow the instructions in section "Setup CoreDNS server" Step 1 to Step 6.

Then SSH into the VM for CoreDNS/Minio service and execute:
```bash
/root/start-minio.sh [bucket_name]
```

Output of the script should look like:
```bash
Please set the following environment variables into hack/vsphere/set_env accordingly, before using kops create cluster:
KOPS_STATE_STORE=s3://[s3_bucket]
S3_ACCESS_KEY_ID=[s3_access_key]
S3_SECRET_ACCESS_KEY=[s3_secret_key]
S3_REGION=[s3_region]
```

Update [kops_dir]hack/vsphere/set_env according to the output of the script and the IP address/service port of the Minio server:
```bash
export KOPS_STATE_STORE=s3://[s3_bucket]
export S3_ACCESS_KEY_ID=[s3_access_key]
export S3_SECRET_ACCESS_KEY=[s3_secret_key]
export S3_REGION=[s3_region]
export S3_ENDPOINT=http://[s3_server_ip]:9000
```

Users can also choose their own S3-compatible storage applications by setting environment variables similarly.

## Kops with vSphere
vSphere cloud provider support in kops is a work in progress. To try out deploying kubernetes cluster on vSphere using kops, some extra steps are required.

### Pre-requisites
+ vSphere with at least one ESX, having sufficient free disk space on attached datastore. ESX VM's should have internet connectivity.
+ Setup DNS and S3 storage service following steps given in relevant Section above.
+ Upload VM template. Steps:
1. Login to vSphere Client.
2. Right-Click on ESX host on which you want to deploy the template.
3. Select Deploy OVF template.
4. Copy and paste URL for [OVA](https://storage.googleapis.com/kops-vsphere/kops_ubuntu_16_04.ova) (uploaded 04/18/2017).
5. Follow next steps according to instructions mentioned in wizard.
**NOTE: DO NOT POWER ON THE IMPORTED TEMPLATE VM.**
+ Update ```[kops_dir]/hack/vsphere/set_env``` setting up necessary environment variables.
+ ```source [kops_dir]/hack/vsphere/set_env```

### Installing
Currently vSphere support is not part of upstream kops releases. Please use the following instructions to use binaries/images with vSphere support.

#### Linux
Download kops binary from [here](https://storage.googleapis.com/kops-vsphere/kops-linux-amd64), then:
```bash
chmod +x kops-linux-amd64                 # Add execution permissions
mv kops-linux-amd64 /usr/local/bin/kops   # Move the kops to /usr/local/bin
```

#### Darwin
Download kops binary from [here](https://storage.googleapis.com/kops-vsphere/kops-darwin-amd64), then:
```bash
chmod +x kops-darwin-amd64                 # Add execution permissions
mv kops-darwin-amd64 /usr/local/bin/kops   # Move the kops to /usr/local/bin
```

### Building from source
Execute following command(s) to build all necessary components required to run kops for vSphere:

```bash
source [kops_dir]/hack/vsphere/set_env
make vsphere-version-dist
```

```make vsphere-version-dist``` will build and upload protokube image and nodeup binary at the target location specified by you in ```[kops_dir]/hack/vsphere/set_env```.

Please note that dns-controller has also been modified to support vSphere. You can continue to use ```export DNSCONTROLLER_IMAGE=cnastorage/dns-controller```. If you have made any local changes to dns-controller and would like to use your custom image you need to build the dns-controller image using ```DOCKER_REGISTRY=[your docker hub repo] make dns-controller-push``` and set ```DNSCONTROLLER_IMAGE``` accordingly. Please see the relevant Section above, on setting up DNS.

### Launching Cluster
Execute following command to launch cluster.

```bash
kops create cluster kubernetes.skydns.local  --cloud=vsphere --zones=vmware-zone --dns-zone=skydns.local --networking=flannel
 --vsphere-server=10.160.97.44 --vsphere-datacenter=VSAN-DC --vsphere-resource-pool=VSAN-Cluster --vsphere-datastore=vsanDatastore --dns private --vsphere-coredns-server=http://10.192.217.24:2379 --image="kops_ubuntu_16_04.ova"
```

If kops doesn't exist in default path, locate it inside .build/dist/linux/amd64/kops for linux machine or .build/dist/darwin/amd64/kops for mac under kops source directory.

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
kops delete cluster yourcluster.skydns.local --yes
```
