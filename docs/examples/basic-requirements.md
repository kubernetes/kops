# Common Basic Requirements For Kops-Related Labs. Pre-Flight Check:

Before rushing-in to replicate any of the exercises, please ensure your basic environment is correctly set-up. See [KOPS AWS tutorial](../getting_started/aws.md) for more information.

Basic requirements:

- Configured AWS cli (aws account set-up with proper permissions/roles needed for kops). Depending on your distro, you can set-up directly from packages, or if you want the most updated version, use `pip` (python package manager) to install by running `pip install awscli` command from your local terminal. Your choice!
- Local ssh key ready on `~/.ssh/id_rsa` / `id_rsa.pub`. You can generate it using `ssh-keygen` command if you don't have one already: `ssh-keygen -t rsa -f ~/.ssh/id_rsa -P ""`.
- AWS Region set. 
  - Throughout most of the exercises, we'll deploy our clusters in us-east-1 region (AZs: us-east-1a, us-east-1b, us-east-1c, us-east-1d, us-east-1e and us-east-1f). 
  - For real HA at the Kubernetes API level, you need [3 masters](../operations/high_availability.md). 

Using `root` to set up the utilities for all users on that machine (either ssh directly to `root` or switch to is by running `sudo su -`):

```bash
cd ~
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
curl -LO https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
chmod 755 kubectl kops-linux-amd64
mv kops-linux-amd64 kops
mv kubectl kops  /usr/local/bin
```

Alternatively, if you don't have `root` access and/or wish to keep the `kops`/`kubectl` utilities in your local profile:

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

Finally, some of our exercises use the `jq` utility which is available on modern linux distributions. Please ensure to [install](https://stedolan.github.io/jq/download/) it as well.
