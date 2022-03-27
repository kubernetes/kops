## Security Notes for Kubernetes

## SSH Access

SSH is allowed to the masters and the nodes, by default from anywhere. However, no public key will be
installed. You can use instead use [ec2 instance connect](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Connect-using-EC2-Instance-Connect.html),
which is installed in the default AMIs.

If you want to use a fixed key for the cluster, you have to specify `--ssh-public-key <public key file>` on the `kops create cluster` command
or use `kops create sshpublickey`. You can also set the following in the cluster spec:

```yaml
spec:
  sshKeyName: <ssh key pair>
```

An EC2 key pair with the name`<ssh key pair>` has to already exist.

By default, SSH is allowed from any address. You can restrict from where SSH connections can be made by
setting either `spec.sshAccess` in the cluster spec or using `kops create cluster --ssh-access`.

To change the SSH public key on an existing cluster:

* `kops delete sshpublickey --name <clustername> sshpublickey`
* `kops create sshpublickey --name <clustername> sshpublickey -i ~/.ssh/newkey.pub`
* `kops update <clustername> --yes` to reconfigure the launch templates.
* `kops rolling-update cluster --name <clustername> --yes` to roll all the machines so they have the new key.

## Docker Configuration

If you are using a private registry such as quay.io, you may be familiar with the inconvenience of managing the `imagePullSecrets` for each namespace. It can also be a pain to use [kOps Hooks](cluster_spec.md#hooks) with private images. To configure docker on all nodes with access to one or more private registries:

* `kops create secret --name <clustername> dockerconfig -f ~/.docker/config.json`
* `kops rolling-update cluster --name <clustername> --yes` to immediately roll all the machines so they have the new key (optional)

This stores the [config.json](https://docs.docker.com/engine/reference/commandline/login/) in `/root/.docker/config.json` on all nodes (include masters) so that both Kubernetes and system containers may use registries defined in it.

Note that this will also work when using containerd.

## Instance IAM roles

All Pods running on your cluster have access to underlying instance IAM role.
Currently, permission scope is quite broad. See [iam_roles.md](iam_roles.md) for details and ways to mitigate that.