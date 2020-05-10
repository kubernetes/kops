# Lyft CNI

The [lyft cni-ipvlan-vpc-k8s](https://github.com/lyft/cni-ipvlan-vpc-k8s) plugin uses Amazon Elastic Network Interfaces (ENI) to assign AWS-managed IPs to Pods using the Linux kernel's IPvlan driver in L2 mode.

## Installing

Read the [prerequisites](https://github.com/lyft/cni-ipvlan-vpc-k8s#prerequisites) before starting. In addition to that, you need to specify the VPC ID as `spec.networkID` in the cluster spec file.

To use the Lyft CNI, specify the following in the cluster spec.

```
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

```
  networking:
    lyftvpc:
      subnetTags:
        KubernetesCluster: myclustername.mydns.io
```

In this example, new interfaces will be attached to subnets tagged with `kubernetes_kubelet = true`.

## Troubleshooting

In case of any issues the directory `/var/log/aws-routed-eni` contains the log files of the CNI plugin. This directory is located in all the nodes in the cluster.