# Terraform 0.12 Naming Compatibility

Terraform 0.12 introduced new restrictions on naming, breaking
compatibility with earlier terraform versions when resource names
start with a number.  Single-zone etcd clusters (and possibly some
other scenarios) would generate terraform names for EBS volumes that
start with a number, which are no longer permitted.

For new clusters, kOps now avoids this problem.  But for existing
clusters, in order for terraform not to erase your data, a manual
state migration is needed first.

In order to prevent against data-loss, kOps will detect the problem
and require you to pass an environment variable to indicate that you
have performed the migration.

NOTE: You must perform this migration with terraform 0.11.

To do this state migration, first run `terraform state list`.

You should see something like this, depending on how many
control-plane nodes you have:

```
...
aws_ebs_volume.1-etcd-events-foo-example-com
aws_ebs_volume.1-etcd-main-foo-example-com
aws_ebs_volume.2-etcd-events-foo-example-com
aws_ebs_volume.2-etcd-main-foo-example-com
aws_ebs_volume.3-etcd-events-foo-example-com
aws_ebs_volume.3-etcd-main-foo-example-com
...
```

We want to prefix each of those names with `ebs-`.

A one liner to do so is:

```
terraform-0.11 state list | grep aws_ebs_volume | cut -d. -f2 | xargs -I {} terraform-0.11 state mv aws_ebs_volume.{} aws_ebs_volume.ebs-{}
```

This is equivalent to the manual form:

```
terraform-0.11 state mv aws_ebs_volume.1-etcd-events-foo-example-com aws_ebs_volume.ebs-1-etcd-events-foo-example-com
terraform-0.11 state mv aws_ebs_volume.1-etcd-main-foo-example-com aws_ebs_volume.ebs-1-main-events-foo-example-com
terraform-0.11 state mv aws_ebs_volume.2-etcd-events-foo-example-com aws_ebs_volume.ebs-2-etcd-events-foo-example-com
terraform-0.11 state mv aws_ebs_volume.2-etcd-main-foo-example-com aws_ebs_volume.ebs-2-etcd-main-foo-example-com
terraform-0.11 state mv aws_ebs_volume.3-etcd-events-foo-example-com aws_ebs_volume.ebs-3-etcd-events-foo-example-com
terraform-0.11 state mv aws_ebs_volume.3-etcd-main-foo-example-com aws_ebs_volume.ebs-3-etcd-main-foo-example-com
```

Finally, you should repeat the kops update command passing
`KOPS_TERRAFORM_0_12_RENAMED=ebs`.

Note that you must then run `terraform init` / `terraform plan` /
`terraform apply` using terraform 0.12.26, not terraform 0.13.

Carefully review the output of `terraform plan` / `terraform apply` to
ensure that the EBS volumes are not being deleted & recreated.  Note
that `aws_security_group_rule` will be deleted and recreated, due to
the same terraform naming restriction.

If you encounter the "A duplicate Security Group rule..." error, you
will likely have to run `terraform apply` twice, because of the
terraform bug described in
`https://github.com/hashicorp/terraform/pull/2376`

Note that you must _always_ pass `KOPS_TERRAFORM_0_12_RENAMED=ebs` to
`kops` for these clusters, as kOps otherwise has no way to know that
the rename has been done.  However, kOps will "fail safe" and simply
refuse to generate terraform in these cases.
