# hget

hget is a tool for downloading files, verifying their integrity using a hash like sha256.

It has a few main goals:

* Scriptable: Make it easy to download files from a script, without worrying about whether curl or wget is available.
* Verifiable: Ensure that the file is downloaded correctly, and that the hash matches the expected value.
* Flexible: Abstracts away the source of the file, so that you can easily use mirrors and (in future) things like local caches.

# Usage

## Direct download

```bash
# Download kOps for linux/amd64 from github
hget --sha256=9253d15938376236d6578384e3d5ee0b973bdaf3303fb5fd6fbb3c59aedb9d8d --output=./kops --url=https://github.com/kubernetes/kops/releases/download/v1.31.0/kops-linux-amd64 --chmod=0755
```

## Use of index files (e.g. SHA256SUMS)

hget can use a SHA256SUMS file to find the file to download.

```bash
# Download kubectl for linux/amd64 from kubernetes v1.32.0
# 646d58f6d98ee670a71d9cdffbf6625aeea2849d567f214bc43a35f8ccb7bf70  bin/linux/amd64/kubectl
hget --sha256=646d58f6d98ee670a71d9cdffbf6625aeea2849d567f214bc43a35f8ccb7bf70 --chmod=0755 --output=./kubectl --index=https://dl.k8s.io/v1.32.0/SHA256SUMS
```

This will download the sha256sum file, locate the matching file and download it, verify the sha256 hash, and then set the permissions to 0755.

The sha256sum index file is not verified, but the file itself is verified.

# Installation

From source:

```bash
go install k8s.io/kops/tools/hget@latest
```
