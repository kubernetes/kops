## Building Kubernetes clusters with terraform

Kops can generate terraform configurations, and you can then apply them using the terraform plan/apply tools. 
This is very handy if you are already using terraform, or if you want to check in the terraform output into
version control.

The terraform output should be reasonably stable (i.e. the text files should only change where something has actually
changed - items should appear in the same order etc).


### Using terraform

To use terraform, you simple run update with `--target=terraform` (but see below for a workaround for a bug
if you are using a terraform version before 0.7)

For example, a complete setup might be:

```
export KOPS_STATE_STORE=s3://<somes3bucket>
export CLUSTER_NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops create cluster ${NAME} --zones us-east-1c
${GOPATH}/bin/kops update cluster ${NAME} --target=terraform

cd out/terraform
terraform plan
terraform apply
```

When you eventually `terraform delete` the cluster, you should still run `kops delete cluster ${CLUSTER_NAME}`,
to remove the kops cluster specification and any dynamically created Kubernetes resources (ELBs or volumes).

### Workaround for Terraform versions before 0.7

Before terraform version 0.7, there was a bug where it could not create AWS tags containing a dot.

We recommend upgrading to version 0.7 or laster, which wil fix this bug.

However, if you need to use an earlier version:

This issue only affects the volumes.

We divide the cloudup model into three parts:
* models/config which contains all the options - this is run automatically by "create cluster"
* models/proto which sets up the volumes and other data which would be hard to recover (e.g. likely keys & secrets in the near future)
* models/cloudup which is the main cloud model for configuring everything else

So the workaround is that you don't use terraform for the `proto` phase (you can't anyway, because of the bug!):

```
export KOPS_STATE_STORE=s3://<somes3bucket>
export CLUSTER_NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops create cluster ${CLUSTER_NAME} --zones=us-east-1c
${GOPATH}/bin/kops update cluster ${CLUSTER_NAME} --model=proto --yes
```

And then you can use terraform to do the remainder of the installation:

```
export CLUSTER_NAME=<kubernetes.mydomain.com>
${GOPATH}/bin/kops update cluster ${CLUSTER_NAME} --model=cloudup --target=terraform
```

Then, to apply using terraform:

```
cd out/terraform

terraform plan
terraform apply
```

You should still run `kops delete cluster ${CLUSTER_NAME}`, to remove the kops cluster specification and any
dynamically created Kubernetes resources (ELBs or volumes), but under this workaround also to remove the primary
ELB volumes from the `proto` phase.
