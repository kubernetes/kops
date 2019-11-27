# USING KOPS WITH A ROUTE53 BASED SUBDOMAIN AND SCALING UP THE CLUSTER

## WHAT WE WANT TO ACCOMPLISH HERE/

The exercise described in this document will focus on the following goals:

- Demonstrate how to use a production-setup with 3 masters and two workers in different availability zones.
- Ensure our masters are deployed on 3 different AWS availability zones.
- Ensure our nodes are deployed on 2 different AWS availability zones.
- Use AWS Route53 service for the cluster DNS sub-domain.
- Show how to properly scale-up our cluster.


## PRE-FLIGHT CHECK:

Please follow our [basic-requirements document](basic-requirements.md) that is common for all our exercises. Ensure the basic requirements are covered before continuing.


## DNS Setup - AWS Route53

For our setup we already have a hosted DNS domain in AWS:

```bash
 aws route53 list-hosted-zones --output=table
------------------------------------------------------------------------------------------------------------------
|                                                 ListHostedZones                                                |
+----------------------------------------------------------------------------------------------------------------+
||                                                  HostedZones                                                 ||
|+---------------------------------------+-----------------------------+--------------+-------------------------+|
||            CallerReference            |             Id              |    Name      | ResourceRecordSetCount  ||
|+---------------------------------------+-----------------------------+--------------+-------------------------+|
||  C0461665-01D8-463B-BF2D-62F1747A16DB |  /hostedzone/ZTKK4EXR1EWR5  |  example.org. |  2                      ||
|+---------------------------------------+-----------------------------+--------------+-------------------------+|
|||                                                   Config                                                   |||
||+-------------------------------------------------------------------+----------------------------------------+||
|||  PrivateZone                                                      |  False                                 |||
||+-------------------------------------------------------------------+----------------------------------------+||
```

We can also check that our domain is reachable from the Internet using "dig". You can use other "dns" tools too, but we recommend to use dig (available on all modern linux distributions and other unix-like operating systems. Normally, dig is part of bind-tools and other bind-related packages):

```bash
dig +short example.org soa

ns-656.awsdns-18.net. awsdns-hostmaster.amazon.com. 1 7200 900 1209600 86400

dig +short example.org ns

ns-1056.awsdns-04.org.
ns-656.awsdns-18.net.
ns-9.awsdns-01.com.
ns-1642.awsdns-13.co.uk.
```

If both the "soa" and "ns" queries anwers OK, and with the data pointing to amazon, we are set and we can continue. Please always check that your Route53 hosted DNS zone is working before doing anything else.

Now, let's create a subdomain that we'll use for our cluster:

```bash
export ID=$(uuidgen)
echo $ID
ae852c68-78b3-41af-85ee-997fc470fd1c

aws route53 \
create-hosted-zone \
--output=json \
--name kopsclustertest.example.org \
--caller-reference $ID | \
jq .DelegationSet.NameServers

[
  "ns-1383.awsdns-44.org",
  "ns-829.awsdns-39.net",
  "ns-346.awsdns-43.com",
  "ns-1973.awsdns-54.co.uk"
]
```

Note that the last command (`aws route53 create-hosted-zone`) will output your name servers for the subdomain:

```bash
[
  "ns-1383.awsdns-44.org",
  "ns-829.awsdns-39.net",
  "ns-346.awsdns-43.com",
  "ns-1973.awsdns-54.co.uk"
]
```

We need the zone parent ID too. We can obtain it with the following command:

```bash
aws route53 --output=json list-hosted-zones | jq '.HostedZones[] | select(.Name=="example.org.") | .Id' | cut -d/ -f3|cut -d\" -f1
```

It's a good idea if we export this ID as a shell variable by using the following command:

```bash
export parentzoneid=`aws route53 --output=json list-hosted-zones | jq '.HostedZones[] | select(.Name=="example.org.") | .Id' | cut -d/ -f3|cut -d\" -f1`
```

Let's check the var:

```bash
echo $parentzoneid
ZTKK4EXR1EWR5
```

With the name servers obtained above, we need to construct a "json" file that we'll pass to amazon for our subdomain:

```bash
cat<<EOF >~/kopsclustertest.example.org.json
{
  "Comment": "Create a subdomain NS record in the parent domain",
  "Changes": [
    {
      "Action": "CREATE",
      "ResourceRecordSet": {
        "Name": "kopsclustertest.example.org",
        "Type": "NS",
        "TTL": 300,
        "ResourceRecords": [
          {
            "Value": "ns-1383.awsdns-44.org"
          },
          {
            "Value": "ns-829.awsdns-39.net"
          },
          {
            "Value": "ns-346.awsdns-43.com"
          },
          {
            "Value": "ns-1973.awsdns-54.co.uk"
          }
        ]
      }
    }
  ]
}
EOF

```

**NOTE:** This step is needed because the subdomain was created, but it does not have "ns" records on it. We are basically adding four NS records to the subdomain here.

With the json file ready, and the parent zone ID exported in the "$parentzoneid" environment variable, we can finish the task and add the NS records to the subdomain using the following command:

```bash
aws route53 change-resource-record-sets \
--output=table \
--hosted-zone-id $parentzoneid \
--change-batch file://~/kopsclustertest.example.org.json
```

The output of the last command will be something like:

```
-------------------------------------------------------------------------------------------------------------------------
|                                               ChangeResourceRecordSets                                                |
+-----------------------------------------------------------------------------------------------------------------------+
||                                                     ChangeInfo                                                      ||
|+----------------------------------------------------+------------------------+----------+----------------------------+|
||                       Comment                      |          Id            | Status   |        SubmittedAt         ||
|+----------------------------------------------------+------------------------+----------+----------------------------+|
||  Create a subdomain NS record in the parent domain |  /change/CJ7FOVJ7U58L0 |  PENDING |  2017-09-06T13:28:12.972Z  ||
|+----------------------------------------------------+------------------------+----------+----------------------------+|
```

Finally, check your records with the following command:

```bash
aws route53 list-resource-record-sets \
--output=table \
--hosted-zone-id `aws route53 --output=json list-hosted-zones | jq '.HostedZones[] | select(.Name=="kopsclustertest.example.org.") | .Id' | cut -d/ -f3|cut -d\" -f1`
```

The last command will output the following info:

```bash
---------------------------------------------------------------------------------------
|                               ListResourceRecordSets                                |
+-------------------------------------------------------------------------------------+
||                                ResourceRecordSets                                 ||
|+----------------------------------------------------+----------------+-------------+|
||                        Name                        |      TTL       |    Type     ||
|+----------------------------------------------------+----------------+-------------+|
||  kopsclustertest.example.org.                       |  172800        |  NS         ||
|+----------------------------------------------------+----------------+-------------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  ns-1383.awsdns-44.org.                                                         |||
|||  ns-829.awsdns-39.net.                                                          |||
|||  ns-346.awsdns-43.com.                                                          |||
|||  ns-1973.awsdns-54.co.uk.                                                       |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+-------------------------------------------------------+------------+--------------+|
||                         Name                          |    TTL     |    Type      ||
|+-------------------------------------------------------+------------+--------------+|
||  kopsclustertest.example.org.                          |  900       |  SOA         ||
|+-------------------------------------------------------+------------+--------------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  ns-1383.awsdns-44.org. awsdns-hostmaster.amazon.com. 1 7200 900 1209600 86400  |||
||+---------------------------------------------------------------------------------+||
```

Also, do a "dig" test in order to check the zone availability on the Internet:

```bash
dig +short kopsclustertest.example.org soa

ns-1383.awsdns-44.org. awsdns-hostmaster.amazon.com. 1 7200 900 1209600 86400

dig +short kopsclustertest.example.org ns

ns-1383.awsdns-44.org.
ns-829.awsdns-39.net.
ns-1973.awsdns-54.co.uk.
ns-346.awsdns-43.com.
```

If both your SOA and NS records are there, then your subdomain is ready to be used by KOPS.


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

Then export the name of your cluster along with the "S3" URL of your bucket. Add your cluster name to the full subdomain:

```bash
export NAME=mycluster01.kopsclustertest.example.org
export KOPS_STATE_STORE=s3://my-kops-s3-bucket-for-cluster-state
```

Some things to note from here:

- "NAME" will be an environment variable that we'll use from now in order to refer to our cluster name. For this practical exercise, our cluster name will be "mycluster01.kopsclustertest.example.org".


## KOPS CLUSTER CREATION:

Let's first create our cluster ensuring a multi-master setup with 3 masters in a multi-az setup, two worker nodes also in a multi-az setup, and using both private networking and a bastion server:

```bash
kops create cluster \
--cloud=aws \
--master-zones=us-east-1a,us-east-1b,us-east-1c \
--zones=us-east-1a,us-east-1b,us-east-1c \
--node-count=2 \
--node-size=t2.micro \
--master-size=t2.micro \
${NAME}
```

A few things to note here:

- The environment variable ${NAME} was previously exported with our cluster name: mycluster01.kopsclustertest.example.org.
- "--cloud=aws": As kops grows and begin to support more clouds, we need to tell the command to use the specific cloud we want for our deployment. In this case: amazon web services (aws).
- For true HA at the master level, we need to pick a region with at least 3 availability zones. For this practical exercise, we are using "us-east-1" AWS region which contains 5 availability zones (az's for short): us-east-1a, us-east-1b, us-east-1c, us-east-1d and us-east-1e.
- The "--master-zones=us-east-1a,us-east-1b,us-east-1c" KOPS argument will actually enforce that we want 3 masters here. "--node-count=2" only applies to the worker nodes (not the masters).
- We are including the arguments "--node-size" and "master-size" to specify the "instance types" for both our masters and worker nodes.
- Because we are just doing a simple LAB, we are using "t2.micro" machines. Please DON'T USE t2.micro on real production systems. Start with "t2.medium" as a minimum realistic/workable machine type.

With those points clarified, let's deploy our cluster:

```bash
kops update cluster ${NAME} --yes
```

The last command will generate the following output:

```bash
I0906 09:42:09.399908   13538 executor.go:91] Tasks: 0 done / 75 total; 38 can run
I0906 09:42:12.033675   13538 vfs_castore.go:422] Issuing new certificate: "master"
I0906 09:42:12.310586   13538 vfs_castore.go:422] Issuing new certificate: "kube-scheduler"
I0906 09:42:12.791469   13538 vfs_castore.go:422] Issuing new certificate: "kube-proxy"
I0906 09:42:13.312675   13538 vfs_castore.go:422] Issuing new certificate: "kops"
I0906 09:42:13.378500   13538 vfs_castore.go:422] Issuing new certificate: "kubelet"
I0906 09:42:13.398070   13538 vfs_castore.go:422] Issuing new certificate: "kube-controller-manager"
I0906 09:42:13.636134   13538 vfs_castore.go:422] Issuing new certificate: "kubecfg"
I0906 09:42:14.684945   13538 executor.go:91] Tasks: 38 done / 75 total; 14 can run
I0906 09:42:15.997588   13538 executor.go:91] Tasks: 52 done / 75 total; 19 can run
I0906 09:42:17.855959   13538 launchconfiguration.go:327] waiting for IAM instance profile "masters.mycluster01.kopsclustertest.example.org" to be ready
I0906 09:42:17.932515   13538 launchconfiguration.go:327] waiting for IAM instance profile "nodes.mycluster01.kopsclustertest.example.org" to be ready
I0906 09:42:18.602180   13538 launchconfiguration.go:327] waiting for IAM instance profile "masters.mycluster01.kopsclustertest.example.org" to be ready
I0906 09:42:18.682038   13538 launchconfiguration.go:327] waiting for IAM instance profile "masters.mycluster01.kopsclustertest.example.org" to be ready
I0906 09:42:29.215995   13538 executor.go:91] Tasks: 71 done / 75 total; 4 can run
I0906 09:42:30.073417   13538 executor.go:91] Tasks: 75 done / 75 total; 0 can run
I0906 09:42:30.073471   13538 dns.go:152] Pre-creating DNS records
I0906 09:42:32.403909   13538 update_cluster.go:247] Exporting kubecfg for cluster
Kops has set your kubectl context to mycluster01.kopsclustertest.example.org

Cluster is starting.  It should be ready in a few minutes.

Suggestions:
 * validate cluster: kops validate cluster
 * list nodes: kubectl get nodes --show-labels
 * ssh to the master: ssh -i ~/.ssh/id_rsa admin@api.mycluster01.kopsclustertest.example.org
The admin user is specific to Debian. If not using Debian please use the appropriate user based on your OS.
 * read about installing addons: https://github.com/kubernetes/kops/blob/master/docs/operations/addons.md
```

Note that KOPS will create a DNS record for your API: api.mycluster01.kopsclustertest.example.org. You can check this record with the following "dig" command:

```bash
dig +short api.mycluster01.kopsclustertest.example.org A
34.228.219.212
34.206.72.126
54.83.144.111
```

KOPS created a DNS round-robin resource record with all the public IP's assigned to the masters. Do you remember we specified 3 masters ?. Well, there are their IP's.

After about 10~15 minutes (depending on how fast or how slow are amazon services during the cluster creation) you can check your cluster:

```bash
kops validate cluster

Using cluster from kubectl context: mycluster01.kopsclustertest.example.org

Validating cluster mycluster01.kopsclustertest.example.org

INSTANCE GROUPS
NAME                    ROLE    MACHINETYPE     MIN     MAX     SUBNETS
master-us-east-1a       Master  t2.micro        1       1       us-east-1a
master-us-east-1b       Master  t2.micro        1       1       us-east-1b
master-us-east-1c       Master  t2.micro        1       1       us-east-1c
nodes                   Node    t2.micro        2       2       us-east-1a,us-east-1b,us-east-1c

NODE STATUS
NAME                            ROLE    READY
ip-172-20-125-42.ec2.internal   master  True
ip-172-20-33-58.ec2.internal    master  True
ip-172-20-43-160.ec2.internal   node    True
ip-172-20-64-116.ec2.internal   master  True
ip-172-20-68-15.ec2.internal    node    True

Your cluster mycluster01.kopsclustertest.example.org is ready

```

Also with "kubectl":

```bash
kubectl get nodes

NAME                            STATUS    AGE       VERSION
ip-172-20-125-42.ec2.internal   Ready     6m        v1.7.2
ip-172-20-33-58.ec2.internal    Ready     6m        v1.7.2
ip-172-20-43-160.ec2.internal   Ready     5m        v1.7.2
ip-172-20-64-116.ec2.internal   Ready     6m        v1.7.2
ip-172-20-68-15.ec2.internal    Ready     5m        v1.7.2
```

Let's try to send a command to our masters using "ssh":

```bash
ssh -i ~/.ssh/id_rsa admin@api.mycluster01.kopsclustertest.example.org "ec2metadata --public-ipv4"
34.206.72.126
```

Our "api.xxxx" resource record is working OK.

## DNS RESOURCE RECORDS CREATED BY KOPS ON ROUTE 53

Let's do a fast review (using aws cli tools) of the resource records created by KOPS inside our subdomain:

```bash
aws route53 list-resource-record-sets \
--output=table \
--hosted-zone-id `aws route53 --output=json list-hosted-zones | jq '.HostedZones[] | select(.Name=="kopsclustertest.example.org.") | .Id' | cut -d/ -f3|cut -d\" -f1`
```

The output:

```
---------------------------------------------------------------------------------------
|                               ListResourceRecordSets                                |
+-------------------------------------------------------------------------------------+
||                                ResourceRecordSets                                 ||
|+----------------------------------------------------+----------------+-------------+|
||                        Name                        |      TTL       |    Type     ||
|+----------------------------------------------------+----------------+-------------+|
||  kopsclustertest.example.org.                       |  172800        |  NS         ||
|+----------------------------------------------------+----------------+-------------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  ns-1383.awsdns-44.org.                                                         |||
|||  ns-829.awsdns-39.net.                                                          |||
|||  ns-346.awsdns-43.com.                                                          |||
|||  ns-1973.awsdns-54.co.uk.                                                       |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+-------------------------------------------------------+------------+--------------+|
||                         Name                          |    TTL     |    Type      ||
|+-------------------------------------------------------+------------+--------------+|
||  kopsclustertest.example.org.                          |  900       |  SOA         ||
|+-------------------------------------------------------+------------+--------------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  ns-1383.awsdns-44.org. awsdns-hostmaster.amazon.com. 1 7200 900 1209600 86400  |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+--------------------------------------------------------------+---------+----------+|
||                             Name                             |   TTL   |  Type    ||
|+--------------------------------------------------------------+---------+----------+|
||  api.mycluster01.kopsclustertest.example.org.                 |  60     |  A       ||
|+--------------------------------------------------------------+---------+----------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  34.206.72.126                                                                  |||
|||  34.228.219.212                                                                 |||
|||  54.83.144.111                                                                  |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+-----------------------------------------------------------------+-------+---------+|
||                              Name                               |  TTL  |  Type   ||
|+-----------------------------------------------------------------+-------+---------+|
||  api.internal.mycluster01.kopsclustertest.example.org.           |  60   |  A      ||
|+-----------------------------------------------------------------+-------+---------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  172.20.125.42                                                                  |||
|||  172.20.33.58                                                                   |||
|||  172.20.64.116                                                                  |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+------------------------------------------------------------------+-------+--------+|
||                               Name                               |  TTL  | Type   ||
|+------------------------------------------------------------------+-------+--------+|
||  etcd-a.internal.mycluster01.kopsclustertest.example.org.         |  60   |  A     ||
|+------------------------------------------------------------------+-------+--------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  172.20.33.58                                                                   |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+------------------------------------------------------------------+-------+--------+|
||                               Name                               |  TTL  | Type   ||
|+------------------------------------------------------------------+-------+--------+|
||  etcd-b.internal.mycluster01.kopsclustertest.example.org.         |  60   |  A     ||
|+------------------------------------------------------------------+-------+--------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  172.20.64.116                                                                  |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+------------------------------------------------------------------+-------+--------+|
||                               Name                               |  TTL  | Type   ||
|+------------------------------------------------------------------+-------+--------+|
||  etcd-c.internal.mycluster01.kopsclustertest.example.org.         |  60   |  A     ||
|+------------------------------------------------------------------+-------+--------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  172.20.125.42                                                                  |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+-------------------------------------------------------------------+------+--------+|
||                               Name                                | TTL  | Type   ||
|+-------------------------------------------------------------------+------+--------+|
||  etcd-events-a.internal.mycluster01.kopsclustertest.example.org.   |  60  |  A     ||
|+-------------------------------------------------------------------+------+--------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  172.20.33.58                                                                   |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+-------------------------------------------------------------------+------+--------+|
||                               Name                                | TTL  | Type   ||
|+-------------------------------------------------------------------+------+--------+|
||  etcd-events-b.internal.mycluster01.kopsclustertest.example.org.   |  60  |  A     ||
|+-------------------------------------------------------------------+------+--------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  172.20.64.116                                                                  |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+-------------------------------------------------------------------+------+--------+|
||                               Name                                | TTL  | Type   ||
|+-------------------------------------------------------------------+------+--------+|
||  etcd-events-c.internal.mycluster01.kopsclustertest.example.org.   |  60  |  A     ||
|+-------------------------------------------------------------------+------+--------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  172.20.125.42                                                                  |||
||+---------------------------------------------------------------------------------+||
```

Maybe with json output and some "jq" parsing:

```bash
aws route53 list-resource-record-sets --output=json --hosted-zone-id `aws route53 --output=json list-hosted-zones | jq '.HostedZones[] | select(.Name=="kopsclustertest.example.org.") | .Id' | cut -d/ -f3|cut -d\" -f1`|jq .ResourceRecordSets[]
```

Output:

```
{
  "TTL": 172800,
  "Name": "kopsclustertest.example.org.",
  "Type": "NS",
  "ResourceRecords": [
    {
      "Value": "ns-1383.awsdns-44.org."
    },
    {
      "Value": "ns-829.awsdns-39.net."
    },
    {
      "Value": "ns-346.awsdns-43.com."
    },
    {
      "Value": "ns-1973.awsdns-54.co.uk."
    }
  ]
}
{
  "TTL": 900,
  "Name": "kopsclustertest.example.org.",
  "Type": "SOA",
  "ResourceRecords": [
    {
      "Value": "ns-1383.awsdns-44.org. awsdns-hostmaster.amazon.com. 1 7200 900 1209600 86400"
    }
  ]
}
{
  "TTL": 60,
  "Name": "api.mycluster01.kopsclustertest.example.org.",
  "Type": "A",
  "ResourceRecords": [
    {
      "Value": "34.206.72.126"
    },
    {
      "Value": "34.228.219.212"
    },
    {
      "Value": "54.83.144.111"
    }
  ]
}
{
  "TTL": 60,
  "Name": "api.internal.mycluster01.kopsclustertest.example.org.",
  "Type": "A",
  "ResourceRecords": [
    {
      "Value": "172.20.125.42"
    },
    {
      "Value": "172.20.33.58"
    },
    {
      "Value": "172.20.64.116"
    }
  ]
}
{
  "TTL": 60,
  "Name": "etcd-a.internal.mycluster01.kopsclustertest.example.org.",
  "Type": "A",
  "ResourceRecords": [
    {
      "Value": "172.20.33.58"
    }
  ]
}
{
  "TTL": 60,
  "Name": "etcd-b.internal.mycluster01.kopsclustertest.example.org.",
  "Type": "A",
  "ResourceRecords": [
    {
      "Value": "172.20.64.116"
    }
  ]
}
{
  "TTL": 60,
  "Name": "etcd-c.internal.mycluster01.kopsclustertest.example.org.",
  "Type": "A",
  "ResourceRecords": [
    {
      "Value": "172.20.125.42"
    }
  ]
}
{
  "TTL": 60,
  "Name": "etcd-events-a.internal.mycluster01.kopsclustertest.example.org.",
  "Type": "A",
  "ResourceRecords": [
    {
      "Value": "172.20.33.58"
    }
  ]
}
{
  "TTL": 60,
  "Name": "etcd-events-b.internal.mycluster01.kopsclustertest.example.org.",
  "Type": "A",
  "ResourceRecords": [
    {
      "Value": "172.20.64.116"
    }
  ]
}
{
  "TTL": 60,
  "Name": "etcd-events-c.internal.mycluster01.kopsclustertest.example.org.",
  "Type": "A",
  "ResourceRecords": [
    {
      "Value": "172.20.125.42"
    }
  ]
}
```

## SCALING-UP YOUR CLUSTER.

Let's see the following scenario: Our load is increasing and we need to add two more worker nodes. First, let's get our instance group names:

```bash
kops get instancegroups
Using cluster from kubectl context: mycluster01.kopsclustertest.example.org

NAME                    ROLE    MACHINETYPE     MIN     MAX     SUBNETS
master-us-east-1a       Master  t2.micro        1       1       us-east-1a
master-us-east-1b       Master  t2.micro        1       1       us-east-1b
master-us-east-1c       Master  t2.micro        1       1       us-east-1c
nodes                   Node    t2.micro        2       2       us-east-1a,us-east-1b,us-east-1c
```

We can see here that our workers instance group name is "nodes". Let's edit the group with the command "kops edit ig nodes"

```bash
kops edit ig nodes
```

An editor (whatever you have on the $EDITOR shell variable) will open with the following text:

```
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-09-06T13:40:39Z
  labels:
    kops.k8s.io/cluster: mycluster01.kopsclustertest.example.org
  name: nodes
spec:
  image: kope.io/k8s-1.7-debian-jessie-amd64-hvm-ebs-2017-07-28
  machineType: t2.micro
  maxSize: 2
  minSize: 2
  role: Node
  subnets:
  - us-east-1a
  - us-east-1b
  - us-east-1c
```

Let's change minSize and maxSize to "3"

```
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: 2017-09-06T13:40:39Z
  labels:
    kops.k8s.io/cluster: mycluster01.kopsclustertest.example.org
  name: nodes
spec:
  image: kope.io/k8s-1.7-debian-jessie-amd64-hvm-ebs-2017-07-28
  machineType: t2.micro
  maxSize: 3
  minSize: 3
  role: Node
  subnets:
  - us-east-1a
  - us-east-1b
  - us-east-1c
```

Save it and review with `kops update cluster $NAME`:

```bash
kops update cluster $NAME
```

The last command will output:

```bash
I0906 10:16:30.619321   13607 executor.go:91] Tasks: 0 done / 75 total; 38 can run
I0906 10:16:32.703865   13607 executor.go:91] Tasks: 38 done / 75 total; 14 can run
I0906 10:16:33.592807   13607 executor.go:91] Tasks: 52 done / 75 total; 19 can run
I0906 10:16:35.009432   13607 executor.go:91] Tasks: 71 done / 75 total; 4 can run
I0906 10:16:35.320078   13607 executor.go:91] Tasks: 75 done / 75 total; 0 can run
Will modify resources:
  AutoscalingGroup/nodes.mycluster01.kopsclustertest.example.org
        MinSize                  2 -> 3
        MaxSize                  2 -> 3

Must specify --yes to apply changes
```

Now, let's apply the change:

```bash
kops update cluster $NAME --yes
```

Go for another coffee (or maybe a tee) and after some minutes check your cluster again with "kops validate cluster"

```bash
kops validate cluster

Using cluster from kubectl context: mycluster01.kopsclustertest.example.org

Validating cluster mycluster01.kopsclustertest.example.org

INSTANCE GROUPS
NAME                    ROLE    MACHINETYPE     MIN     MAX     SUBNETS
master-us-east-1a       Master  t2.micro        1       1       us-east-1a
master-us-east-1b       Master  t2.micro        1       1       us-east-1b
master-us-east-1c       Master  t2.micro        1       1       us-east-1c
nodes                   Node    t2.micro        3       3       us-east-1a,us-east-1b,us-east-1c

NODE STATUS
NAME                            ROLE    READY
ip-172-20-103-68.ec2.internal   node    True
ip-172-20-125-42.ec2.internal   master  True
ip-172-20-33-58.ec2.internal    master  True
ip-172-20-43-160.ec2.internal   node    True
ip-172-20-64-116.ec2.internal   master  True
ip-172-20-68-15.ec2.internal    node    True

Your cluster mycluster01.kopsclustertest.example.org is ready

```

You can see how your cluster scaled up to 3 nodes.

**SCALING RECOMMENDATIONS:**
- Always think ahead. If you want to ensure to have the capability to scale-up to all available zones in the region, ensure to add them to the "--zones=" argument when using the "kops create cluster" command. Example: --zones=us-east-1a,us-east-1b,us-east-1c,us-east-1d,us-east-1e. That will make things simpler later.
- For the masters, always consider "odd" numbers starting from 3. Like many other cluster, odd numbers starting from "3" are the proper way to create a fully redundant multi-master solution. In the specific case of "kops", you add masters by adding zones to the "--master-zones" argument on "kops create command".

## DELETING OUR CLUSTER AND CHECKING OUR DNS SUBDOMAIN:

If we don't need our cluster anymore, let's use a kops command in order to delete it:

```bash
kops delete cluster ${NAME} --yes
```

After a short while, you'll see the following message:

```
Deleted kubectl config for mycluster01.kopsclustertest.example.org

Deleted cluster: "mycluster01.kopsclustertest.example.org"
```

Now, let's check our DNS records:

```bash
aws route53 list-resource-record-sets \
--output=table \
--hosted-zone-id `aws route53 --output=json list-hosted-zones | jq '.HostedZones[] | select(.Name=="kopsclustertest.example.org.") | .Id' | cut -d/ -f3|cut -d\" -f1`
```

The output:

```
---------------------------------------------------------------------------------------
|                               ListResourceRecordSets                                |
+-------------------------------------------------------------------------------------+
||                                ResourceRecordSets                                 ||
|+----------------------------------------------------+----------------+-------------+|
||                        Name                        |      TTL       |    Type     ||
|+----------------------------------------------------+----------------+-------------+|
||  kopsclustertest.example.org.                       |  172800        |  NS         ||
|+----------------------------------------------------+----------------+-------------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  ns-1383.awsdns-44.org.                                                         |||
|||  ns-829.awsdns-39.net.                                                          |||
|||  ns-346.awsdns-43.com.                                                          |||
|||  ns-1973.awsdns-54.co.uk.                                                       |||
||+---------------------------------------------------------------------------------+||
||                                ResourceRecordSets                                 ||
|+-------------------------------------------------------+------------+--------------+|
||                         Name                          |    TTL     |    Type      ||
|+-------------------------------------------------------+------------+--------------+|
||  kopsclustertest.example.org.                          |  900       |  SOA         ||
|+-------------------------------------------------------+------------+--------------+|
|||                                 ResourceRecords                                 |||
||+---------------------------------------------------------------------------------+||
|||                                      Value                                      |||
||+---------------------------------------------------------------------------------+||
|||  ns-1383.awsdns-44.org. awsdns-hostmaster.amazon.com. 1 7200 900 1209600 86400  |||
||+---------------------------------------------------------------------------------+||
```

All kops-created resource records are deleted too. Only the NS records added by us are still there.

END.-
