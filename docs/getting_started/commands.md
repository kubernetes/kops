# Commands & Arguments
This page lists the most common kops commands.
Please refer to the kops [cli reference](../cli/kops.md) for full documentation.

## `kops create

`kops create` registers a cluster. There are two ways of registering a cluster: using a cluster spec file or using cli arguments.

### `kops create -f <cluser spec>`

`kops create -f <cluster spec>` will register a cluster using a kops spec yaml file. After the cluster has been registered you need to run `kops update cluster --yes` to create the cloud resources.

### `kops create cluster`

`kops create cluster <clustername>` creates a cloud specification in the registry using cli arguments. In most cases, you will need to edit the cluster spec using `kops edit` before actually creating the cloud resources. If you are sure you do not need to do any moditication, you can add the `--yes` flag to immediately create the cluster including cloud resource.

## `kops update cluster`

`kops update cluster <clustername>` creates or updates the cloud resources to match the cluster spec.

It is recommended that you run it first in 'preview' mode with `kops update cluster --name <name>`, and then
when you are happy that it is making the right changes you run`kops update cluster --name <name> --yes`.

## `kops rolling-update cluster`

`kops update cluster <clustername>` updates a kubernetes cluster to match the cloud and kops specifications.

It is recommended that you run it first in 'preview' mode with `kops rolling-update cluster --name <name>`, and then
when you are happy that it is making the right changes you run`kops rolling-update cluster --name <name> --yes`.

## `kops get clusters`

`kops get clusters` lists all clusters in the registry.

## `kops delete cluster`

`kops delete cluster` deletes the cloud resources (instances, DNS entries, volumes, ELBs, VPCs etc) for a particular
cluster.  It also removes the cluster from the registry.

It is recommended that you run it first in 'preview' mode with `kops delete cluster --name <name>`, and then
when you are happy that it is deleting the right things you run `kops delete cluster --name <name> --yes`.

## `kops toolbox template`

`kops toolbox template` lets you generate a kops spec using go templates. This is very handy if you want to consistently manage multiple clusters.

## `kops version`

`kops version` will print the version of the code you are running.
