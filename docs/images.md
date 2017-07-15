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

`image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20160830`

You can find the name for an image using e.g. `aws ec2 describe-images --image-id ami-a3641cb4`

(Please note that ubuntu is currently undergoing validation testing with k8s - use at your own risk!)

If you are creating a new cluster you can use the `--image` flag when running `kops create cluster`,
which should be easier than editing your instance groups.

In addition, we support a few-well known aliases for the owner:

* `kope.io` => `383156758163`
* `redhat.com` => `309956199498`

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

## Ubuntu

Ubuntu is not the default platform, but is believed to be entirely functional.

Ubuntu 16.04 or later is required (we require systemd).

For example, to use Ubuntu 16.04, you could specify:

`image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20160830`

You can find the name for an image by first consulting [Ubuntu's image finder](https://cloud-images.ubuntu.com/locator/),
and then using e.g. `aws ec2 describe-images --image-id ami-a3641cb4`

## CentOS

CentOS7 support is still experimental, but should work.  Please report any issues.

The following steps are known:

* You must accept the agreement at http://aws.amazon.com/marketplace/pp?sku=aw0evgkw8e5c1q413zgy5pjce
* Specify the AMI by id (there are no tags): us-east-1: `ami-6d1c2007`
* You may find images from the [CentOS AWS page](https://wiki.centos.org/Cloud/AWS)
* You can also query by product-code: `aws ec2 describe-images --region=us-west-2 --filters Name=product-code,Values=aw0evgkw8e5c1q413zgy5pjce`

Be aware of the following limitations:

* CentOS 7.2 is the recommended minimum version
* CentOS7 AMIs are running an older kernel than we prefer to run elsewhere

## RHEL7

RHEL7 support is still experimental, but should work.  Please report any issues.

The following steps are known:

* Redhat AMIs can be found using `aws ec2 describe-images --region=us-east-1 --owner=309956199498 --filters Name=virtualization-type,Values=hvm`
* You can specify the name using the 'redhat.com` owner alias, for example `redhat.com/RHEL-7.2_HVM-20161025-x86_64-1-Hourly2-GP2`

Be aware of the following limitations:

* RHEL 7.2 is the recommended minimum version
* RHEL7 AMIs are running an older kernel than we prefer to run elsewhere

## CoreOS

CoreOS support is highly experimental.  Please report any issues.

The following steps are known:

* The latest stable CoreOS AMI can be found using:
```
aws ec2 describe-images --region=us-east-1 --owner=595879546273 \
    --filters "Name=virtualization-type,Values=hvm" "Name=name,Values=CoreOS-stable*" \
    --query 'sort_by(Images,&CreationDate)[-1].{id:ImageLocation}'
```

* You can specify the name using the `coreos.com` owner alias, for example `coreos.com/CoreOS-stable-1353.8.0-hvm`

> Note: SSH username will be `core`
