# High Availability (HA)
## Introduction
When we create kubernetes cluster using kops we are able to create a multiple K8S masters and by doing so we can assure that in a case of 
a master instance failure our K8S functionality won't damaged. 
In scenario of a single master instance that fails our minions/pods will keep running but there will be no new pods schedualing, kubectl
won't work (since the api-server will be down).


## Kops HA
We can create HA cluster using kops only in the first creation of the cluster by using the "--master-zones" flag
example: https://github.com/kubernetes/kops/blob/master/docs/advanced_create.md
K8S relay on a key value DB named "etcd", which is using the the Qurom concept that the cluster need at least 51% of nodes to be
available, for the cluster to work.
Kops currently doesn't support cross region HA

As a result there are few considerations that need to be taken into account when using kops with HA:
* Only odd number of masters instances can be created. 
* Currently Kops can't create more than 1 master in a single AZ.
* Kops can't create HA cluster on a region with 2 AZ. (There is no point to create 2 masters due to the fact the we need at least
51% of the nodes to be avilable so failure of one of the master will cause the whole HA to fail, thus running 2 masters only
increase the chance of the cluster failures).
