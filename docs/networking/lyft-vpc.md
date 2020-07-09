# Lyft CNI

The [lyft cni-ipvlan-vpc-k8s](https://github.com/lyft/cni-ipvlan-vpc-k8s) plugin uses Amazon Elastic Network Interfaces (ENI) to assign AWS-managed IPs to Pods using the Linux kernel's IPvlan driver in L2 mode.

## Installing

Read the [prerequisites](https://github.com/lyft/cni-ipvlan-vpc-k8s#prerequisites) before starting. In addition to that, you need to specify the VPC ID as `spec.networkID` in the cluster spec file.

To use the Lyft CNI, specify the following in the cluster spec.

```yaml
  networking:
    lyftvpc: {}
```

in the cluster spec file or pass the `--networking lyftvpc` option on the command line to kops:

```console
$ export ZONES=mylistofzones
$ kops create cluster \
  --zones $ZONES \
  --master-zones $ZONES \
  --master-size m4.large \
  --node-size m4.large \
  --networking lyftvpc \
  --yes \
  --name myclustername.mydns.io
```

## Configuring

### Specify subnet selector

You can specify which subnets to use for allocating Pod IPs by specifying

```yaml
  networking:
    lyftvpc:
      subnetTags:
        KubernetesCluster: myclustername.mydns.io
```

In this example, new interfaces will be attached to subnets tagged with `KubernetesCluster = myclustername.mydns.io`.

### Change the download location

By default the plugin is downloaded from Github at node startup.  This location can be changed using environment variables

```bash
export LYFT_VPC_DOWNLOAD_URL="https://example.com/cni-ipvlan-vpc-k8s-amd64-v0.6.0.tar.gz"
export LYFT_VPC_DOWNLOAD_HASH="3aadcb32ffda53990153790203eb72898e55a985207aa5b4451357f9862286f0"
```

The hash can be MD5, SHA1 or SHA256.

## Troubleshooting

In case of any issues the directory `/var/log/aws-routed-eni` contains the log files of the CNI plugin. This directory is located in all the nodes in the cluster.
