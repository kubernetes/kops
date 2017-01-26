<p align="center">
  <img src="img/k8s-aws.png"> </image>
</p>



# Getting Started

## Install kops

From Homebrew:

```bash
brew update && brew install --HEAD kops
```

From Source:

```bash
go get -d k8s.io/kops
cd ${GOPATH}/src/k8s.io/kops/
git checkout release
make
```

See our [installation guide](build.md) for more information

## Install kubectl

It is a good idea to grab a fresh copy of `kubectl` now if you don't already have it.

#### OS X

```
brew install kubernetes-cli
```

#### Other Platforms

* [Kubernetes Latest Release](https://github.com/kubernetes/kubernetes/releases/latest)

* [Installation Guide](http://kubernetes.io/docs/user-guide/prereqs/)


## Setup your environment

### Setting up a kops IAM user


In this example we will be using a dedicated IAM user to use with kops. This user will need basic API security credentials in order to use kops. Create the user and credentials using the AWS console. [More information](http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSGettingStartedGuide/AWSCredentials.html).

Kubernetes kops uses the official AWS Go SDK, so all we need to do here is set up your system to use the official AWS supported methods of registering security credentials defined [here](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials). Here is an example using the aws command line tool to set up your security credentials.

#### OS X

##### Installing aws cli

The officially supported way of installing the tool is with `pip` as in

```bash
pip install awscli
```

You can also grab the tool with homebrew, although this is not officially supported.

```bash
brew update && brew install awscli
```

Now configure the tool, and verify it works.

```bash
aws configure # Input your credentials here
aws iam list-users
```

PyPi is the officially supported `aws cli` download avenue, and kops suggests using it. [More information](https://pypi.python.org/pypi/awscli) on the package.

#### Other Platforms

Official documentation [here](http://docs.aws.amazon.com/cli/latest/userguide/installing.html)

We should now be able to pull a list of IAM users from the API, verifying that our credentials are working as expected.

## Configure DNS

We will now need to set up DNS for cluster, find one of the scenarios below (A,B,C) that match your situation.

### (A) Setting up DNS for your cluster, with AWS as your registrar

If you bought your domain with AWS, then you should already have a hosted zone in Route53.

If you plan on using your base domain, then no more work is needed. 

#### Setting up a subdomain

If you plan on using a subdomain to build your clusters on you will need to create a 2nd hosted zone in Route53, and then set up route delegation. This is basically copying the NS servers of your **SUBDOMAIN** up to the **PARENT** domain in Route53. 

  - Create the subdomain, and note your **SUBDOMAIN** name servers (If you have already done this you can also [get the values](ns.md))

```bash
ID=$(uuidgen) && aws route53 create-hosted-zone --name subdomain.kubernetes.com --caller-reference $ID | jq .DelegationSet.NameServers
```

  - Note your **PARENT** hosted zone id

```bash
aws route53 list-hosted-zones | jq '.HostedZones[] | select(.Name=="kubernetes.com.") | .Id' 
```

 - Create a new JSON file with your values (`subdomain.json`)
 
 Note: The NS values here are for the **SUBDOMAIN**
 
 ```
 {
   "Comment": "Create a subdomain NS record in the parent domain",
   "Changes": [
     {
       "Action": "CREATE",
       "ResourceRecordSet": {
         "Name": "subdomain.kubernetes.com",
         "Type": "NS",
         "TTL": 300,
         "ResourceRecords": [
           {
             "Value": "ns-1.awsdns-1.co.uk"
           },
           {
             "Value": "ns-2.awsdns-2.org"
           },
           {
             "Value": "ns-3.awsdns-3.com"
           },
           {
             "Value": "ns-4.awsdns-4.net"
           }
         ]
       }
     }
   ]
 }
 ```
 
 - Apply the **SUBDOMAIN** NS records to the **PARENT** hosted zone
 
 ```
 aws route53 change-resource-record-sets \
  --hosted-zone-id <parent-zone-id> \
  --change-batch file://subdomain.json
```

Now traffic to `*.kubernetes.com` will be routed to the correct subdomain hosted zone in Route 53.

### (B) Setting up DNS for your cluster, with another registrar.

If you bought your domain elsewhere, and would like to dedicate the entire domain to AWS you should follow the guide [here](http://docs.aws.amazon.com/Route53/latest/DeveloperGuide/domain-transfer-to-route-53.html)

### (C) Setting up a subdomain for clusters, with another registrar while keeping your top level domain the same

If you bought your domain elsewhere, but **only want to use a subdomain in AWS Route53** you must modify your registrar's NS (NameServer) records. See the example below.

Here we will be creating a hosted zone in AWS Route53, and migrating the subdomain's NS records to your other registrar.

You might need to grab [jq](https://github.com/stedolan/jq/wiki/Installation) for some of these.

  - Create the subdomain, and note your name servers (If you have already done this you can also [get the values](ns.md))

```bash
ID=$(uuidgen) && aws route53 create-hosted-zone --name subdomain.kubernetes.com --caller-reference $ID | jq .DelegationSet.NameServers
```

 - You will now go to your registrars page and log in. You will need to create a new **SUBDOMAIN**, and use the 4 NS records listed above for the new **SUBDOMAIN**. This **MUST** be done in order to use your cluster. Do **NOT** change your top level NS record, or you might take your site offline.

 - Information on adding NS records with [Godaddy.com](https://www.godaddy.com/help/set-custom-nameservers-for-domains-registered-with-godaddy-12317)
 - Information on adding NS records with [Google Cloud Platform](https://cloud.google.com/dns/update-name-servers)

#### Using Public/Private DNS (1.5+)

Kops by default will assume that the NS records created above are publicly available. If the values above are not publicly available, kops will have undesired results.

Note: There is a DNS flag that can be configured if you plan on using private DNS records

```
kops create cluster --dns private $NAME
```

## Testing your DNS setup

You should now able to dig your domain (or subdomain) and see the AWS Name Servers on the other end.

```bash
dig ns subdomain.kubernetes.com
```

```
;; ANSWER SECTION:
subdomain.kubernetes.com.        172800  IN  NS  ns-1.awsdns-1.net.
subdomain.kubernetes.com.        172800  IN  NS  ns-2.awsdns-2.org.
subdomain.kubernetes.com.        172800  IN  NS  ns-3.awsdns-3.com.
subdomain.kubernetes.com.        172800  IN  NS  ns-4.awsdns-4.co.uk.
```

This is a critical component of setting up the cluster. If you are experiencing problems with the Kubernetes API not coming up, chances are something is amiss around DNS. 

**Please DO NOT MOVE ON until you have validated your NS records!**

## Setting up a state store for your cluster

In this example we will be creating a dedicated S3 bucket for kops to use. This is where kops will store the state of your cluster and the representation of your cluster, and serves as the source of truth for our cluster configuration throughout the process. We will call this kubernetes-com-state-store. We recommend keeping the creation confined to us-east-1, otherwise more input will be needed here.

```bash
aws s3api create-bucket --bucket kubernetes-com-state-store --region us-east-1
```

Note: We **STRONGLY** recommend versioning your S3 bucket in case you ever need to revert or recover a previous state store.

## Creating your first cluster

#### Setup your environment for kops

Okay! We are ready to start creating our first cluster. Lets first set up a few environmental variables to make this process as clean as possible.

```bash
export NAME=myfirstcluster.kubernetes.com
export KOPS_STATE_STORE=s3://kubernetes-com-state-store
```

Note: You don’t have to use environmental variables here. You can always define the values using the –name and –state flags later.

#### Form your create cluster command

We will need to note which availability zones are available to us. In this example we will be deploying our cluster to the us-west-2 region.

```bash
aws ec2 describe-availability-zones --region us-west-2
```

Lets form our create cluster command. This is the most basic example, a more verbose example on can be found [here](advanced_create.md)

```bash
kops create cluster \
    --zones us-west-2a \
    ${NAME}
```

kops will deploy these instances using AWS auto scaling groups, so each instance should be ephemeral and will rebuild itself if taken offline for any reason.

#### Cluster Configuration

We now have created the underlying cluster configuration, lets take a look at every aspect that will define our cluster.

```bash
kops edit cluster ${NAME}
```

This will open in your text editor of choice. You can always change your editor of choice

```bash
cat "export EDITOR=/usr/bin/emacs" ~/.bash_profile && source ~/.bash_profile
```

This will open up the cluster config (that is actually stored in the S3 bucket we created earlier!) in your favorite text editor. Here is where we can optionally really tweak our cluster for our use case. In this tutorial, we leave it default for now.

#### Apply the changes

```bash
kops update cluster ${NAME} --yes
```


## Accessing your cluster

A friendly reminder that kops runs asynchronously, and it will take your cluster a few minutes to come online.

Remember when you installed `kubectl` earlier? The configuration for your cluster was automatically generated and written to `~/.kube/config` for you!

A simple Kubernetes API call can be used to check if the API is online and listening. Let's use `kubectl`

```bash
kubectl get nodes
```

You will see a list of nodes that should match the `--zones` flag defined earlier. This is a great sign that your Kubernetes cluster is online and working.

Also kops ships with a handy validation tool that can be ran to ensure your cluster is working as expected.

```bash
kops validate cluster
```

Another great one liner

```
kubectl -n kube-system get po
```

## What's next?

Kops has a ton of great features, and an amazing support team. We recommend researching [other interesting modes](commands.md#other-interesting-modes) to learn more about generating Terraform configurations, or running your cluster in HA (Highly Available). You might want to take a peek at the [cluster spec docs](cluster_spec.md) for helping to configure these "other interesting modes". Also be sure to check out how to run a [private network topology](topology.md) in AWS.



Explore the program, and work on getting your `cluster config` hammered out!

## Feedback

We love feedback from the community, and if you are reading this we would love to hear from you and get your thoughts. Read more about [getting involved](https://github.com/kubernetes/kops/blob/master/README.md#getting-involved) to find out how to track us down.


###### Legal

*AWS Trademark used with limited permission under the [AWS Trademark Guidelines](https://aws.amazon.com/trademark-guidelines/)*

*Kubernetes Logo used with permission under the [Kubernetes Branding Guidelines](https://github.com/kubernetes/kubernetes/blob/master/logo/usage_guidelines.md)*
