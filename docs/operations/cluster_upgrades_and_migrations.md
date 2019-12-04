# Cluster Version Upgrades and Migrations

At some point you will almost definitely want to upgrade the Kubernetes version of your cluster, or even migrate from a cluster managed/provisioned by another tool to one managed by `kops`. There are a few different ways to accomplish this depending on your existing cluster situation and any requirements for zero-downtime migrations.

- Upgrade an existing `kube-up` managed cluster to one managed by `kops`
    + [The simple method with downtime](#kube-up---kops-downtime)
    + [The more complex method with zero-downtime](#kube-up---kops-sans-downtime)
- [Upgrade a `kops` cluster from one Kubernetes version to another](updates_and_upgrades.md)

## `kube-up` -> `kops`, with downtime

`kops` lets you upgrade an existing 1.x cluster, installed using `kube-up`, to a cluster managed by `kops` running the latest kubernetes version (or the version of your choice).

**This is an experimental and slightly risky procedure, so we recommend backing up important data before proceeding.
Take a snapshot of your EBS volumes; export all your data from kubectl etc.**

Limitations:
* kops splits etcd onto two volumes now: `main` and `events`.  We will keep the `main` data, but you will lose your events history.
* Doubtless others not yet known - please open issues if you encounter them!

### Overview

There are a few steps to upgrade a kubernetes cluster from 1.2 to 1.3:

* First you import the existing cluster state, so you can see and edit the configuration
* You verify the cluster configuration
* You move existing AWS resources to your new cluster
* You bring up the new cluster
* You can then delete the old cluster and its associated resources

### Importing the existing cluster

The `import cluster` command reverse engineers an existing cluster, and creates a cluster configuration.

Make sure you have set `export KOPS_STATE_STORE=s3://<mybucket>`

Then import the cluster; setting `--name` and `--region` to match the old cluster. If you're not sure of the old cluster name, you can find it by looking at the `KubernetesCluster` tag on your AWS resources.

```
export OLD_NAME=kubernetes
export REGION=us-west-2
kops import cluster --region ${REGION} --name ${OLD_NAME}
```

### Verify the cluster configuration

Now have a look at the cluster configuration, to make sure it looks right. If it doesn't, please open an issue.
```
kops get cluster ${OLD_NAME} -oyaml
```

## Move resources to a new cluster

The upgrade moves some resources so they will be adopted by the new cluster.  There are a number of things this step does:

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

### Import the SSH public key

The SSH public key is not easily retrieved from the old cluster, so you must add it:
```
kops create secret --name ${NEW_NAME} sshpublickey admin -i ~/.ssh/id_rsa.pub
```

### Bring up the new cluster

Use the update command to bring up the new cluster:
```
kops update cluster ${NEW_NAME}
```

Things to check are that it is reusing the existing volume for the _main_ etcd cluster (but not the events clusters).

And then when you are happy:
```
kops update cluster ${NEW_NAME} --yes
```


### Export kubecfg settings to access the new cluster

You can export a kubecfg (although update cluster did this automatically): `kops export kubecfg ${NEW_NAME}`

### Workaround for secret import failure

The import procedure tries to preserve the CA certificates, but unfortunately this isn't supported in kubernetes until [#34029](https://github.com/kubernetes/kubernetes/pull/34029) ships (should be in 1.5).

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
kubectl delete pods -lk8s-app=kube-dns-autoscaler --namespace=kube-system
```

You probably also want to delete the imported DNS services from prior versions:
```
kubectl delete rc -lk8s-app=kube-dns --namespace=kube-system
```


Within a few minutes the new cluster should be running.

Try `kubectl get nodes --show-labels`, `kubectl get pods --all-namespaces` etc until you are sure that all is well.

This should work even without being SSH-ed into the master, although it can take a few minutes for DNS to propagate.  If it doesn't work, double-check that you have specified a valid domain name for your cluster, that records have been created in Route53, and that you can resolve those records from your machine (using `nslookup` or `dig`).

### Other fixes

* If you're using a manually created ELB, the auto-scaling groups change, so you will need to reconfigure your ELBs to include the new auto-scaling group(s).

* It is recommended to delete any old kubernetes system services that we might have imported (and replace them with newer versions):

```
kubectl delete rc -lk8s-app=kube-dns --namespace=kube-system

kubectl delete rc -lk8s-app=elasticsearch-logging --namespace=kube-system
kubectl delete rc -lk8s-app=kibana-logging --namespace=kube-system
kubectl delete rc -lk8s-app=kubernetes-dashboard --namespace=kube-system
kubectl delete rc -lk8s-app=influxGrafana --namespace=kube-system

kubectl delete deployment -lk8s-app=heapster --namespace=kube-system
```

## Delete remaining resources of the old cluster

`kops delete cluster ${OLD_NAME}`

```
TYPE                    NAME                                    ID
autoscaling-config      kubernetes-minion-group-us-west-2a      kubernetes-minion-group-us-west-2a
autoscaling-group       kubernetes-minion                       kubernetes-minion-group-us-west-2a
instance                kubernetes-master                       i-67af2ec8
```

And once you've confirmed it looks right, run with `--yes`

You will also need to release the old ElasticIP manually.

## `kube-up` -> `kops`, sans downtime

### Overview

This method provides zero-downtime when migrating a cluster from `kube-up` to `kops`. It does so by creating a logically separate `kops`-managed cluster in the existing `kube-up` VPC and then swapping the DNS entries (or your reverse proxy's upstream) to point to the new cluster's services.

Limitations:
- If you're using the default networking (`kubenet`), there is a account limit of 50 entries in a VPC's route table. If your cluster contains more than ~25 nodes, this strategy, as-is, will not work.
    + Shifting to a CNI-compatible overlay network like `weave`, `kopeio-vxlan` (`kopeio`), `calico`, `canal`, `romana`, and similar. See the [kops networking docs](../networking.md) for more information.
    + One solution is to gradually shift traffic from one cluster to the other, scaling down the number of nodes on the old cluster, and scaling up the number of nodes on the new cluster.

### Steps

1. If using another service to manage a domain's DNS records, delegate cluster-level DNS resolution to Route53 by adding appropriate NS records pointing `cluster.example.com` to Route53's Hosted Zone's nameservers.
2. Create the new cluster's configuration files with kops. For example:
    - `kops create cluster --cloud=aws --zones=us-east-1a,us-east-1b --admin-access=12.34.56.78/32 --dns-zone=cluster.example.com --kubernetes-version=1.4.0 --node-count=14 --node-size=c3.xlarge --master-zones=us-east-1a --master-size=m4.large --vpc=vpc-123abcdef --network-cidr=172.20.0.0/16 cluster.example.com`
    - `--vpc` is the resource id of the existing VPC.
    - `--network-cidr` is the CIDR of the existing VPC.
    - note that `kops` will propose re-naming the existing VPC but the change never occurs.
        - After this process you can manually rename the VPC for consistency.
3. Verify that the CIDR on each of the zone subnets does not overlap with an existing subnet's.
4. Verify the planned changes with `kops update cluster cluster.example.com`
5. Create the cluster with `kops update cluster cluster.example.com --yes`
6. Wait around for the cluster to fully come up and be available. `k get nodes` should return `(master + minions) = 15` available nodes.
7. (Optional) Create the Dashboard with `kubectl create -f https://raw.githubusercontent.com/kubernetes/dashboard/master/src/deploy/recommended/kubernetes-dashboard.yaml`
8. Deploy the existing resource configuration to the new cluster.
9. Confirm that pods on the new cluster are able to access remote resources.
    - For AWS-hosted services, add the generated `nodes.cluster.example.com` security group to the resources that may need it (i.e. ElastiCache, RDS, etc).
10. Confirm that your application works as expected by hitting the services directly.
    - If you have a `LoadBalancer` service, you should be able to access the ELB's DNS name directly (although perhaps with an SSL error) and use your application as expected.
11. Transition traffic from the old cluster to the new cluster. This depends a bit on your infrastructure, but
    - if using a DNS server, update the `CNAME` record for `example.com` to point to the new ELB's DNS name.
    - if using a reverse proxy, update the upstream to point to the new ELB's DNS name.
    - note that if you're proxying through Cloudflare or similar, changes are instantaneous because it's technically a reverse proxy and not a DNS record.
    - if not using Cloudflare or similar, you'll want to update your DNS record's TTL to a very low duration about 48 hours in advance of this change (and then change it back to the previous value once the shift has been finalized).
12. Rejoice.
13. Once traffic has shifted from the old cluster, delete the old resources after confirming that traffic has stabilized and that no new errors are generated.
    - autoscaling groups
        + turn the ASG down to 0 nodes to delete the instances
    - launch configurations
    - all associated EBS volumes (some may not be released after the instances terminate)
    - security groups (`tag:KubernetesCluster : kubernetes`)

## Recovery/Rollback

The only part of this procedure that should affect the users actively using the site is the DNS swap, which should be relatively instantaneous because we're using Cloudflare as a reverse proxy, not just as a nameserver.

To revert back to the old cluster, simply re-swap the entries pointing to the new cluster with the entries from the old cluster.
