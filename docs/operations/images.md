# Images

As of Kubernetes 1.18 the default images used by kOps are the **[official Ubuntu 20.04](#ubuntu-2004-focal)** images.

You can choose a different image for an instance group by editing it with `kops edit ig nodes`. You should see an `image` field in one of the following formats:

* `ami-abcdef` - specifies an AMI by id directly
* `<owner>/<name>` specifies an AMI by its owner's account ID  and name properties
* `<alias>/<name>` specifies an AMI by its [owner's alias](#owner-aliases) and name properties

Using the AMI id is precise, but ids vary by region. It is often more convenient to use the `<owner/alias>/<name>` if equivalent images with the same name have been copied to other regions. 

```yaml
image: ami-00579fbb15b954340
image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-20200423
image: ubuntu/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-20200423
```

You can find the name for an image using:

`aws ec2 describe-images --region us-east-1 --image-id ami-00579fbb15b954340`

## Security Updates

Automated security updates are handled by kOps for Debian, Flatcar and Ubuntu distros. This can be disabled by editing the cluster configuration:

```yaml
spec:
  updatePolicy: external
```

## Distros Support Matrix

The following table provides the support status for various distros with regards to kOps version: 

| Distro                              | Experimental | Stable | Deprecated | Removed | 
|-------------------------------------|-------------:|-------:|-----------:|--------:|
| [Amazon Linux 2](#amazon-linux-2)   |         1.10 |   1.18 |          - |       - |
| [CentOS 7](#centos-7)               |            - |    1.5 |       1.21 |       - |
| [CentOS 8](#centos-8)               |         1.15 |      - |       1.21 |       - |
| CoreOS                              |          1.6 |    1.9 |       1.17 |    1.18 |
| Debian 8                            |            - |    1.5 |       1.17 |    1.18 |
| [Debian 9](#debian-9-stretch)       |          1.8 |   1.10 |       1.21 |       - |
| [Debian 10](#debian-10-buster)      |         1.13 |   1.17 |          - |       - |
| [Debian 11](#debian-11-bullseye)    |       1.21.1 |      - |          - |       - |
| [Flatcar](#flatcar)                 |       1.15.1 |   1.17 |          - |       - |
| [Kope.io](#kopeio)                  |            - |      - |       1.18 |       - |
| [RHEL 7](#rhel-7)                   |            - |    1.5 |       1.21 |       - |
| [RHEL 8](#rhel-8)                   |         1.15 |   1.18 |          - |       - |
| [RHEL 8](#rhel-8)                   |         1.15 |   1.18 |          - |       - |
| Ubuntu 16.04                        |          1.5 |   1.10 |       1.17 |    1.20 |
| [Ubuntu 18.04](#ubuntu-1804-bionic) |         1.10 |   1.16 |          - |       - |
| [Ubuntu 20.04](#ubuntu-2004-focal)  |       1.16.2 |   1.18 |          - |       - |

## Supported Distros

### Amazon Linux 2

Amazon Linux 2 has variants using Kernel versions 4.14 and 5.10. Be sure to use the 5.10 images as specified in the image filter below. More information is available in the [AWS Documentation](https://aws.amazon.com/amazon-linux-2/faqs/).

For kOps versions 1.16 and 1.17, the only supported Docker version is `18.06.3`. Newer versions of Docker cannot be installed due to missing dependencies for `container-selinux`. This issue is fixed in kOps **1.18**.

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 137112412989 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=amzn2-ami-kernel-5.10-hvm-2*-x86_64-gp2"
```

### Debian 10 (Buster)

Debian 10 is based on Kernel version **4.19** which fixes some of the bugs present in Debian 9 and effects are less visible.

One notable change is the addition of `iptables` NFT, which is by default. This is not yet supported by most CNI plugins and seems to be [slower](https://youtu.be/KHMnC3kj3Js?t=771) than the legacy version. It is recommended to switch to `iptables` legacy by using the following script in `additionalUserData` for each instance group:

```yaml
additionalUserData:
  - name: busterfix.sh
    type: text/x-shellscript
    content: |
      #!/bin/sh
      update-alternatives --set iptables /usr/sbin/iptables-legacy
      update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
      update-alternatives --set arptables /usr/sbin/arptables-legacy
      update-alternatives --set ebtables /usr/sbin/ebtables-legacy
```

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 136693071363 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=debian-10-amd64-*"
```

### Debian 11 (Bullseye)

Debian 11 is based on Kernel version **5.10** which has no known major Kernel bugs and fully supports all Cilium features.

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 136693071363 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=debian-11-amd64-*"
```

### Flatcar

Flatcar is a friendly fork of CoreOS and as such, compatible with it.

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 075585003325 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=Flatcar-stable-*-hvm"
```

### RHEL 8

RHEL 8 is based on Kernel version **4.18** which fixes some of the bugs present in RHEL/CentOS 7 and effects are less visible.

One notable change is the addition of `iptables` NFT, which is the only iptables backend available. This may not be supported by some CNI plugins and should be used with care.

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 309956199498 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=RHEL-8.*x86_64*"
```

### Ubuntu 18.04 (Bionic)

Ubuntu 18.04.5 is based on Kernel version **5.4** which fixes all the known major Kernel bugs.
Earlier patch versions may still work, but are not recommended nor supported.

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 099720109477 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-*"
```

### Ubuntu 20.04 (Focal)

Ubuntu 20.04 is based on Kernel version **5.4** which fixes all the known major Kernel bugs.

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 099720109477 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-*"
```

## Deprecated Distros

### CentOS 7

CentOS 7 is based on Kernel version **3.10** which has a considerable number of known bugs that affect it and may be noticed in production clusters:

* [kubernetes/kubernetes#56903](https://github.com/kubernetes/kubernetes/issues/56903)
* [kubernetes/kubernetes#67577](https://github.com/kubernetes/kubernetes/issues/67577)

The minimum supported version is **7.4**. Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 125523088429 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=CentOS 7.*x86_64"
```

### CentOS 8

CentOS 8 has announced its End Of Life is December 31, 2021.

CentOS 8 is based on Kernel version **4.18** which fixes some of the bugs present in RHEL/CentOS 7 and effects are less visible.

One notable change is the addition of `iptables` NFT, which is the only iptables backend available. This may not be supported by some CNI plugins and should be used with care.

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 125523088429 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=CentOS 8.*x86_64"
```

### Debian 9 (Stretch)

Debian 9 is based on Kernel version **4.9** which has a number of known bugs that affect it and which may be noticed with larger clusters:

This release is **EOL**, which means that the Debian Security Team no longer handles security fixes. That is now the responsibility/purview of the LTS team, which is a group of volunteers who are paid by donations to Debian LTS.

* [kubernetes/kubernetes#56903](https://github.com/kubernetes/kubernetes/issues/56903)
* [kubernetes/kubernetes#67577](https://github.com/kubernetes/kubernetes/issues/67577)

Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 379101102735 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=debian-stretch-hvm-x86_64-gp2-*"
```

### Kope.io

Support for _kope.io_ images is deprecated. These images were the default until Kubernetes 1.18, when they were replaced by the [official Ubuntu 20.04](#ubuntu-2004-focal) images. 

The _kope.io_ images were based on [Debian 9 (Stretch)](#debian-9-stretch) and had all packages required by kOps pre-installed. Other than that, the changes to the official Debian images were [minimal](https://github.com/kubernetes-sigs/image-builder/blob/master/images/kube-deploy/imagebuilder/templates/1.18-stretch.yml#L174-L198).

### RHEL 7

RHEL 7 is based on Kernel version **3.10** which has a considerable number of known bugs that affect it and may be noticed in production clusters:

* [kubernetes/kubernetes#56903](https://github.com/kubernetes/kubernetes/issues/56903)
* [kubernetes/kubernetes#67577](https://github.com/kubernetes/kubernetes/issues/67577)

The minimum supported version is **7.4**. Available images can be listed using:

```bash
aws ec2 describe-images --region us-east-1 --output table \
  --owners 309956199498 \
  --query "sort_by(Images, &CreationDate)[*].[CreationDate,Name,ImageId]" \
  --filters "Name=name,Values=RHEL-7.*x86_64*"
```

## Owner aliases 

kOps supports owner aliases for the official accounts of supported distros:

* `kope.io` => `383156758163`
* `amazon` => `137112412989`
* `centos` => `125523088429`
* `debian9` => `379101102735`
* `debian10` => `136693071363`
* `debian11` => `136693071363`
* `flatcar` => `075585003325`
* `redhat` => `309956199498`
* `ubuntu` => `099720109477`
