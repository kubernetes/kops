# Development process and hacks for vSphere

This document contains few details, guidelines and tips about ongoing effort for vSphere support for kops.

## Contact
We are using [#sig-onprem channel](https://kubernetes.slack.com/messages/sig-onprem/) for discussing vSphere support for kops. Please feel free to join and talk to us.

## Process
Here is a [list of requirements and tasks](https://docs.google.com/document/d/10L7I98GuW7o7QuX_1QTouxC0t0aEO_68uHKNc7o4fXY/edit#heading=h.6wyer21z75n9 "Kops-vSphere specification") that we are working on. Once the basic infrastructure for vSphere is ready, we will move these tasks to issues.

## Setting up DNS
Since vSphere doesn't have built-in DNS service, we use CoreDNS to support the DNS requirement in vSphere provider. This requires the users to setup a CoreDNS server before creating a kubernetes cluster. Please follow the following instructions to setup.
Before the support of CoreDNS becomes stable, use env parameter "VSPHERE_DNS=coredns" to enable using CoreDNS. Or else AWS Route53 will be the default DNS service. To use Route53, follow instructions on: https://github.com/vmware/kops/blob/vsphere-develop/docs/aws.md

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

## Hacks

### Nodeup and protokube testing
This Section talks about testing nodeup and protokube changes on a standalone VM, running on standalone esx or vSphere.

#### Pre-requisites
Following manual steps are pre-requisites for this testing, until vSphere support for kops starts to create this infrastructure.

+ Setup password free ssh to the VM
```bash
cat ~/.ssh/id_rsa.pub | ssh <username>@<vm_ip> 'cat >> .ssh/authorized_keys'
```
+ Nodeup configuration file needs to be present on the VM. It can be copied from an existing AWS created master (or worker, whichever you are testing), from this location /var/cache/kubernetes-install/kube_env.yaml on your existing cluster node. Sample nodeup cofiguation file-
 ```yaml
Assets:
- 5e486d4a2700a3a61c4edfd97fb088984a7f734f@https://storage.googleapis.com/kubernetes-release/release/v1.5.2/bin/linux/amd64/kubelet
- 10e675883b167140f78ddf7ed92f936dca291647@https://storage.googleapis.com/kubernetes-release/release/v1.5.2/bin/linux/amd64/kubectl
- 19d49f7b2b99cd2493d5ae0ace896c64e289ccbb@https://storage.googleapis.com/kubernetes-release/network-plugins/cni-07a8a28637e97b22eb8dfe710eeae1344f69d16e.tar.gz
ClusterName: cluster3.mangoreviews.com
ConfigBase: s3://your-objectstore/cluster1.yourdomain.com
InstanceGroupName: master-us-west-2a
Tags:
- _automatic_upgrades
- _aws
- _cni_bridge
- _cni_host_local
- _cni_loopback
- _cni_ptp
- _kubernetes_master
- _kubernetes_pool
- _protokube
channels:
- s3://your-objectstore/cluster1.yourdomain.com/addons/bootstrap-channel.yaml
protokubeImage:
  hash: 6805cba0ea13805b2fa439914679a083be7ac959
  name: protokube:1.5.1
  source: https://kubeupv2.s3.amazonaws.com/kops/1.5.1/images/protokube.tar.gz

 ```
+ Currently vSphere code is using AWS S3 for storing all configurations, spec, etc. You need valid AWS credentials.
+ s3://your-objectstore/cluster1.yourdomain.com folder should have all necessary configuration, spec, addons, etc. (If you don't know how to get this, then read more on kops and how to deploy a cluster using kops)

#### Testing your changes
Once you are done making your changes in nodeup and protokube code, you would want to test them on a VM. In order to do so you will need to build nodeup binary and copy it on the desired VM. You would also want to modify nodeup code so that it accesses protokube container image that contains your changes. All this can be done by setting few environment variables, minor code updates and running 'make push-vsphere'.

 + Create or use existing docker hub registry to create 'protokube' repo for your custom image. Update the registry details in Makefile, by modifying DOCKER_REGISTRY variable. Don't forget to do 'docker login' with your registry credentials once.
 + Export TARGET environment variable, setting its value to username@vm_ip of your test VM.
 + Update $KOPS_DIR/upup/models/nodeup/_protokube/services/protokube.service.template-
 ```
 ExecStart=/usr/bin/docker run -v /:/rootfs/ -v /var/run/dbus:/var/run/dbus -v /run/systemd:/run/systemd --net=host --privileged -e AWS_ACCESS_KEY_ID='something' -e AWS_SECRET_ACCESS_KEY='something'  <your-registry>/protokube:<image-tag> /usr/bin/protokube "$DAEMON_ARGS"
 ```
+ Run 'make push-vsphere'. This will build nodeup binary, scp it to your test VM, build protokube image and upload it to your registry.
+ SSH to your test VM and set following environment variables-
  ```bash
  export AWS_REGION=us-west-2
  export AWS_ACCESS_KEY_ID=something
  export AWS_SECRET_ACCESS_KEY=something
  ```
+ Run './nodeup --conf kube_env.yaml' to test your custom build nodeup and protokube.

**Tip:** Consider adding following code to $KOPS_DIR/upup/pkg/fi/nodeup/nodetasks/load_image.go to avoid downloading protokube image. Your custom image will be downloaded directly when systemd will run protokube.service (because of the changes we made in protokube.service.template).
 ```go
 	// Add this after url variable has been populated.
 	if strings.Contains(url, "protokube") {
 		fmt.Println("Skipping protokube image download and loading.")
 		return nil
 	}
 ```


 **Note:** Same testing can also be done using alternate steps (these steps are _not working_ currently due to hash match failure):
  + Run 'make protokube-export' and 'make nodeup' to build and export protokube image as tar.gz, and to build nodeup binary. Both located in $KOPS_DIR/.build/dist/images/protokube.tar.gz and $KOPS_DIR/.build/dist/nodeup, respectively.
  + Copy nodeup binary to the test VM.
  + Upload $KOPS_DIR/.build/dist/images/protokube.tar.gz and $KOPS_DIR/.build/dist/images/protokube.tar.gz.sha1, with appropriate permissions, to a location from where it can be accessed from the test VM. Eg: your development machine's public_html, if working on linux based machine.
  + Update hash value to protokube.tar.gz.sha1's value and source to the uploaded location, in kube_env.yaml (see pre-requisite steps).
  + SSH to your test VM, set necessary environment variables and run './nodeup --conf kube_env.yaml'.
