# USING KOPS WITH COREOS - A MULTI-MASTER/MULTI-NODE PRACTICAL EXAMPLE

## WHAT WE WANT TO ACCOMPLISH HERE?

The exercise described in this document will focus on the following goals:

- Demonstrate how to use a production-setup with 3 masters and multiple working nodes (two).
- Change our default base-distro (Debian 8) for CoreOS stable, available too as an AMI on AWS.
- Ensure our masters are deployed on 3 different AWS availability zones.
- Ensure our nodes are deployed on 2 different AWS availability zones.


## PRE-FLIGHT CHECK:

Please follow our [basic-requirements document](basic-requirements.md) that is common for all our exercises. Ensure the basic requirements are covered before continuing.


## AWS/KOPS ENVIRONMENT INFORMATION SETUP:

First, using some scripting and assuming you already configured your "aws" environment on your linux system, use the following commands in order to export your AWS access/secret (this will work if you are using the default profile):

```bash
export AWS_ACCESS_KEY_ID=`grep aws_access_key_id ~/.aws/credentials|awk '{print $3}'`
export AWS_SECRET_ACCESS_KEY=`grep aws_secret_access_key ~/.aws/credentials|awk '{print $3}'`
echo "$AWS_ACCESS_KEY_ID $AWS_SECRET_ACCESS_KEY"
```

If you are using multiple profiles (and not the default one), you should use the following command instead in order to export your profile:

```bash
export AWS_PROFILE=name_of_your_profile
```

Create a bucket (if you don't already have one) for your cluster state:

```bash
aws s3api create-bucket --bucket my-kops-s3-bucket-for-cluster-state --region us-east-1
```

Then export the name of your cluster along with the "S3" URL of your bucket:

```bash
export NAME=coreosbasedkopscluster.k8s.local
export KOPS_STATE_STORE=s3://my-kops-s3-bucket-for-cluster-state
```

Some things to note from here:

- "NAME" will be an environment variable that we'll use from now in order to refer to our cluster name. For this practical exercise, our cluster name is "coreosbasedkopscluster.k8s.local".
- Because we'll use gossip DNS instead of a valid DNS domain on AWS ROUTE53 service, our cluster name needs to include the string **".k8s.local"** at the end (this is covered on our AWS tutorials). You can see more about this on our [Getting Started Doc.](../getting_started/aws.md)


## COREOS IMAGE INFORMATION:

CoreOS webpage includes a "json" with the updated list of latest images: [https://coreos.com/dist/aws/aws-stable.json](https://coreos.com/dist/aws/aws-stable.json)

By using "jq" you can obtain the "ami" for a specific region


```bash
curl -s https://coreos.com/dist/aws/aws-stable.json | jq -r '.["us-east-1"].hvm'
"ami-32705b49"
```

The last command will check the all "hvm" CoreOS images on us-east-1 region. Please, always use "hvm" images.

At the moment we created this document, our ami was: "ami-32705b49". More info about the image can be obtained by using the following "aws-cli" command:

```bash
aws ec2 describe-images --image-id ami-32705b49 --output table

--------------------------------------------------------------------------
|                             DescribeImages                             |
+------------------------------------------------------------------------+
||                                Images                                ||
|+---------------------+------------------------------------------------+|
||  Architecture       |  x86_64                                        ||
||  CreationDate       |  2017-08-10T02:07:16.000Z                      ||
||  Description        |  CoreOS Container Linux stable 1409.8.0 (HVM)  ||
||  EnaSupport         |  True                                          ||
||  Hypervisor         |  xen                                           ||
||  ImageId            |  ami-32705b49                                  ||
||  ImageLocation      |  595879546273/CoreOS-stable-1409.8.0-hvm       ||
||  ImageType          |  machine                                       ||
||  Name               |  CoreOS-stable-1409.8.0-hvm                    ||
||  OwnerId            |  595879546273                                  ||
||  Public             |  True                                          ||
||  RootDeviceName     |  /dev/xvda                                     ||
||  RootDeviceType     |  ebs                                           ||
||  SriovNetSupport    |  simple                                        ||
||  State              |  available                                     ||
||  VirtualizationType |  hvm                                           ||
|+---------------------+------------------------------------------------+|
|||                         BlockDeviceMappings                        |||
||+-----------------------------------+--------------------------------+||
|||  DeviceName                       |  /dev/xvda                     |||
|||  VirtualName                      |                                |||
||+-----------------------------------+--------------------------------+||
||||                                Ebs                               ||||
|||+------------------------------+-----------------------------------+|||
||||  DeleteOnTermination         |  True                             ||||
||||  Encrypted                   |  False                            ||||
||||  SnapshotId                  |  snap-00d2949d7084cd408           ||||
||||  VolumeSize                  |  8                                ||||
||||  VolumeType                  |  standard                         ||||
|||+------------------------------+-----------------------------------+|||
|||                         BlockDeviceMappings                        |||
||+----------------------------------+---------------------------------+||
|||  DeviceName                      |  /dev/xvdb                      |||
|||  VirtualName                     |  ephemeral0                     |||
||+----------------------------------+---------------------------------+||
```

Also, you can obtaing the image owner/name using the following aws-cli command:

```bash
aws ec2 describe-images --region=us-east-1 --owner=595879546273 \
    --filters "Name=virtualization-type,Values=hvm" "Name=name,Values=CoreOS-stable*" \
    --query 'sort_by(Images,&CreationDate)[-1].{id:ImageLocation}' \
	--output table


---------------------------------------------------
|                 DescribeImages                  |
+----+--------------------------------------------+
|  id|  595879546273/CoreOS-stable-1409.8.0-hvm   |
+----+--------------------------------------------+
```

Then, our image for CoreOS, in "AMI" format is "ami-32705b49", or in owner/name format "595879546273/CoreOS-stable-1409.8.0-hvm". Note that KOPS default image is a debian-jessie based one (more specifically: "kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2017-05-02" at the moment we are writing this document).

**NOTE:** Always obtain the latest image before deploying KOPS. CoreOS updates it's AWS image very often. Don't rely on the versions included on this document. Always check first.


## KOPS CLUSTER CREATION AND MODIFICATION:

Let's first create our cluster ensuring a multi-master setup with 3 masters in a multi-az setup, two worker nodes also in a multi-az setup, and specifying the CoreOS AMI:

```bash
kops create cluster \
--master-zones=us-east-1a,us-east-1b,us-east-1c \
--zones=us-east-1a,us-east-1b,us-east-1c \
--node-count=2 \
--image ami-32705b49 \
${NAME}
```

A few things to note here:

- The environment variable ${NAME} was previously exported with our cluster name: coreosbasedkopscluster.k8s.local.
- For true HA at the master level, we need to pick a region with at least 3 availability zones. For this practical exercise, we are using "us-east-1" AWS region which contains 5 availability zones (az's for short): us-east-1a, us-east-1b, us-east-1c, us-east-1d and us-east-1e.
- The "--master-zones=us-east-1a,us-east-1b,us-east-1c" KOPS argument will actually enforce that we want 3 masters here. "--node-count=2" only applies to the worker nodes (not the masters).
- The "--image ami-32705b49" KOPS argument will enforce the usage or our desired image: CoreOS Stable 1409.8.0. You can use here any of the aforementioned formats: "ami-32705b49" or "595879546273/CoreOS-stable-1409.8.0-hvm". KOPS will understand both ways to indicate the AMI we want to use here.

With those points clarified, let's deploy our cluster:


```bash
kops update cluster ${NAME} --yes
```

Go for a coffee or just take a 10~15 minutes walk. After that, the cluster will be up-and-running. We can check this with the following commands:

```bash
kops validate cluster

Using cluster from kubectl context: coreosbasedkopscluster.k8s.local

Validating cluster coreosbasedkopscluster.k8s.local

INSTANCE GROUPS
NAME                    ROLE    MACHINETYPE     MIN     MAX     SUBNETS
master-us-east-1a       Master  m3.medium       1       1       us-east-1a
master-us-east-1b       Master  c4.large        1       1       us-east-1b
master-us-east-1c       Master  m3.medium       1       1       us-east-1c
nodes                   Node    t2.medium       2       2       us-east-1a,us-east-1b,us-east-1c

NODE STATUS
NAME                            ROLE    READY
ip-172-20-125-216.ec2.internal  node    True
ip-172-20-125-90.ec2.internal   master  True
ip-172-20-48-12.ec2.internal    master  True
ip-172-20-79-203.ec2.internal   master  True
ip-172-20-92-185.ec2.internal   node    True

Your cluster coreosbasedkopscluster.k8s.local is ready

```

Before continuing, let's note something interesting here: Can you see your masters? Two of them (master-us-east-1a and master-us-east-1c) are using "m3.medium" "aws instance type", but "master-us-east-1b" is using "c4.large". This happens because KOPS uses the AWS API in order to determine if the required instance type is available on the "az". At the moment we launched this cluster, "m3.medium" was unavailable on "us-east-1b". This forced KOPS to choose the nearest instance type candidate on the AZ.

If you don't want KOPS to auto-select the instance type, you can use the following arguments in order to enforce the instance types for both masters and nodes:

- Specify the node size: --node-size=m4.large
- Specify the master size: --master-size=m4.large

But, before doing that, always ensure the instance types are available on your desired AZ.

NOTE: More arguments and kops commands are described [here](../cli/kops.md).

Let's continue exploring our cluster, but now with "kubectl":

```bash
kubectl get nodes --show-labels

NAME                             STATUS    AGE       VERSION   LABELS
ip-172-20-125-216.ec2.internal   Ready     6m        v1.7.0    beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=t2.medium,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=us-east-1,failure-domain.beta.kubernetes.io/zone=us-east-1c,kubernetes.io/hostname=ip-172-20-125-216.ec2.internal,kubernetes.io/role=node,node-role.kubernetes.io/node=
ip-172-20-125-90.ec2.internal    Ready     7m        v1.7.0    beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=m3.medium,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=us-east-1,failure-domain.beta.kubernetes.io/zone=us-east-1c,kubernetes.io/hostname=ip-172-20-125-90.ec2.internal,kubernetes.io/role=master,node-role.kubernetes.io/master=
ip-172-20-48-12.ec2.internal     Ready     3m        v1.7.0    beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=m3.medium,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=us-east-1,failure-domain.beta.kubernetes.io/zone=us-east-1a,kubernetes.io/hostname=ip-172-20-48-12.ec2.internal,kubernetes.io/role=master,node-role.kubernetes.io/master=
ip-172-20-79-203.ec2.internal    Ready     7m        v1.7.0    beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=c4.large,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=us-east-1,failure-domain.beta.kubernetes.io/zone=us-east-1b,kubernetes.io/hostname=ip-172-20-79-203.ec2.internal,kubernetes.io/role=master,node-role.kubernetes.io/master=
ip-172-20-92-185.ec2.internal    Ready     6m        v1.7.0    beta.kubernetes.io/arch=amd64,beta.kubernetes.io/instance-type=t2.medium,beta.kubernetes.io/os=linux,failure-domain.beta.kubernetes.io/region=us-east-1,failure-domain.beta.kubernetes.io/zone=us-east-1b,kubernetes.io/hostname=ip-172-20-92-185.ec2.internal,kubernetes.io/role=node,node-role.kubernetes.io/node=
```

```bash
kubectl -n kube-system get pods

NAME                                                    READY     STATUS    RESTARTS   AGE
dns-controller-3497129722-rt4nv                         1/1       Running   0          7m
etcd-server-events-ip-172-20-125-90.ec2.internal        1/1       Running   0          7m
etcd-server-events-ip-172-20-48-12.ec2.internal         1/1       Running   0          3m
etcd-server-events-ip-172-20-79-203.ec2.internal        1/1       Running   0          7m
etcd-server-ip-172-20-125-90.ec2.internal               1/1       Running   0          7m
etcd-server-ip-172-20-48-12.ec2.internal                1/1       Running   0          3m
etcd-server-ip-172-20-79-203.ec2.internal               1/1       Running   0          7m
kube-apiserver-ip-172-20-125-90.ec2.internal            1/1       Running   0          7m
kube-apiserver-ip-172-20-48-12.ec2.internal             1/1       Running   0          3m
kube-apiserver-ip-172-20-79-203.ec2.internal            1/1       Running   0          7m
kube-controller-manager-ip-172-20-125-90.ec2.internal   1/1       Running   0          7m
kube-controller-manager-ip-172-20-48-12.ec2.internal    1/1       Running   0          3m
kube-controller-manager-ip-172-20-79-203.ec2.internal   1/1       Running   0          7m
kube-dns-479524115-28zqc                                3/3       Running   0          8m
kube-dns-479524115-7xv6b                                3/3       Running   0          6m
kube-dns-autoscaler-1818915203-zf0gd                    1/1       Running   0          8m
kube-proxy-ip-172-20-125-216.ec2.internal               1/1       Running   0          6m
kube-proxy-ip-172-20-125-90.ec2.internal                1/1       Running   0          7m
kube-proxy-ip-172-20-48-12.ec2.internal                 1/1       Running   0          3m
kube-proxy-ip-172-20-79-203.ec2.internal                1/1       Running   0          7m
kube-proxy-ip-172-20-92-185.ec2.internal                1/1       Running   0          7m
kube-scheduler-ip-172-20-125-90.ec2.internal            1/1       Running   0          7m
kube-scheduler-ip-172-20-48-12.ec2.internal             1/1       Running   0          3m
kube-scheduler-ip-172-20-79-203.ec2.internal            1/1       Running   0          8m

```


## LAUNCHING A SIMPLE REPLICATED APP ON THE CLUSTER.

Before doing the tasks ahead, we created a simple "webservers" security group inside our KOPS's cluster VPC (using the AWS WEB-UI) allowing inbound port 80 and applied it to our two nodes (not the masters). Then, with the following command we proceed to create a simple replicated app in our coreos-based kops-launched cluster:

```
kubectl run apache-simple-replicated \
--image=httpd:2.4-alpine \
--replicas=2 \
--port=80 \
--hostport=80
```

Then check it:

```bash
kubectl get pods -o wide
NAME                                        READY     STATUS    RESTARTS   AGE       IP           NODE
apache-simple-replicated-1977341696-3hxxx   1/1       Running   0          31s       100.96.2.3   ip-172-20-92-185.ec2.internal
apache-simple-replicated-1977341696-zv4fn   1/1       Running   0          31s       100.96.3.4   ip-172-20-125-216.ec2.internal
```

Using our public IP's (the ones from our kube nodes, again, not the masters):

```bash
curl http://54.210.119.98
<html><body><h1>It works!</h1></body></html>

curl http://34.200.247.63
<html><body><h1>It works!</h1></body></html>

```

**NOTE:** If you are replicating this exercise in a production environment, use a "real" load balancer in order to expose your replicated services. We are here just testing things so we really don't care right now about that, but, if you are doing this for a "real" production environment, either use an AWS ELB service, or an nginx ingress controller as described in our documentation: [NGINX Based ingress controller](https://github.com/kubernetes/kops/tree/master/addons/ingress-nginx).

Now, let's delete our recently-created deployment:

```bash
kubectl delete deployment apache-simple-replicated
```

NOTE: In the AWS Gui, we also deleted our "webservers" security group (after removing it from out instance nodes).

Check again:

```bash
kubectl get pods -o wide
No resources found.
```

Finally, let's destroy our cluster:

```
kops delete cluster ${NAME} --yes
```

After a brief time, your cluster will be fully deleted on AWS and you'll see the following output:

```bash
Deleted cluster: "coreosbasedkopscluster.k8s.local"
```

**NOTE:** Before destroying the cluster, "really ensure" any extra security group "not created" directly by KOPS has been removed by you. Otherwise, KOPS will be unable to delete the cluster.