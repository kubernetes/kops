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

* Ensure you have a DNS hosted zone set up in Route 53, e.g. `mydomain.com`

* Pick a DNS name under this zone to be the name of your cluster.  kops will set up DNS so your cluster
can be reached on this name.  For example, if your zone was `mydomain.com`, a good name would be
`kubernetes.mydomain.com`, or `dev.k8s.mydomain.com`, or even `dev.k8s.myproject.mydomain.com`. We'll call this `NAME`.

* Set `AWS_PROFILE` (if you need to select a profile for the AWS CLI to work)

* Pick an S3 bucket that you'll use to store your cluster configuration - this is called your state store.

* Execute:
```
export NAME=<kubernetes.mydomain.com>
export KOPS_STATE_STORE=s3://<somes3bucket>
${GOPATH}/bin/kops create cluster --v=0 --cloud=aws --zones=us-east-1c --name=${NAME}
```

(protip: the --cloud=aws argument is optional if the cloud can be inferred from the zones)

If you have problems, please set `--v=8` and open an issue, and ping justinsb on slack!

## Build a kubectl file

The kops tool is a CLI for doing administrative tasks.  You can use it to create the kubecfg configuration,
for use with kubectl:

```
export NAME=<kubernetes.mydomain.com>
export KOPS_STATE_STORE=s3://<somes3bucket>
${GOPATH}/bin/kops export kubecfg --name=${NAME}
```

## Delete the cluster

When you're done, you can also have kops delete the cluster.  It will delete all AWS resources tagged
with the cluster name in the specified region.

```
export NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops delete cluster --region=us-east-1 --name=${NAME} # --yes
```

You must pass --yes to actually delete resources (without the `#` comment!)

## Other interesting modes:

* See changes that would be applied: `--dryrun`

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
On the `kops create cluster` side, Tasks manage cloud resources: instances, networks, disks etc.

## Workaround for terraform bug

Terraform currently has a bug where it can't create AWS tags containing a dot.  Until this is fixed,
you can't use terraform to build EC2 resources that are tagged with `k8s.io/...` tags.  Thankfully this is only
the volumes, and it isn't the worst idea to build these separately anyway.

We divide the 'kops create cluster' model into three parts:
* models/config which contains all the options
* models/proto which sets up the volumes and other data which would be hard to recover (e.g. likely keys & secrets in the near future)
* models/cloudup which is the main cloud model for configuring everything else

So you don't use terraform for the 'proto' phase (you can't anyway, because of the bug!):

```
export KOPS_STATE_STORE=s3://<somes3bucket>
export CLUSTER_NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops create cluster --v=0 --zones=us-east-1c --name=${CLUSTER_NAME} --model=config,proto
```

And then you can use terraform to do the remainder of the installation:

```
export CLUSTER_NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops create cluster --v=0 --zones=us-east-1c --name=${CLUSTER_NAME} --model=config,cloudup --target=terraform
```

Then, to apply using terraform:

```
cd out/terraform

terraform plan
terraform apply
```
