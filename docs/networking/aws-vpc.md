# Amazon VPC

The Amazon VPC CNI uses the native AWS networking for Pods. Every pod gets an Elastic Network Interface (ENI) on the node it is running and an IP address belonging to the subnets assigned to the node.

**WARNING**: The Amazon VPC CNI is not compatible with Ubuntu 22.04 and kOps versions earlier than 1.29. See [kubernetes/kops#15720](https://github.com/kubernetes/kops/issues/15720) and [aws/amazon-vpc-cni-k8s#2103](https://github.com/aws/amazon-vpc-cni-k8s/issues/2103) for more info (Fix was applied via [kubernetes/kops#16313](https://github.com/kubernetes/kops/issues/16313) in kOps 1.29).

## Installing

To use Amazon VPC, specify the following in the cluster spec:

```yaml
  networking:
    amazonvpc: {}
```

in the cluster spec file or pass the `--networking amazonvpc` option on the command line to kOps:

```sh
export ZONES=<mylistofzones>
kops create cluster \
  --zones $ZONES \
  --networking amazonvpc \
  --yes \
  --name myclustername.mydns.io
```

**Important:** pods use the VPC CIDR, i.e. there is no isolation between the master, node/s and the internal k8s network. Network policies are not enforced unless enabled as described below.

## Configuration

[Configuration options for the Amazon VPC CNI plugin](https://github.com/aws/amazon-vpc-cni-k8s/tree/master#cni-configuration-variables) can be set through env vars defined in the cluster spec:

```yaml
  networking:
    amazonvpc:
      env:
      - name: WARM_IP_TARGET
        value: "10"
      - name: AWS_VPC_K8S_CNI_LOGLEVEL
        value: debug
```

## Network policies
{{ kops_feature_table(kops_added_default='1.38') }}

Network policy enforcement is disabled by default. To enable it, set:

```yaml
  networking:
    amazonvpc:
      enableNetworkPolicy: true
```

This enables the network policy agent in the `aws-node` DaemonSet and deploys the
[Amazon network policy controller](https://github.com/aws/amazon-network-policy-controller-k8s) on the control plane nodes.
The controller creates the `networking.k8s.aws` CRDs it needs.
Setting `enableNetworkPolicy` back to `false` removes the controller.
The node agent uses eBPF and requires a Linux kernel version 5.10 or later.
Override the controller image with `networkPolicyControllerImage`.

## Troubleshooting

In case of any issues the directory `/var/log/aws-routed-eni` contains the log files of the CNI plugin. This directory is located in all the nodes in the cluster.
