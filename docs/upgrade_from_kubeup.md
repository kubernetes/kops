# Upgrading from kube-up to kops

Kops let you upgrade an existing kubernetes cluster installed using kube-up, to a cluster managed by
kops.

** This is a slightly risky procedure, so we recommend backing up important data before proceeding. 
Take a snapshot of your EBS volumes; export all your data from kubectl etc. **

Limitations:

* kops splits etcd onto two volumes now: `main` and `events`.  We will keep the `main` data, but
  you will lose your events history.

## Overview

There are a few steps to upgrade a kubernetes cluster:

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


## Workaround for secret import failure

The import procedure tries to preserve the CA certificates, but unfortunately this isn't supported
in kubernetes until [#34029](https://github.com/kubernetes/kubernetes/pull/34029) ships (should be
in 1.5).

So you will need to delete the service-accounts, so they can be recreated with the correct keys.

Unfortunately, until you do this, some services (most notably internal & external DNS) will not work.
Because of that you must SSH to the master to do this repair.

You can get the public IP address of the master from the AWS console, or by doing this:

```
aws ec2 --region $REGION describe-instances \
    --filter Name=tag:KubernetesCluster,Values=${NEW_NAME} \
             Name=tag-key,Values=k8s.io/role/master \
             Name=instance-state-name,Values=running \
    --query Reservations[].Instances[].PublicIpAddress \
    --output text
```

Then `ssh admin@<ip>` (the SSH key will be the one you added above, i.e. `~/.ssh/id_rsa.pub`), and run:

First check that the apiserver is running:
```
kubectl get nodes
```

You should see only one node (the master).  Then run
```
NS=`kubectl get namespaces -o 'jsonpath={.items[*].metadata.name}'`
for i in ${NS}; do kubectl get secrets --namespace=${i} --no-headers | grep "kubernetes.io/service-account-token" | awk '{print $1}' | xargs -I {} kubectl delete secret --namespace=$i {}; done
sleep 60 # Allow for new secrets to be created
kubectl delete pods -lk8s-app=dns-controller --namespace=kube-system
kubectl delete pods -lk8s-app=kube-dns --namespace=kube-system
```


You probably also want to delete the imported DNS services from prior versions:

```
kubectl delete rc -lk8s-app=kube-dns --namespace=kube-system # Will work for k8s <= 1.4
kubectl delete deployment --namespace=kube-system  kube-dns # Will work for k8s >= 1.5
```


Within a few minutes the new cluster should be running.

Try `kubectl get nodes --show-labels`, `kubectl get pods --all-namespaces` etc until you are sure that all is well.

This should work even without being SSH-ed into the master, although it can take a few minutes
for DNS to propagate.  If it doesn't work, double-check that you have specified a valid
domain name for your cluster, that records have been created in Route53, and that you
can resolve those records from your machine (using `nslookup` or `dig`).

## Other fixes

* If you're using a manually created ELB, the auto-scaling groups change, so you will need to reconfigure
your ELBs to include the new auto-scaling group(s).

* It is recommended to delete old kubernetes system services that we imported (and replace them with newer versions):

```
kubectl delete rc -lk8s-app=kube-dns --namespace=kube-system       # <= 1.4
kubectl delete deployment --namespace=kube-system  kube-dns        # 1.5

kubectl delete rc -lk8s-app=elasticsearch-logging --namespace=kube-system

kubectl delete rc -lk8s-app=kibana-logging --namespace=kube-system                   # <= 1.4
kubectl delete deployment -lk8s-app=kibana-logging --namespace=kube-system           # 1.5

kubectl delete rc -lk8s-app=kubernetes-dashboard --namespace=kube-system             # <= 1.4
kubectl delete deployment -lk8s-app=kubernetes-dashboard --namespace=kube-system     # 1.5

kubectl delete rc -lk8s-app=influxGrafana --namespace=kube-system

kubectl delete deployment -lk8s-app=heapster --namespace=kube-system
```

## Delete remaining resources of the old cluster

`kops delete cluster ${OLD_NAME}`
> 
```
TYPE                    NAME                                    ID
autoscaling-config      kubernetes-minion-group-us-west-2a      kubernetes-minion-group-us-west-2a
autoscaling-group       kubernetes-minion                       kubernetes-minion-group-us-west-2a
instance                kubernetes-master                       i-67af2ec8
```

And once you've confirmed it looks right, run with `--yes`

You will also need to release the old ElasticIP manually.
