## `kops get clusters`

`kops get clusters` lists all clusters in the registry.

## `kops delete cluster`

`kops delete cluster` deletes the cloud resources (instances, DNS entries, volumes, ELBs, VPCs etc) for a particular
cluster.  It also removes the cluster from the registry.

It is recommended that you run it first in 'preview' mode with `kops delete cluster --name <name>`, and then
when you are happy that it is deleting the right things you do `kops delete cluster --name <name> --yes`.


## `kops version`

`kops version` will print the version of the code you are running.