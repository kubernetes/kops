## Building Kubernetes clusters with Terraform

Kops can generate Terraform configurations, and then you can apply them using the `terraform plan` and `terraform apply` tools. This is very handy if you are already using Terraform, or if you want to check in the Terraform output into version control.

The gist of it is that, instead of letting kops apply the changes, you tell kops what you want, and then kops spits out what it wants done into a `.tf` file. **_You_** are then responsible for turning those plans into reality.

The Terraform output should be reasonably stable (i.e. the text files should only change where something has actually changed - items should appear in the same order etc). This is extremely useful when using version control as you can diff your changes easily.

Note that if you modify the Terraform files that kops spits out, it will override your changes with the configuration state defined by its own configs. In other terms, kops's own state is the ultimate source of truth (as far as kops is concerned), and Terraform is a representation of that state for your convenience.

Ps: Steps below assume a recent version of Terraform. There's a workaround for a bug if you are using a Terraform version before 0.7 that you should be aware of (see [_"Caveats"_ section](#caveats)).

### Using Terraform

#### Set up remote state

You could keep your Terraform state locally, but we **strongly recommend** saving it on S3 with versioning turned on that bucket. Configure a remote S3 store with a setting like below:

```
terraform {
  backend "s3" {
    bucket = "mybucket"
    key    = "path/to/my/key"
    region = "us-east-1"
  }
}
```

Then run:

```
$ terraform init
```
to set up s3 backend.
Learn more [about Terraform state here](https://www.terraform.io/docs/state/remote.html).

#### Initialize/create a cluster

For example, a complete setup might be:

```
$ kops create cluster \
  --name=kubernetes.mydomain.com \
  --state=s3://mycompany.kubernetes \
  --dns-zone=kubernetes.mydomain.com \
  [... your other options ...]
  --out=. \
  --target=terraform
```

The above command will create kops state on S3 (defined in `--state`) and output a representation of your configuration into Terraform files. Thereafter you can preview your changes and then apply as shown below:

```
$ terraform plan
$ terraform apply
```

Wait for the cluster to initialize. If all goes well, you should have a working Kubernetes cluster!

#### Editing the cluster

It's possible to use Terraform to make changes to your infrastructure as defined by kops. In the example below we'd like to change some cluster configs:

```
$ kops edit cluster \
  --name=kubernetes.mydomain.com \
  --state=s3://mycompany.kubernetes

# editor opens, make your changes ...
```

Then output your changes/edits to kops cluster state into the Terraform files. Run `kops update` with `--target` and `--out` parameters:

```
$ kops update cluster \
  --name=kubernetes.mydomain.com \
  --state=s3://mycompany.kubernetes \
  --out=. \
  --target=terraform
```

Then apply your changes after previewing what changes will be applied:

```
$ terraform plan
$ terraform apply
```

Ps: You aren't limited to cluster edits i.e. `kops edit cluster`. You can also edit instances groups e.g. `kops edit instancegroup nodes|bastions` etc.

Keep in mind that some changes will require a `kops rolling-update` to be applied. When in doubt, run the command and check if any nodes needs to be updated. For more information see the [caveats](#caveats) section below.

#### Teardown the cluster

When you eventually `terraform destroy` the cluster, you should still run `kops delete cluster`, to remove the kops cluster specification and any dynamically created Kubernetes resources (ELBs or volumes). To do this, run:

```
$ terraform plan -destroy
$ terraform destroy
$ kops delete cluster --yes \
  --name=kubernetes.mydomain.com \
  --state=s3://mycompany.kubernetes
```

Ps: You don't have to `kops delete cluster` if you just want to recreate from scratch. Deleting kops cluster state means that you've have to `kops create` again.


### Caveats

#### `kops rolling-update` might be needed after editing the cluster

Changes made with `kops edit` (like enabling RBAC and / or feature gates) will result in changes to the launch configuration of your cluster nodes. After a `terraform apply`, they won't be applied right away since terraform will not launch new instances as part of that.

To see your changes applied to the cluster you'll also need to run `kops rolling-update` after running `terraform apply`. This will ensure that all nodes' changes have the desired settings configured with `kops edit`.

#### Workaround for Terraform <0.7

Before terraform version 0.7, there was a bug where it could not create AWS tags containing a dot. We recommend upgrading to version 0.7 or later, which will fix this bug. Please note that this issue only affects the volumes.

There's a workaround if you need to use an earlier version. We divide the cloudup model into three parts:

* `models/config` which contains all the options - this is run automatically by "create cluster"
* `models/proto` which sets up the volumes and other data which would be hard to recover (e.g. likely keys & secrets in the near future)
* `models/cloudup` which is the main cloud model for configuring everything else

The workaround is that you don't use terraform for the `proto` phase (you can't anyway, because of the bug!):

```
$ kops create cluster \
  --name=kubernetes.mydomain.com \
  --state=s3://mycompany.kubernetes \
  [... your other options ...]
  --out=. \
  --target=terraform

$ kops update cluster \
  --name=kubernetes.mydomain.com \
  --state=s3://mycompany.kubernetes \
  --model=proto \
  --yes
```

And then you can use terraform to do the remainder of the installation:

```
$ kops update cluster \
  --name=kubernetes.mydomain.com \
  --state=s3://mycompany.kubernetes \
  --model=cloudup \
  --out=. \
  --target=terraform
```

Then, to apply using terraform:

```
$ terraform plan
$ terraform apply
```

You should still run `kops delete cluster ${CLUSTER_NAME}`, to remove the kops cluster specification and any dynamically created Kubernetes resources (ELBs or volumes), but under this workaround also to remove the primary ELB volumes from the `proto` phase.
