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


## CentOS

CentOS7 support is still experimental.

The following steps are known:

* Accept the agreement at http://aws.amazon.com/marketplace/pp?sku=aw0evgkw8e5c1q413zgy5pjce
* Specify the AMI by id (there are no tags): us-east-1: ami-6d1c2007
* CentOS7 AMIs are running an older kernel than we prefer to run elsewhere

## RHEL7

RHEL7 support is still experimental.

The following steps are known:

* RHEL 7.2 is the recommended minimum version
* RHEL7 AMIs are running an older kernel than we prefer to run elsewhere
* Redhat AMIs can be found using `aws ec2 describe-images --region=us-east-1 --owner=309956199498 --filters Name=virtualization-type,Values=hvm`
* You can specify the name using the owner alias, for example `redhat.com/RHEL-7.2_HVM-20161025-x86_64-1-Hourly2-GP2`
