## Security Notes for Kubernetes

## SSH Access

SSH is allowed to the masters and the nodes, by default from anywhere.

To change the CIDR allowed to access SSH (and HTTPS), set AdminAccess on the cluster spec.

When using the default images, the SSH username will be `admin`, and the SSH private key is be
the private key corresponding to the public key in `kops get secrets --type sshpublickey admin`.  When
creating a new cluster, the SSH public key can be specified with the `--ssh-public-key` option, and it
defaults to `~/.ssh/id_rsa.pub`.

> Note: In CoreOS, SSH username will be `core`.

To change the SSH public key on an existing cluster:

* `kops delete secret --name <clustername> sshpublickey admin`
* `kops create secret --name <clustername> sshpublickey admin -i ~/.ssh/newkey.pub`
* `kops update cluster --yes` to reconfigure the auto-scaling groups
* `kops rolling-update cluster --name <clustername> --yes` to immediately roll all the machines so they have the new key (optional)

## IAM roles

All Pods running on your cluster have access to underlying instance IAM role.
Currently permission scope is quite broad. See [iam_roles.md](iam_roles.md) for details and ways to mitigate that.

## authRole ALPHA SUPPORT

This configuration allows a cluster to utilize existing auth roles.  Currently this configuration only supports aws.  
In order to use this feature you have to have to have the arn of a pre-existing role, and use the kops feature flag by setting
`export KOPS_FEATURE_FLAGS=+CustomRoleSupport`.  This feature is in ALPHA release only, and can cause very unusual behavior
with Kubernetes if use incorrectly.

AuthRole example:

```yaml
spec:
  authRole:
    master: arn:aws:iam::123417490108:role/kops-custom-master-role
    node: arn:aws:iam::123417490108:role/kops-custom-node-role
```

## Kubernetes API

(this section is a work in progress)

Kubernetes has a number of authentication mechanisms:

### API Bearer Token

The API bearer token is a secret named 'admin'.

`kops get secrets --type secret admin -oplaintext` will show it

### Admin Access

Access to the administrative API is stored in a secret named 'kube':

`kops get secrets kube -oplaintext` or `kubectl config view --minify` to reveal
