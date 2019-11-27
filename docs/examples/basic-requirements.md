# COMMON BASIC REQUIREMENTS FOR KOPS-RELATED LABS. PRE-FLIGHT CHECK:

Before rushing in to replicate any of the exercises, please ensure your basic environment is correctly setup. See the [KOPS AWS tutorial for more information](../getting_started/aws.md).

Ensure that the following points are covered and working in your environment:

- AWS cli fully configured (aws account already with proper permissions/roles needed for kops). Depending on your distro, you can setup directly from packages, or if you want the most updated version, use "pip" and install awscli by issuing a "pip install awscli" command. Your choice!
- Local ssh key ready on ~/.ssh/id_rsa / id_rsa.pub. You can generate it using "ssh-keygen" command if you don't have one already: `ssh-keygen -t rsa -f ~/.ssh/id_rsa -P ""`.
- Region set to us-east-1 (az's: us-east-1a, us-east-1b, us-east-1c, us-east-1d and us-east-1e). For most of our exercises we'll deploy our clusters in "us-east-1". For real HA at kubernetes master level, you need 3 masters. If you want to ensure that each master is deployed on a different availability zone, then a region with "at least" 3 availability zones is required here. You can still deploy a multi-master kubernetes setup on regions with just 2 az's or even 1 az but this mean that two or all your masters will be deployed on a single az and if this az goes offline then you'll lose two or all your masters. If possible, always pick a region with at least 3 different availability zones for real H.A. You always can check amazon regions and az's on the link: [AWS Global Infrastructure](https://aws.amazon.com/about-aws/global-infrastructure/). Remember: The masters are Kubernetes control plane. If your masters die, you loose control of your Kubernetes cluster.
- kubectl and kops installed. For this last part, you can do this with using following commands. Next commands assume you are running a amd64/x86_64 linux distro:

As root (either ssh directly to root, local root console, or by using "sudo su -" previously):

```bash
cd ~
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
curl -LO https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
chmod 755 kubectl kops-linux-amd64
mv kops-linux-amd64 kops
mv kubectl kops  /usr/local/bin
```

If you are not root and/or do you want to keep the kops/kubectl utilities in your own account:

```bash
cd ~
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
curl -LO https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
chmod 755 kubectl kops-linux-amd64
mv kops-linux-amd64 kops
mkdir ~/bin
export PATH=$PATH:~/bin
mv kubectl kops  ~/bin
```

Finally, some of our exercises use the "jq" utility that is available on modern linux distributions. Please ensure to install it too. Some examples of how to do it:

**Centos 7:**

```bash
yum -y install epel-release
yum -y install jq
```

**Debian7/Debian8/Debian9/Ubuntu1404lts/Ubuntu1604lts:**

```bash
apt-get -y update
apt-get -y install jq
```

Also, if you are using **OS X** you can install jq using ["Homebrew"](https://brew.sh):

```bash
brew install jq
```

More information about "jq" on the following site: [https://stedolan.github.io/jq/download/](https://stedolan.github.io/jq/download/)

