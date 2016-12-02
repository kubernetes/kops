# Kubernetes on AWS

<p align="center">
  <img src="img/k8s-aws.png"> </image>
</p>



## Getting Started

#### kops

From Homebrew:

```
brew install kops
```

From Source:

```
go get -d k8s.io/kops
cd ${GOPATH}/src/k8s.io/kops/
git checkout release
make
```

#### kubectl

It is a good idea to grab a fresh copy of `kubectl` now if you don't already have it.

[Kubernetes Latest Release](https://github.com/kubernetes/kubernetes/releases/latest)
[Installation Guide](http://kubernetes.io/docs/user-guide/prereqs/)

```
wget -O https://github.com/kubernetes/kubernetes/releases/download/v1.4.6/kubernetes.tar.gz
sudo cp kubernetes/platforms/darwin/amd64/kubectl /usr/local/bin/kubectl
```


See our [installation guide](build.md) for more information


#### Your environment

1) Set up a DNS hosted zone in Route 53, e.g. `mydomain.com`, and set up the DNS nameservers as normal so that domains will resolve.  You can reuse an existing domain name (e.g. `mydomain.com`), or you can create a "child" hosted zone (e.g. `myclusters.mydomain.com`) if you want to isolate them.

**Note**: that with AWS Route53, you can have subdomains in a single hosted zone, so you can have `cluster1.testclusters.mydomain.com` under `mydomain.com`.

2) Pick a DNS name under this zone to be the name of your cluster.  kops will set up DNS so your cluster can be reached on this name.  For example, if your zone was `mydomain.com`, a good name would be `kubernetes.mydomain.com`, or `dev.k8s.mydomain.com`, or even `dev.k8s.myproject.mydomain.com`. We'll call this `NAME`.

3) Kops uses the [Official AWS Go SDK](https://github.com/aws/aws-sdk-go). You will need to set up your AWS credentials as defined [here](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)

4) Pick an S3 bucket that you'll use to store your cluster configuration - this is called your state store.  You can `export KOPS_STATE_STORE=s3://<mystatestorebucket>` and then kops will use this location by default.  We suggest putting this in your bash profile or similar.  A single registry can hold multiple clusters, and it can also be shared amongst your ops team (which is much easier than passing around kubecfg files!)


#### Create your first cluster

Before we create a cluster, we need to generate a `cluster config`

In these examples we assume you have already exported `KOPS_STATE_STORE`, otherwise you will need to append each command with `--state s3://<mystatestorebucket>`

```
kops create cluster --zones=us-east-1c ${NAME}
```

Notice in this example we defined a single AWS Availability Zone. These represent the nodes in your cluster, that are typically recommended to run each in their own Availability Zone. Best practices dictate uses more than one here.

We can customize our cluster by editing our `cluster config`

```
kops edit cluster ${NAME}
```

This will open up the cluster manifest YAML file in your favorite text editor. Here is where a user can define very specific parts of their cluster. For now, lets leave this as default.


Lets go ahead and create the cluster in AWS!

```
kops update cluster ${NAME} --yes
```

Think of the `--yes` flag as a way of saying "*Yes! I am very sure I want to create this cluster, and I understand it will cost me money*". You will notice this flag in a few other places as well.

# Accessing your cluster

A friendly reminder that kops runs asynchronously, and it will take your cluster a few minutes to come online.

Remember when you installed `kubectl` earlier? The configuration for your cluster was automatically generated and written to `~/.kube/config` for you!

A simple Kubernetes API call can be used to check if the API is online and listening. Let's use `kubectl`

```
kubectl get nodes
```

You will see a list of nodes that should match the `--zones` flag defined earlier. This is a great sign that your Kubernetes cluster is online and working.

# What's next?

Kops has a ton of great features, and an amazing support team. We recommend researching [other interesting modes](commands.md#other-interesting-modes) to learn more about generating Terraform configurations, or running your cluster in HA (Highly Available). Also be sure to check out how to run a[private network topology](topology.md) in AWS.

Explore the program, and work on getting your `cluster config` hammered out!

# Feedback

We love feedback from the community, and if you are reading this we would love to hear from you and get your thoughts. Read more about [getting involved](https://github.com/kubernetes/kops/blob/master/README.md#getting-involved) to find out how to track us down.