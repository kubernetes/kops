# Romana

## Installing

To use Romana, specify the following in the cluster spec:

```yaml
  networking:
    romana: {}
```

The following command sets up a cluster with Romana as the CNI.

```sh
export ZONES=mylistofzones
kops create cluster \
  --zones $ZONES \
  --networking romana \
  --yes \
  --name myclustername.mydns.io
```

Romana uses the cluster's etcd as a backend for storing information about routes, hosts, host-groups and IP allocations.
This does not affect normal etcd operations or require special treatment when upgrading etcd.
The etcd port (4001) is opened between masters and nodes when using this networking option.

## Getting help

For problems with deploying Romana please post an issue to Github:

- [Romana Issues](https://github.com/romana/romana/issues)

You can also contact the Romana team on Slack

- [Romana Slack](https://romana.slack.com) (invite required - email [info@romana.io](mailto:info@romana.io))