## Security Notes for Kubernetes

## SSH Access

SSH is allowed to the masters and the nodes, by default from anywhere.

To change the CIDR allowed to access SSH (and HTTPS), set AdminAccess on the cluster spec.

When using the default images, the SSH username will be `admin`, and the SSH private key will be
the private key corresponding to the public key in `kops get secrets --type sshpublickey admin`.  When
creating a new cluster, the SSH public key can be specified with the `--ssh-public-key` option, and it
defaults to `~/.ssh/id_rsa.pub`.

> Note: In CoreOS, SSH username will be `core`.

To change the SSH public key on an existing cluster:

* `kops delete secret --name <clustername> sshpublickey admin`
* `kops create secret --name <clustername> sshpublickey admin -i ~/.ssh/newkey.pub`
* `kops update cluster --yes` to reconfigure the auto-scaling groups
* `kops rolling-update cluster --name <clustername> --yes` to immediately roll all the machines so they have the new key (optional)

## Docker Configuration

If you are using a private registry such as quay.io, you may be familiar with the inconvenience of managing the `imagePullSecrets` for each namespace. It can also be a pain to use [Kops Hooks](cluster_spec.md#hooks) with private images. To configure docker on all nodes with access to one or more private registries:

* `kops create secret --name <clustername> dockerconfig -f ~/.docker/config.json`
* `kops rolling-update cluster --name <clustername> --yes` to immediately roll all the machines so they have the new key (optional)

This stores the [config.json](https://docs.docker.com/engine/reference/commandline/login/) in `/root/.docker/config.json` on all nodes (include masters) so that both Kubernetes and system containers may use registries defined in it.

## IAM roles

All Pods running on your cluster have access to underlying instance IAM role.
Currently permission scope is quite broad. See [iam_roles.md](iam_roles.md) for details and ways to mitigate that.

## Kubernetes API

(this section is a work in progress)

Kubernetes has a number of authentication mechanisms:

## Kubelet API

By default AnonymousAuth on the kubelet is 'on' and so communication between kube-apiserver and kubelet api is not authenticated. In order to switch on authentication;

```YAML
# In the cluster spec
spec:
  kubelet:
    anonymousAuth: false
```

**Note** on an existing cluster with 'anonymousAuth' unset you would need to first roll out the masters and then update the node instance groups.

### API Bearer Token

The API bearer token is a secret named 'admin'.

`kops get secrets --type secret admin -oplaintext` will show it

### Admin Access

Access to the administrative API is stored in a secret named 'kube':

`kops get secrets kube -oplaintext` or `kubectl config view --minify` to reveal
