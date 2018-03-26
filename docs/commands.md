# Documentation

Please refer to the [cli](cli) directory for full documentation.

## `kops create cluster`

`kops create cluster <clustername>` creates a cloud specification in the registry.  It will not create the cloud resources unless
you specify `--yes`, so that you have the chance to `kops edit` them.  (You will likely `kops update cluster` after
creating it).


## `kops update cluster`

`kops update cluster <clustername>` creates or updates the cloud resources to match the cluster spec.

It is recommended that you run it first in 'preview' mode with `kops update cluster --name <name>`, and then
when you are happy that it is making the right changes you run`kops update cluster --name <name> --yes`.

## `kops get clusters`

`kops get clusters` lists all clusters in the registry.

## `kops delete cluster`

`kops delete cluster` deletes the cloud resources (instances, DNS entries, volumes, ELBs, VPCs etc) for a particular
cluster.  It also removes the cluster from the registry.

It is recommended that you run it first in 'preview' mode with `kops delete cluster --name <name>`, and then
when you are happy that it is deleting the right things you run `kops delete cluster --name <name> --yes`.


## `kops version`

`kops version` will print the version of the code you are running.

## Other interesting modes:

* Build a terraform model: `--target=terraform`  The terraform model will be built in `out/terraform`

* Build a Cloudformation model: `--target=cloudformation`  The Cloudformation json file will be built in 'out/cloudformation'

* Specify the k8s build to run: `--kubernetes-version=1.2.2`

* Run nodes in multiple zones: `--zones=us-east-1b,us-east-1c,us-east-1d`

* Run with a HA master: `--master-zones=us-east-1b,us-east-1c,us-east-1d`

* Specify the number of nodes: `--node-count=4`

* Specify the node size: `--node-size=m4.large`

* Specify the master size: `--master-size=m4.large`

* Override the default DNS zone: `--dns-zone=<my.hosted.zone>`

