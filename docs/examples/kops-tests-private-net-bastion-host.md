# USING KOPS WITH PRIVATE NETWORKING AND A BASTION HOST IN A HIGHLY-AVAILABLE SETUP

## WHAT WE WANT TO ACCOMPLISH HERE?

The exercise described in this document will focus on the following goals:

- Demonstrate how to use a production-setup with 3 masters and two workers in different availability zones.
- Demonstrate how to use a private networking setup with a bastion host.
- Ensure our masters are deployed on 3 different AWS availability zones.
- Ensure our nodes are deployed on 2 different AWS availability zones.
- Add true high-availability to the bastion instance group.


## PRE-FLIGHT CHECK:

Please follow our [basic-requirements document](basic-requirements.md) that is common for all our exercises. Ensure the basic requirements are covered before continuing.


## AWS/KOPS ENVIRONMENT SETUP:

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
export NAME=privatekopscluster.k8s.local
export KOPS_STATE_STORE=s3://my-kops-s3-bucket-for-cluster-state
```

Some things to note from here:

- "NAME" will be an environment variable that we'll use from now in order to refer to our cluster name. For this practical exercise, our cluster name is "privatekopscluster.k8s.local".
- Because we'll use gossip DNS instead of a valid DNS domain on AWS ROUTE53 service, our cluster name need to include the string **".k8s.local"** at the end (this is covered on our AWS tutorials). You can see more about this on our [Getting Started Doc.](../getting_started/aws.md)


## KOPS PRIVATE CLUSTER CREATION:

Let's first create our cluster ensuring a multi-master setup with 3 masters in a multi-az setup, two worker nodes also in a multi-az setup, and using both private networking and a bastion server:

```bash
kops create cluster \
--cloud=aws \
--master-zones=us-east-1a,us-east-1b,us-east-1c \
--zones=us-east-1a,us-east-1b,us-east-1c \
--node-count=2 \
--topology private \
--networking kopeio-vxlan \
--node-size=t2.micro \
--master-size=t2.micro \
${NAME}
```

A few things to note here:

- The environment variable ${NAME} was previously exported with our cluster name: privatekopscluster.k8s.local.
- "--cloud=aws": As kops grows and begin to support more clouds, we need to tell the command to use the specific cloud we want for our deployment. In this case: amazon web services (aws).
- For true HA (high availability) at the master level, we need to pick a region with 3 availability zones. For this practical exercise, we are using "us-east-1" AWS region which contains 5 availability zones (az's for short): us-east-1a, us-east-1b, us-east-1c, us-east-1d and us-east-1e. We used "us-east-1a,us-east-1b,us-east-1c" for our masters.
- The "--master-zones=us-east-1a,us-east-1b,us-east-1c" KOPS argument will actually enforce we want 3 masters here. "--node-count=2" only applies to the worker nodes (not the masters). Again, real "HA" on Kubernetes control plane requires 3 masters.
- The "--topology private" argument will ensure that all our instances will have private IP's and no public IP's from amazon.
- We are including the arguments "--node-size" and "master-size" to specify the "instance types" for both our masters and worker nodes.
- Because we are just doing a simple LAB, we are using "t2.micro" machines. Please DON'T USE t2.micro on real production systems. Start with "t2.medium" as a minimum realistic/workable machine type.
- And finally, the "--networking kopeio-vxlan" argument. With the private networking model, we need to tell kops which networking subsystem to use. More information about kops supported networking models can be obtained from the [KOPS Kubernetes Networking Documentation](../networking.md). For this exercise we'll use "kopeio-vxlan" (or "kopeio" for short).

**NOTE**: You can add the "--bastion" argument here if you are not using "gossip dns" and create the bastion from start, but if you are using "gossip-dns" this will make this cluster to fail (this is a bug we are correcting now). For the moment don't use "--bastion" when using gossip DNS. We'll show you how to get around this by first creating the private cluster, then creation the bastion instance group once the cluster is running.

With those points clarified, let's deploy our cluster:

```bash
kops update cluster ${NAME} --yes
```

Go for a coffee or just take a 10~15 minutes walk. After that, the cluster will be up-and-running. We can check this with the following commands:

```bash
kops validate cluster

Using cluster from kubectl context: privatekopscluster.k8s.local

Validating cluster privatekopscluster.k8s.local

INSTANCE GROUPS
NAME                    ROLE    MACHINETYPE     MIN     MAX     SUBNETS
master-us-east-1a       Master  t2.micro        1       1       us-east-1a
master-us-east-1b       Master  t2.micro        1       1       us-east-1b
master-us-east-1c       Master  t2.micro        1       1       us-east-1c
nodes                   Node    t2.micro        2       2       us-east-1a,us-east-1b,us-east-1c

NODE STATUS
NAME                            ROLE    READY
ip-172-20-111-44.ec2.internal   master  True
ip-172-20-44-102.ec2.internal   node    True
ip-172-20-53-10.ec2.internal    master  True
ip-172-20-64-151.ec2.internal   node    True
ip-172-20-74-55.ec2.internal    master  True

Your cluster privatekopscluster.k8s.local is ready
```

The ELB created by kops will expose the Kubernetes API trough "https" (configured on our ~/.kube/config file):

```bash
grep server ~/.kube/config

server: https://api-privatekopscluster-k8-djl5jb-1946625559.us-east-1.elb.amazonaws.com
```

But, all the cluster instances (masters and worker nodes) will have private IP's only (no AWS public IP's). Then, in order to reach our instances, we need to add a "bastion host" to our cluster.


## ADDING A BASTION HOST TO OUR CLUSTER.

We mentioned earlier that we can't add the "--bastion" argument to our "kops create cluster" command if we are using "gossip dns" (a fix it's on the way as we speaks). That forces us to add the bastion afterwards, once the cluster is up and running.

Let's add a bastion here by using the following command:

```bash
kops create instancegroup bastions --role Bastion --subnet utility-us-east-1a --name ${NAME}
```

**Explanation of this command:**
- This command will add to our cluster definition a new instance group called "bastions" with the "Bastion" role on the aws subnet "utility-us-east-1a". Note that the "Bastion" role need the first letter to be a capital (Bastion=ok, bastion=not ok).
- The subnet "utility-us-east-1a" was created when we created our cluster the first time. KOPS add the "utility-" prefix to all subnets created on all specified AZ's. In other words, if we instructed kops to deploy our instances on us-east-1a, use-east-1b and use-east-1c, kops will create the subnets "utility-us-east-1a", "utility-us-east-1b" and "utility-us-east-1c". Because we need to tell kops where to deploy our bastion (or bastions), we need to specify the subnet.

You'll see the following output in your editor when you can change your bastion group size and add more networks.

```bash
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: null
  name: bastions
spec:
  image: kope.io/k8s-1.7-debian-jessie-amd64-hvm-ebs-2017-07-28
  machineType: t2.micro
  maxSize: 1
  minSize: 1
  role: Bastion
  subnets:
  - utility-us-east-1a
```

If want a H.A. setup for your bastions, modify minSize and maxSize and add more subnets. We'll do this later on this exercise.

Save this and deploy the changes:

```bash
kops update cluster ${NAME} --yes
```

You will see an output like the following:

```bash
I0828 13:06:33.153920   16528 apply_cluster.go:420] Gossip DNS: skipping DNS validation
I0828 13:06:34.686722   16528 executor.go:91] Tasks: 0 done / 116 total; 40 can run
I0828 13:06:36.181677   16528 executor.go:91] Tasks: 40 done / 116 total; 26 can run
I0828 13:06:37.602302   16528 executor.go:91] Tasks: 66 done / 116 total; 34 can run
I0828 13:06:39.116916   16528 launchconfiguration.go:327] waiting for IAM instance profile "bastions.privatekopscluster.k8s.local" to be ready
I0828 13:06:49.761535   16528 executor.go:91] Tasks: 100 done / 116 total; 9 can run
I0828 13:06:50.897272   16528 executor.go:91] Tasks: 109 done / 116 total; 7 can run
I0828 13:06:51.516158   16528 executor.go:91] Tasks: 116 done / 116 total; 0 can run
I0828 13:06:51.944576   16528 update_cluster.go:247] Exporting kubecfg for cluster
Kops has set your kubectl context to privatekopscluster.k8s.local

Cluster changes have been applied to the cloud.


Changes may require instances to restart: kops rolling-update cluster
```

This is "kops" creating the instance group with your bastion instance. Let's validate our cluster:

```bash
kops validate cluster
Using cluster from kubectl context: privatekopscluster.k8s.local

Validating cluster privatekopscluster.k8s.local

INSTANCE GROUPS
NAME                    ROLE    MACHINETYPE     MIN     MAX     SUBNETS
bastions                Bastion t2.micro        1       1       utility-us-east-1a
master-us-east-1a       Master  t2.micro        1       1       us-east-1a
master-us-east-1b       Master  t2.micro        1       1       us-east-1b
master-us-east-1c       Master  t2.micro        1       1       us-east-1c
nodes                   Node    t2.micro        2       2       us-east-1a,us-east-1b,us-east-1c

NODE STATUS
NAME                            ROLE    READY
ip-172-20-111-44.ec2.internal   master  True
ip-172-20-44-102.ec2.internal   node    True
ip-172-20-53-10.ec2.internal    master  True
ip-172-20-64-151.ec2.internal   node    True
ip-172-20-74-55.ec2.internal    master  True

Your cluster privatekopscluster.k8s.local is ready
```

Our bastion instance group is there. Also, kops created an ELB for our "bastions" instance group that we can check with the following command:

```bash
aws elb --output=table describe-load-balancers|grep DNSName.\*bastion|awk '{print $4}'
bastion-privatekopscluste-bgl0hp-1327959377.us-east-1.elb.amazonaws.com
```

For this LAB, the "ELB" FQDN is "bastion-privatekopscluste-bgl0hp-1327959377.us-east-1.elb.amazonaws.com" We can "ssh" to it:

```bash
ssh -i ~/.ssh/id_rsa admin@bastion-privatekopscluste-bgl0hp-1327959377.us-east-1.elb.amazonaws.com

The programs included with the Debian GNU/Linux system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Debian GNU/Linux comes with ABSOLUTELY NO WARRANTY, to the extent
permitted by applicable law.
Last login: Mon Aug 28 18:07:16 2017 from 172.20.0.238
```

Because we really want to use a ssh-agent, start it first (this will :

```bash
eval `ssh-agent -s`
```

And add your key to the agent with "ssh-add":

```bash
ssh-add ~/.ssh/id_rsa

Identity added: /home/kops/.ssh/id_rsa (/home/kops/.ssh/id_rsa)
```

Then, ssh to your bastion ELB FQDN

```bash
ssh -A admin@bastion-privatekopscluste-bgl0hp-1327959377.us-east-1.elb.amazonaws.com
```

Or if you want to automate it:

```bash
ssh -A admin@`aws elb --output=table describe-load-balancers|grep DNSName.\*bastion|awk '{print $4}'`
```

And from the bastion, you can ssh to your masters or workers:

```bash
admin@ip-172-20-2-64:~$ ssh admin@ip-172-20-53-10.ec2.internal

The authenticity of host 'ip-172-20-53-10.ec2.internal (172.20.53.10)' can't be established.
ECDSA key fingerprint is d1:30:c6:5e:77:ff:cd:d2:7d:1f:f9:12:e3:b0:28:e4.
Are you sure you want to continue connecting (yes/no)? yes
Warning: Permanently added 'ip-172-20-53-10.ec2.internal,172.20.53.10' (ECDSA) to the list of known hosts.

The programs included with the Debian GNU/Linux system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Debian GNU/Linux comes with ABSOLUTELY NO WARRANTY, to the extent
permitted by applicable law.

admin@ip-172-20-53-10:~$
```

**NOTE:** Remember that you can obtain the local DNS names from your "kops validate cluster" command, or, with the "kubectl get nodes" command. We recommend the first (kops validate cluster) because it will tell you who are the masters and who the worker nodes:


```bash
kops validate cluster
Using cluster from kubectl context: privatekopscluster.k8s.local

Validating cluster privatekopscluster.k8s.local

INSTANCE GROUPS
NAME                    ROLE    MACHINETYPE     MIN     MAX     SUBNETS
bastions                Bastion t2.micro        1       1       utility-us-east-1a
master-us-east-1a       Master  t2.micro        1       1       us-east-1a
master-us-east-1b       Master  t2.micro        1       1       us-east-1b
master-us-east-1c       Master  t2.micro        1       1       us-east-1c
nodes                   Node    t2.micro        2       2       us-east-1a,us-east-1b,us-east-1c

NODE STATUS
NAME                            ROLE    READY
ip-172-20-111-44.ec2.internal   master  True
ip-172-20-44-102.ec2.internal   node    True
ip-172-20-53-10.ec2.internal    master  True
ip-172-20-64-151.ec2.internal   node    True
ip-172-20-74-55.ec2.internal    master  True

Your cluster privatekopscluster.k8s.local is ready
```

## MAKING THE BASTION LAYER "HIGHLY AVAILABLE".

If for any reason any "legendary monster from the comics" decides to destroy the amazon AZ that contains our bastion, we'll basically be unable to enter to our instances. Let's add some H.A. to our bastion layer and force amazon to deploy additional bastion instances on other availability zones.

First, let's edit our "bastions" instance group:

```bash
kops edit ig bastions --name ${NAME}
```

And change minSize/maxSize to 3 (3 instances) and add more subnets:

```bash
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-08-28T17:05:23Z
  labels:
    kops.k8s.io/cluster: privatekopscluster.k8s.local
  name: bastions
spec:
  image: kope.io/k8s-1.7-debian-jessie-amd64-hvm-ebs-2017-07-28
  machineType: t2.micro
  maxSize: 3
  minSize: 3
  role: Bastion
  subnets:
  - utility-us-east-1a
  - utility-us-east-1b
  - utility-us-east-1c
```

Save the changes, and update your cluster:

```bash
kops update cluster ${NAME} --yes
```

**NOTE:** After the update command, you'll see the following recurring error:

```bash
W0828 15:22:46.461033    5852 executor.go:109] error running task "LoadBalancer/bastion.privatekopscluster.k8s.local" (1m5s remaining to succeed): subnet changes on LoadBalancer not yet implemented: actual=[subnet-c029639a] -> expected=[subnet-23f8a90f subnet-4a24ef2e subnet-c029639a]
```

This happens because the original ELB created by "kops" only contained the subnet "utility-us-east-1a" and it can't add the additional subnets. In order to fix this, go to your AWS console and add the remaining subnets in your ELB. Then the recurring error will disappear and your bastion layer will be fully redundant.

**NOTE:** Always think ahead: If you are creating a fully redundant cluster (with fully redundant bastions), always configure the redundancy from the beginning.

When you are finished playing with kops, then destroy/delete your cluster:

Finally, let's destroy our cluster:

```bash
kops delete cluster ${NAME} --yes
```
