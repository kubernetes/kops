# Instance IAM Roles

By default, kOps creates two instance IAM roles for the cluster: one for the control plane and one for the worker nodes.

> As of kOps 1.22, new clusters running Kubernetes 1.22 on AWS will restrict Pod access to the instance metadata service.
> This means that Pods will also be prevented from directly assuming instance roles.
> See [IAM Roles for ServiceAccounts](/cluster_spec/#service-account-issuer-discovery-and-aws-iam-roles-for-service-accounts-irsa) and [instance metadata](/instance_groups/#instancemetadata) documentation.
> Before this, all Pods running on your cluster have access to the instance IAM role.
> Consider enabling the protection mentioned above and use IRSA for your own workloads.

## Access to AWS EC2 Container Registry (ECR)

The default IAM roles will not grant nodes access to the AWS EC2 Container Registry (ECR). To grant access to ECR, update your Cluster Spec with the following and then perform a cluster update:
```yaml
iam:
  allowContainerRegistry: true
```

Adding ECR permissions will extend the IAM policy documents as below:
- Control Plane Nodes: https://github.com/kubernetes/kops/blob/master/pkg/model/iam/tests/iam_builder_master_strict_ecr.json
- Worker Nodes: https://github.com/kubernetes/kops/blob/master/pkg/model/iam/tests/iam_builder_node_strict_ecr.json

The additional permissions are:
```json
{
  "Sid": "kOpsK8sECR",
  "Effect": "Allow",
  "Action": [
    "ecr:BatchCheckLayerAvailability",
    "ecr:BatchGetImage",
    "ecr:BatchImportUpstreamImage",
    "ecr:DescribeRepositories",
    "ecr:GetAuthorizationToken",
    "ecr:GetDownloadUrlForLayer",
    "ecr:GetRepositoryPolicy",
    "ecr:ListImages"
  ],
  "Resource": [
    "*"
  ]
}
```
## Permissions Boundaries
{{ kops_feature_table(kops_added_default='1.19') }}

AWS Permissions Boundaries enable you to use a policy (managed or custom) to set the maximum permissions that roles created by kOps will be able to grant to instances they're attached to. It can be useful to prevent possible privilege escalations.

To set a Permissions Boundary for kOps' roles, update your Cluster Spec with the following and then perform a cluster update:
```yaml
iam:
  permissionsBoundary: aws:arn:iam:123456789000:policy:test-boundary
```

*NOTE: Currently, kOps only supports using a single Permissions Boundary for all roles it creates. In case you need to set per-role Permissions Boundaries, we recommend that you refer to this [section](#use-existing-aws-instance-profiles) below, and provide your own roles to kOps.*

## Adding External Policies

{{ kops_feature_table(kops_added_default='1.18') }}

At times, you may want to attach policies shared to you by another AWS account or that are maintained by an outside application. You can specify managed policies through the `policyOverrides` spec field.

Policy Overrides are specified by their ARN on AWS and are grouped by their role type. See the example below:

```yaml
spec:
  externalPolicies:
    node:
    - arn:aws:iam::123456789000:policy/test-policy
    master:
    - arn:aws:iam::123456789000:policy/test-policy
    bastion:
    - arn:aws:iam::123456789000:policy/test-policy
```

External Policy attachments are treated declaratively. Any policies declared will be attached to the role, any policies not specified will be detached _after_ new policies are attached. This does not replace or affect built in kOps policies in any way.

It's important to note that externalPolicies will only handle the attachment and detachment of policies, not creation, modification, or deletion of them.

## Adding Additional Policies

Sometimes you may need to extend the kOps instance IAM roles to add additional policies. You can do this
through the `additionalPolicies` spec field. For instance, let's say you want
to add DynamoDB and Elasticsearch permissions to your nodes.

Edit your cluster via `kops edit cluster ${CLUSTER_NAME}` and add the following to the spec:

```yaml
spec:
  additionalPolicies:
    node: |
      [
        {
          "Effect": "Allow",
          "Action": ["dynamodb:*"],
          "Resource": ["*"]
        },
        {
          "Effect": "Allow",
          "Action": ["es:*"],
          "Resource": ["*"]
        }
      ]
```

After you're finished editing, your cluster spec should look something like this:

```yaml
metadata:
  name: ${CLUSTER_NAME}
spec:
  cloudProvider: aws
  networkCIDR: 10.100.0.0/16
  networkID: vpc-a80734c1
  nonMasqueradeCIDR: 100.64.0.0/10
  zones:
  - cidr: 10.100.32.0/19
    name: eu-central-1a
  additionalPolicies:
    node: |
      [
        {
          "Effect": "Allow",
          "Action": ["dynamodb:*"],
          "Resource": ["*"]
        },
        {
          "Effect": "Allow",
          "Action": ["es:*"],
          "Resource": ["*"]
        }
      ]
```

Now you can run a cluster update to have the changes take effect:

```shell
kops update cluster ${CLUSTER_NAME} --yes
```

You can have an additional set of policies for each kOps instance role (node, master, bastion). For instance, if you wanted to apply one set of additional permissions to the master instances, and another to the nodes, you could do the following:

```yaml
spec:
  additionalPolicies:
    node: |
      [
        {
          "Effect": "Allow",
          "Action": ["es:*"],
          "Resource": ["*"]
        }
      ]
    master: |
      [
        {
          "Effect": "Allow",
          "Action": ["dynamodb:*"],
          "Resource": ["*"]
        }
      ]
```

## Use existing AWS Instance Profiles

Rather than having kOps create and manage IAM roles and instance profiles, it is possible to use an existing instance profile. This is useful in organizations where security policies prevent tools from creating their own IAM roles and policies.
kOps will still output any differences in the IAM Inline Policy for each IAM Role.
This is convenient for determining policy changes that need to be made when upgrading kOps.
**Using IAM Managed Policies will not output these differences; it is up to the user to track expected changes to policies.**

*NOTE: Currently kOps only supports using existing instance profiles for every instance group in the cluster, not a mix of existing and managed instance profiles.
This is due to the lifecycle overrides being used to prevent creation of the IAM-related resources.*

To do this, get a list of instance group names for the cluster:

```shell
kops get ig --name ${CLUSTER_NAME}
```

And update every instance group's spec with the desired instance profile ARNs:

```shell
kops edit ig --name ${CLUSTER_NAME} ${INSTANCE_GROUP_NAME}
```

Adding the following `iam` section to the spec:

```yaml
spec:
  iam:
    profile: arn:aws:iam::1234567890108:instance-profile/kops-custom-node-role
```

Now run a cluster update to create the new launch template version, using [lifecycle overrides](./cli/kops_update_cluster.md#options) to prevent IAM-related resources from being created:

```shell
kops update cluster ${CLUSTER_NAME} --yes --lifecycle-overrides IAMRole=ExistsAndWarnIfChanges,IAMRolePolicy=ExistsAndWarnIfChanges,IAMInstanceProfileRole=ExistsAndWarnIfChanges
```

*Every time `kops update cluster` is run, it must include the above `--lifecycle-overrides` unless a non-`security` phase is specified.*

Finally, perform a rolling update in order to replace EC2 instances in the ASG with the new launch template version:

```shell
kops rolling-update cluster ${CLUSTER_NAME} --yes
```
