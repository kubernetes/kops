# Images

Changing the image for an instance group

You can choose a different AMI for an instance group.

If you `kops edit ig nodes`, you should see an `image` member of the spec.

Various syntaxes are available:

* `ami-abcdef` specifies an AMI by id directly.
* `<owner>/<name>` specifies an AMI by its owner and Name properties

The ami spec is precise, but AMIs vary by region.  So it is often more convenient to use the `<owner>/<name>`
specifier, if equivalent images have been copied to various regions with the same name.

For example, to use Ubuntu 16.04, you could specify:

`image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20180405`

You can find the name for an image using e.g. `aws ec2 describe-images --image-id ami-493f2f29`

(Please note that ubuntu is currently undergoing validation testing with k8s - use at your own risk!)

If you are creating a new cluster you can use the `--image` flag when running `kops create cluster`,
which should be easier than editing your instance groups.

In addition, we support a few-well known aliases for the owner:

* `kope.io` => `383156758163`
* `redhat.com` => `309956199498`
* `coreos.com` => `595879546273`
* `amazon.com` => `137112412989`

## Debian

A Debian image with a custom kubernetes kernel is the primary (default) platform for kops.

We run a Debian Jessie image, with a 4.4 (stable series) kernel that is built with kubernetes-specific settings.

The tooling used to build these images is open source:

* [imagebuilder](https://github.com/kubernetes/kube-deploy/tree/master/imagebuilder) is used to build an image
  as defined by a bootstrap-vz [template](https://github.com/kubernetes/kube-deploy/tree/master/imagebuilder/templates)
* The [kubernetes-kernel](https://github.com/kopeio/kubernetes-kernel) project has the build scripts / configuration
  used for building the kernel.

The latest image name is kept in the [stable channel manifest](https://github.com/kubernetes/kops/blob/master/channels/stable),
but an example is `kope.io/k8s-1.4-debian-jessie-amd64-hvm-ebs-2016-10-21`.  This means to look for an image published
by `kope.io`, (which is a well-known alias to account `383156758163`), with the name
`k8s-1.4-debian-jessie-amd64-hvm-ebs-2016-10-21`.  By using a name instead of an AMI, we can reference an image
irrespective of the region in which it is located.

kops should also now work on stock Debian 9 (Stretch) images.  Stock Debian 8 (Jessie) images are not recommended,
as they typically do not have a suitable kernel and kernel options configured.

## Ubuntu

Ubuntu is not the default platform, but is believed to be entirely functional.

Ubuntu 16.04 or later is required (we require systemd).

For example, to use Ubuntu 16.04, you could specify:

`image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20180405`

You can find the name for an image by first consulting [Ubuntu's image finder](https://cloud-images.ubuntu.com/locator/),
and then using e.g. `aws ec2 describe-images --image-id ami-493f2f29`

## CentOS

CentOS7 support is still experimental, but should work. Please report any issues.

The following steps are known:

* You must accept the agreement at http://aws.amazon.com/marketplace/pp?sku=aw0evgkw8e5c1q413zgy5pjce
* Specify the AMI by id (there are no tags): us-east-1: `ami-01ed306a12b7d1c96`
* You may find public images from the [CentOS AWS page](https://wiki.centos.org/Cloud/AWS) but the table may not be up to date.
* You can also query by product-code: `aws ec2 describe-images --region=us-west-2 --filters Name=product-code,Values=aw0evgkw8e5c1q413zgy5pjce Name=architecture,Values=x86_64 'Name=name,Values=CentOS*' --query 'sort_by(Images,&Name)'`

Be aware of the following limitations:

* CentOS 7.2 is the recommended minimum version
* CentOS7 AMIs are running an older kernel than we prefer to run elsewhere


## RHEL7

RHEL7 support is still experimental, but should work. Please report any issues.

The following steps are known:

* Redhat AMIs can be found using `aws ec2 describe-images --region=us-east-1 --owner=309956199498 --filters Name=virtualization-type,Values=hvm 'Name=name,Values=RHEL-*GA*' Name=architecture,Values=x86_64 --query 'sort_by(Images,&Name)'`
* You can specify the name using the `redhat.com` owner alias, for example `redhat.com/RHEL-7.2_HVM-20161025-x86_64-1-Hourly2-GP2`

Be aware of the following limitations:

* RHEL 7.2 is the recommended minimum version
* RHEL7 AMIs are running an older kernel than we prefer to run elsewhere

## CoreOS

CoreOS has been tested enough to be considered ready for production with kops, but if you encounter any problem please report it to us.

The following steps are known:

* The latest stable CoreOS AMI can be found using:
```bash
aws ec2 describe-images --region=us-east-1 --owner=595879546273 \
    --filters "Name=virtualization-type,Values=hvm" "Name=name,Values=CoreOS-stable*" \
    --query 'sort_by(Images,&CreationDate)[-1].{id:ImageLocation}'
```

Also, you can obtain the "AMI ID" from CoreOS web page too. They publish their AMI's using a json file at [https://coreos.com/dist/aws/aws-stable.json](https://coreos.com/dist/aws/aws-stable.json). Using some scripting and a "json" parser (like jq) you can obtain the AMI ID from a specific availability zone:

```bash
curl -s https://coreos.com/dist/aws/aws-stable.json | jq -r '.["us-east-1"].hvm'
"ami-32705b49"
```

* You can specify the name using the `coreos.com` owner alias, for example `coreos.com/CoreOS-stable-1409.8.0-hvm` or leave it at `595879546273/CoreOS-stable-1409.8.0-hvm` if you prefer to do so.

As part of our documentation, you will find a practical exercise using CoreOS with KOPS. See the file ["coreos-kops-tests-multimaster.md"](../examples/coreos-kops-tests-multimaster.md) in the "examples" directory. This exercise covers not only using kops with CoreOS, but also a practical view of KOPS with a multi-master kubernetes setup.

> Note: SSH username for CoreOS based instances will be `core`

## Amazon Linux 2

Amazon Linux 2 support is still experimental, but should work. Please report any issues.

The following steps are known:

* The latest Amazon Linux 2 AMI can be found using:
```bash
aws ec2 describe-images --region=us-east-1 --owner=137112412989 \
    --filters "Name=name,Values=amzn2-ami-hvm-2*-gp2" \
    --query 'sort_by(Images,&CreationDate)[-1].{name:Name}'
```
* You can specify the name using the `amazon.com` owner alias, for example `amazon.com/amzn2-ami-hvm-2.0.20180622.1-x86_64-gp2`

Be aware of the following limitations:

* [Amazon Linux 2 LTS](https://aws.amazon.com/amazon-linux-2/release-notes/) is the recommended minimum version, a previous version called just "Amazon Linux AMI" is not supported.

> Note: SSH username for Amazon Linux 2 based instances will be `ec2-user`

## Flatcar

Flatcar is a friendly fork of CoreOS and as such, compatible with it. If some issues occurs with it, it is likely that also CoreOS miight be affected. If you encounter any problem please report it to us.

The following steps are known:

* The latest stable Flatcar AMI can be found using:
```bash
aws ec2 describe-images --region=us-east-1 --owner=075585003325 \
    --filters "Name=virtualization-type,Values=hvm" "Name=name,Values=Flatcar-stable*" \
    --query 'sort_by(Images,&CreationDate)[-1].{id:ImageLocation}'
```

Also, you can obtain the "AMI ID" from Flatcar web page too. They publish their AMI's using a json file at [https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_ami_all.json](https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_ami_all.json). Using some scripting and a "json" parser (like jq) you can obtain the AMI ID from a specific region:

```bash
curl -s https://stable.release.flatcar-linux.net/amd64-usr/current/flatcar_production_ami_all.json | jq -r '.amis[] | select(.name == "us-east-1") | .hvm'
"ami-096be41989ec7e569"
```

> Note: SSH username for Flatcar based instances will be `core`
