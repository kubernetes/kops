# Canal

[Canal](https://github.com/projectcalico/canal) is a project that combines [Flannel](https://github.com/coreos/flannel) and [Calico](http://docs.projectcalico.org/latest/getting-started/kubernetes/installation/) for CNI Networking.  It uses Flannel for networking pod traffic between hosts via VXLAN and Calico for network policy enforcement and pod to pod traffic.

## Installing

The following command sets up a cluster using Canal.

```sh
export ZONES=mylistofzones
kops create cluster \
  --zones $ZONES \
  --networking canal \
  --yes \
  --name myclustername.mydns.io
```

## Getting help

For problems with deploying Canal please post an issue to Github:

- [Canal Issues](https://github.com/projectcalico/canal/issues)

For support with Calico Policies you can reach out on Slack or Github:

- [Calico Github](https://github.com/projectcalico/calico)
- [Calico Users Slack](https://calicousers.slack.com)

For support with Flannel you can submit an issue on Github:

- [Flannel](https://github.com/coreos/flannel/issues)