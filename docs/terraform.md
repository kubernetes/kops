## Building Kubernetes clusters with Terraform

Kops can generate Terraform configurations, and then you can apply them using the `terraform plan` and `terraform apply` tools. This is very handy if you are already using Terraform, or if you want to check in the Terraform output into version control.

The gist of it is that, instead of letting kops apply the changes, you tell kops what you want, and then kops spits out what it wants done into a `.tf` file. **_You_** are then responsible for turning those plans into reality.

The Terraform output should be reasonably stable (i.e. the text files should only change where something has actually changed - items should appear in the same order etc). This is extremely useful when using version control as you can diff your changes easily.

Note that if you modify the Terraform files that kops spits out, it will override your changes with the configuration state defined by its own configs. In other terms, kops's own state is the ultimate source of truth (as far as kops is concerned), and Terraform is a representation of that state for your convenience.

### Terraform Version Compatibility
| Kops Version | Terraform Version | Feature Flag Notes |
|--------------|-------------------|-------|
| >= 1.18      | >= 0.12           | HCL2 supported by default |
| >= 1.18      | < 0.12            | `KOPS_FEATURE_FLAGS=-Terraform-0.12` |
| >= 1.17      | >= 0.12           | `KOPS_FEATURE_FLAGS=TerraformJSON` outputs JSON |
| <= 1.17      | < 0.12            | Supported by default |

### Using Terraform

#### Set up remote state

You could keep your Terraform state locally, but we **strongly recommend** saving it on S3 with versioning turned on that bucket. Configure a remote S3 store with a setting like below:

```terraform
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

#### Terraform JSON output

With terraform 0.12 JSON is now officially supported as configuration language. To enable JSON output instead of HCLv1 output you need to enable it through a feature flag.
```
export KOPS_FEATURE_FLAGS=TerraformJSON
kops update cluster .....
```

This is an alternative to of using terraforms own configuration syntax HCL. Be sure to delete the existing kubernetes.tf file. Terraform will otherwise use both and then complain. 

Kops will require terraform 0.12 for JSON configuration. Inofficially (partially) it was also supported with terraform 0.11, so you can try and remove the `required_version` in `kubernetes.tf.json`. 

