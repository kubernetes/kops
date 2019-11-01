# Commands & Arguments
Please refer to the kops [cli reference](../cli/kops.md) for full documentation.

## `kops create cluster`

`kops create cluster <clustername>` creates a cloud specification in the registry.  It will not create the cloud resources unless
you specify `--yes`, so that you have the chance to `kops edit` them.  (You will likely `kops update cluster` after
creating it).


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


## `kops version`

`kops version` will print the version of the code you are running.
