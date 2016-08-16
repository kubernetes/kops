# Upgrading from kubernetes 1.2 to kubernetes 1.3

Kops let you upgrade an existing 1.2 cluster, installed using kube-up, to a cluster managed by
kops running kubernetes version 1.3.

** This is an experimental and slightly risky procedure, so we recommend backing up important data before proceeding. 
Take a snapshot of your EBS volumes; export all your data from kubectl etc. **

Limitations:

* kops splits etcd onto two volumes now: `main` and `events`.  We will keep the `main` data, but
  you will lose your events history.
* Doubtless others not yet known - please open issues if you encounter them!

## Overview

There are a few steps to upgrade a kubernetes cluster from 1.2 to 1.3:

* First you import the existing cluster state, so you can see and edit the configuration
* You verify the cluster configuration
* You move existing AWS resources to your new cluster
* You bring up the new cluster
* You can then delete the old cluster

## Importing the existing cluster

The `import cluster` command reverse engineers an existing cluster, and creates a cluster
configuration.

Make sure you have set `export KOPS_STATE_STORE=s3://<mybucket>`

Then import the cluster; setting `--name` and `--region` to match the old cluster.   If you're not sure
of the old cluster name, you can find it by looking at the `KubernetesCluster` tag on your AWS resources.

```
export OLD_NAME=kubernetes
export REGION=us-west-2
kops import cluster --region ${REGION} --name ${OLD_NAME}
```

## Verify the cluster configuration

Now have a look at the cluster configuration, to make sure it looks right.  If it doesn't, please
open an issue.

```
kops get cluster ${OLD_NAME} -oyaml
````

## Move resources to a new cluster

The upgrade moves some resources so they will be adopted by the new cluster.  There are a number of things
this step does:

* It resizes existing autoscaling groups to size 0
* It will stop the existing master
* It detaches the master EBS volume from the master
* It re-tags resources to associate them with the new cluster: volumes, ELBs
* It re-tags the VPC to associate it with the new cluster

The upgrade procedure forces you to choose a new cluster name (e.g. `k8s.mydomain.com`)

```
export NEW_NAME=k8s.mydomain.com
kops toolbox convert-imported --newname ${NEW_NAME} --name ${OLD_NAME}
```

If you now list the clusters, you should see both the old cluster & the new cluster

```
kops get clusters
```

You can also list the instance groups: `kops get ig --name ${NEW_NAME}`

## Import the SSH public key

The SSH public key is not easily retrieved from the old cluster, so you must add it:

```
kops create secret --name ${NEW_NAME} sshpublickey admin -i ~/.ssh/id_rsa.pub
```

## Bring up the new cluster

Use the update command to bring up the new cluster:

```
kops update cluster ${NEW_NAME}
```

Things to check are that it is reusing the existing volume for the _main_ etcd cluster (but not the events clusters).

And then when you are happy:

```
kops update cluster ${NEW_NAME} --yes
```


## Export kubecfg settings to access the new cluster

You can export a kubecfg (although update cluster did this automatically): `kops export kubecfg ${NEW_NAME}`

Within a few minutes the new cluster should be running. 

Try `kubectl get nodes --show-labels`, `kubectl get pods` etc until you are sure that all is well.

## Workaround for secret import failure

The import procedure tries to preserve the CA certificates, but it doesn't seem to be working right now.

So you will need to delete the service-account-tokens - they will be recreated with the correct keys.

Otherwise some services (most notably DNS) will not work


`kubectl get secrets --all-namespaces`
> ```
NAMESPACE     NAME                              TYPE                                  DATA      AGE
default       default-token-4dgib               kubernetes.io/service-account-token   3         53m
kube-system   default-token-lhfkx               kubernetes.io/service-account-token   3         53m
kube-system   token-admin                       Opaque                                1         53m
kube-system   token-kube-proxy                  Opaque                                1         53m
kube-system   token-kubelet                     Opaque                                1         53m
kube-system   token-system-controller-manager   Opaque                                1         53m
kube-system   token-system-dns                  Opaque                                1         53m
kube-system   token-system-logging              Opaque                                1         53m
kube-system   token-system-monitoring           Opaque                                1         53m
kube-system   token-system-scheduler            Opaque                                1         53m
```

Delete the tokens of type `kubernetes.io/service-account-token`:

```
kubectl delete secret default-token-4dgib
kubectl delete secret --namespace kube-system default-token-lhfkx
```

Then restart the kube-dns pod so it picks up a valid secret:
`kubectl delete pods --namespace kube-system --selector "k8s-app=kube-dns"`

## Other fixes

* If you're using a manually created ELB, the auto-scaling groups change, so you will need to reconfigure
your ELBs to include the new auto-scaling group(s).

## Delete remaining resources of the old cluster

`kops delete cluster ${OLD_NAME}`
> ```
TYPE                    NAME                                    ID
autoscaling-config      kubernetes-minion-group-us-west-2a      kubernetes-minion-group-us-west-2a
autoscaling-group       kubernetes-minion                       kubernetes-minion-group-us-west-2a
instance                kubernetes-master                       i-67af2ec8
```

And once you've confirmed it looks right, run with `--yes`

You will also need to release the old ElasticIP manually.
