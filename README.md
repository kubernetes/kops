## Kops - Kubernetes Ops

kops is the easiest way to get a production Kubernetes up and running.  We like to think
of it as "kubectl for clusters".

(Currently work in progress, but working.  Some of these statements are forward-looking.)

Some of the more interesting features:

* Written in go, so hopefully easier to maintain and extend, as complexity inevitably increases
* Uses a state-sync model, so we get things like a dry-run mode and idempotency automatically
* Based on a simple meta-model defined in a directory tree
* Can produce configurations in other formats (currently Terraform & Cloud-Init), so that we can have working
  configurations for other tools also.

## Recent changes

* Create command was split into create and update [Jul 21 2016](CHANGES.md#jul-21-2016)

## Installation

Build the code (make sure you have set GOPATH):
```
go get -d k8s.io/kops
cd ${GOPATH}/src/k8s.io/kops/
make
```

(Note that the code uses the relatively new Go vendoring, so building requires Go 1.6 or later,
or you must export GO15VENDOREXPERIMENT=1 when building with Go 1.5.  The makefile sets
GO15VENDOREXPERIMENT for you.  Go code generation does not honor the env var in 1.5, so for development
you should use Go 1.6 or later)

## Bringing up a cluster on AWS

* Ensure you have kubectl installed and on your path.  (We need it to set kubecfg configuration.)

* Set up a DNS hosted zone in Route 53, e.g. `mydomain.com`, and set up the DNS nameservers as normal
  so that domains will resolve.  You can reuse an existing domain name (e.g. `mydomain.com`), or you can create
  a "child" hosted zone (e.g. `myclusters.mydomain.com`) if you want to isolate them.  Note that with AWS Route53,
  you can have subdomains in a single hosted zone, so you can have `cluster1.testclusters.mydomain.com` under
  `mydomain.com`.

* Pick a DNS name under this zone to be the name of your cluster.  kops will set up DNS so your cluster
  can be reached on this name.  For example, if your zone was `mydomain.com`, a good name would be
  `kubernetes.mydomain.com`, or `dev.k8s.mydomain.com`, or even `dev.k8s.myproject.mydomain.com`. We'll call this `NAME`.

* Set `AWS_PROFILE` (if you need to select a profile for the AWS CLI to work)

* Pick an S3 bucket that you'll use to store your cluster configuration - this is called your state store.  You
  can `export KOPS_STATE_STORE=s3://<mystatestorebucket>` and then kops will use this location by default.  We
  suggest putting this in your bash profile or similar.  A single registry can hold multiple clusters, and it
  can also be shared amongst your ops team (which is much easier than passing around kubecfg files!)

* Run "kops create cluster" to create your cluster configuration:
```
${GOPATH}/bin/kops create cluster --cloud=aws --zones=us-east-1c ${NAME}
```
(protip: the --cloud=aws argument is optional if the cloud can be inferred from the zones)

* Run "kops update cluster" to build your cluster:
```
${GOPATH}/bin/kops update cluster ${NAME} --yes
```

If you have problems, please set `--v=8` and open an issue, and ping justinsb on slack!

## Create kubecfg settings for kubectl

(This step is actually optional; `update cluster` will do it automatically after cluster creation.
 But we expect that if you're part of a team you might share the KOPS_STATE_STORE, and then you can do
 this on different machines instead of having to share kubecfg files)

To create the kubecfg configuration settings for use with kubectl:

```
export KOPS_STATE_STORE=s3://<somes3bucket>
# NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops export kubecfg ${NAME}
```

You can now use kubernetes using the kubectl tool (after allowing a few minutes for the cluster to come up):

```kubectl get nodes```

## Cluster management

* Set up [add-ons](docs/addons.md), to add important functionality to k8s.

* Learn about [InstanceGroups](docs/instance_groups.md), which let you change instance types, cluster sizes etc.

## Delete the cluster

When you're done, you can also have kops delete the cluster.  It will delete all AWS resources tagged
with the cluster name in the specified region.

```
# NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops delete cluster ${NAME} # --yes
```

You must pass --yes to actually delete resources (without the `#` comment!)

## Other interesting modes:

* Build a terraform model: `--target=terraform`  The terraform model will be built in `out/terraform`

* Specify the k8s build to run: `--kubernetes-version=1.2.2`

* Run nodes in multiple zones: `--zones=us-east-1b,us-east-1c,us-east-1d`

* Run with a HA master: `--master-zones=us-east-1b,us-east-1c,us-east-1d`

* Specify the number of nodes: `--node-count=4`

* Specify the node size: `--node-size=m4.large`

* Specify the master size: `--master-size=m4.large`

* Override the default DNS zone: `--dns-zone=<my.hosted.zone>`

# How it works

Everything is driven by a local configuration directory tree, called the "model".  The model represents
the desired state of the world.

Each file in the tree describes a Task.

On the nodeup side, Tasks can manage files, systemd services, packages etc.
On the `kops update cluster` side, Tasks manage cloud resources: instances, networks, disks etc.

## Workaround for terraform bug

Terraform currently has a bug where it can't create AWS tags containing a dot.  Until this is fixed,
you can't use terraform to build EC2 resources that are tagged with `k8s.io/...` tags.  Thankfully this is only
the volumes, and it isn't the worst idea to build these separately anyway.

We divide the cloudup model into three parts:
* models/config which contains all the options - this is run automatically by "create cluster"
* models/proto which sets up the volumes and other data which would be hard to recover (e.g. likely keys & secrets in the near future)
* models/cloudup which is the main cloud model for configuring everything else

So you don't use terraform for the 'proto' phase (you can't anyway, because of the bug!):

```
export KOPS_STATE_STORE=s3://<somes3bucket>
export NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops create cluster --v=0 --zones=us-east-1c ${NAME}
${GOPATH}/bin/kops update cluster --v=0 ${NAME} --model=proto --yes
```

And then you can use terraform to do the remainder of the installation:

```
export CLUSTER_NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops update cluster --v=0 ${NAME} --model=cloudup --target=terraform
```

Then, to apply using terraform:

```
cd out/terraform

terraform plan
terraform apply
```
