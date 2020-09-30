# Commands & Arguments
This page lists the most common kops commands.
Please refer to the kops [cli reference](../cli/kops.md) for full documentation.

## `kops create`

`kops create` registers a cluster. There are two ways of registering a cluster: using a cluster spec file or using cli arguments.

### `kops create -f <cluster spec>`

`kops create -f <cluster spec>` will register a cluster using a kops spec yaml file. After the cluster has been registered you need to run `kops update cluster --yes` to create the cloud resources.

### `kops create cluster`

`kops create cluster <clustername>` creates a cloud specification in the registry using cli arguments. In most cases, you will need to edit the cluster spec using `kops edit` before actually creating the cloud resources. 
Once confirmed you don't need any modifications, you can add the `--yes` flag to immediately create the cluster including cloud resource.

## `kops update cluster`

`kops update cluster <clustername>` creates or updates the cloud resources to match the cluster spec.

As a precaution, it is safer run in 'preview' mode first using `kops update cluster --name <name>`, and once confirmed 
the output matches your expectations, you can apply the changes by adding `--yes` to the command - `kops update cluster --name <name> --yes`.

## `kops rolling-update cluster`

`kops update cluster <clustername>` updates a kubernetes cluster to match the cloud and kops specifications.

As a precaution, it is safer run in 'preview' mode first using `kops rolling-update cluster --name <name>`, and once confirmed 
the output matches your expectations, you can apply the changes by adding `--yes` to the command - `kops rolling-update cluster --name <name> --yes`.

## `kops get clusters`

`kops get clusters` lists all clusters in the registry.

## `kops delete cluster`

`kops delete cluster` deletes the cloud resources (instances, DNS entries, volumes, ELBs, VPCs etc) for a particular
cluster.  It also removes the cluster from the registry.

As a precaution, it is safer run in 'preview' mode first using `kops delete cluster --name <name>`, and once confirmed 
the output matches your expectations, you can perform the actual deletion by adding `--yes` to the command - `kops delete cluster --name <name> --yes`.

## `kops toolbox template`

`kops toolbox template` lets you generate a kops spec using `go` templates. This is very handy if you want to consistently manage multiple clusters.

## `kops version`

`kops version` will print the version of the code you are running.
