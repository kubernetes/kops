# Machineset
`machineset` is an example client-side implementation of a MachineSet, which may
be included in the Cluster API in the future. It allows you to declaratively
scale a group of Machines identified by a label selector.

## Building

```bash
$ cd $GOPATH/src/k8s.io/
$ git clone git@github.com:kubernetes/kube-deploy.git
$ cd kube-deploy/cluster-api/examples/machineset
$ go build
```

## Running
1) Create a cluster using the `cluster-api` tool.
   - By default, the master and node Machines from `machines.yaml` have the
     labels `set=master` and `set=node`, respectively.
2) To print out the Machines in a set, run `./machineset get set=node`.
3) To scale the number of Machines up, run `./machineset scale set=node -r 3`.
4) To scale them back down, run `./machineset scale set=node -r 1`.
5) To see the full usage information, run `./machineset help`.
