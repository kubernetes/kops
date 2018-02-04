# Cluster Upgrader

`upgrader` is a standalone tool to upgrade an entire cluster, including the
control plane and all nodes. It is an example of a tool that can be written on
top of the Cluster API in a completely cloud-agnostic way.

## Building

```bash
$ cd $GOPATH/src/k8s.io/
$ git clone git@github.com:kubernetes/kube-deploy.git
$ cd kube-deploy/cluster-api/upgrader
$ go build
```

## Running
1) First, create a cluster using the `cluster-api` tool (the default Kubernetes version should be `1.7.4`)
2) To update the entire cluster to `v1.8.3`, run `./upgrader -v 1.8.3`
