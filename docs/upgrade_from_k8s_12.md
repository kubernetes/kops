# Upgrading from k8s 1.2

** This is an experimental and slightly risky procedure, so we recommend backing up important data before proceeding. 
Take a snapshot of your EBS volumes; export all your data from kubectl etc. **

Limitations:

* kops splits etcd onto two volumes now: `main` and `events`.  We will keep the `main` data, but
  you will lose your events history.
* Doubtless others not yet known - please open issues if you encounter them!

## Overview

There are a few steps:

* First you import the existing cluster state, so you can see and edit the configuration
* You verify the cluster configuration
* You move existing AWS resources to your new cluster
* You bring up the new cluster
* You probably need to do a little manual cleanup (for example of ELBs)
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
kops edit cluster --name ${OLD_NAME}
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
kops upgrade cluster --newname ${NEW_NAME} --name ${OLD_NAME}
```

If you now list the clusters, you should see both the old cluster & the new cluster

```kops get clusters```

## Bring up the new cluster

Use the normal tool to bring up the new cluster:

```
kops create cluster --name ${NEW_NAME} --dryrun
```

Things to check are that it is reusing the existing volume for the _main_ etcd cluster (but not the events clusters).

And then when you are happy:

```
kops create cluster --name ${NEW_NAME}
```


## Export kubecfg settings to access the new cluster

```
kops export kubecfg ${NEW_NAME}
```

Within a few minutes the new cluster should be running.  Try `kubectl get nodes --show-labels`, `kubectl get pods` etc until you are sure that all is well.

## Workaround for secret import failure

The import procedure tries to preserve the CA certificates, but it doesn't seem to be working right now.

So you will need to delete the service-account-tokens - they will be recreated with the correct keys.

Otherwise some services (most notably DNS) will not work


```
kubectl get secrets --all-namespaces
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

```
kubectl delete secret default-token-4dgib
kubectl delete secret --namespace kube-system default-token-lhfkx
```

Then restart the kube-dns pod so it picks up a valid secret:

```
kubectl get pods --namespace kube-system
NAME                                                                  READY     STATUS             RESTARTS   AGE
etcd-server-events-ip-172-20-59-170.us-west-2.compute.internal        1/1       Running            0          8m
etcd-server-ip-172-20-59-170.us-west-2.compute.internal               1/1       Running            0          7m
kope-aws-ip-172-20-59-170.us-west-2.compute.internal                  1/1       Running            0          8m
kube-apiserver-ip-172-20-59-170.us-west-2.compute.internal            1/1       Running            3          8m
kube-controller-manager-ip-172-20-59-170.us-west-2.compute.internal   1/1       Running            0          8m
kube-dns-v14-4ygfz                                                    2/3       CrashLoopBackOff   5          8m
kube-proxy-ip-172-20-56-231.us-west-2.compute.internal                1/1       Running            0          8m
kube-proxy-ip-172-20-56-232.us-west-2.compute.internal                1/1       Running            0          8m
kube-proxy-ip-172-20-56-233.us-west-2.compute.internal                1/1       Running            0          8m
kube-scheduler-ip-172-20-59-170.us-west-2.compute.internal            1/1       Running            0          8m
```

```
kubectl delete pod --namespace=kube-system kube-dns-v14-4ygfz
```

## Workaround to re-enable ELBs

Due to a limitation in ELBs (you can't replace all the subnets), if you are using ELBs you have to follow a
workaround procedure to get them to bind to the new subnet.   If you aren't using any services with Type
LoadBalancer, you likely aren't using ELBs and can skip this section!

* `kops edit cluster --name ${NEW_NAME}`
* Add a zone to the `zones` section and save the file (it normally suffices to just add `- name: us-west-2b` or whatever
  zone you are adding; kops will auto-populate the CIDR.
* Preview: `kops create cluster --name ${NEW_NAME} --dryrun`.  You expect to see something like:
```
    *awstasks.AutoscalingGroup    autoscalingGroup/nodes.upgraded.awsdata.com
      Subnets [id:subnet-e3b3d987] -> [id:subnet-e3b3d987, id:<nil>]
```
* Apply changes: `kops create cluster --name ${NEW_NAME}`

In the AWS control panel open the "Load Balancers" section, and for each ELB: 
* On the "Actions" menu click "Edit subnets"
* Add the newly created zone's subnet, then save
* On the "Actions" menu click "Edit subnets" (again)
* Add the other zone's subnet (which will replace the old subnet with the new subnet), and Save

You should now have an ELB in your new zones; within a few minutes k8s should reconcile it and attach the new instances.

## Delete remaining resources of the old cluster

```
kops delete cluster --name ${OLD_NAME}
```

And once you've confirmed it looks right, run with `--yes`

You will also need to release the old ElasticIP manually.

Note that there is an issue in EC2/ELB: it seems that the NetworkInterfaces for the ELB aren't immediately deleted,
and this prevents full teardown of the old resources (the subnet in particular).  A workaround is to delete
the "Network Interfaces" for the old ELB subnet in the AWS console.
