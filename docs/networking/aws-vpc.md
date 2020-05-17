# Amazon VPC

The Amazon VPC CNI uses the native AWS networking for Pods. Every pod gets an Elastic Network Interface (ENI) on the node it is running and an IP address beloning to the subnets assigned to the node.

## Installing

To use Amazon VPC, specify the following in the cluster spec:

```yaml
  networking:
    amazonvpc: {}
```

in the cluster spec file or pass the `--networking amazon-vpc-routed-eni` option on the command line to kops:

```sh
export ZONES=<mylistofzones>
kops create cluster \
  --zones $ZONES \
  --networking amazon-vpc-routed-eni \
  --yes \
  --name myclustername.mydns.io
```

**Important:** pods use the VPC CIDR, i.e. there is no isolation between the master, node/s and the internal k8s network. In addition, this CNI does not enforce network policies.


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

## Troubleshooting

In case of any issues the directory `/var/log/aws-routed-eni` contains the log files of the CNI plugin. This directory is located in all the nodes in the cluster.