# IAM Roles

By default Kops creates two IAM roles for the cluster: one for the masters, and one for the nodes.

> Please note that currently all Pods running on your cluster have access to the instance IAM role.
> Consider using projects such as [kube2iam](https://github.com/jtblin/kube2iam) to prevent that.

Work has been done on scoping permissions to the minimum required for a functional Kubernetes Cluster, resulting in a fully revised set of IAM policies for both master & compute nodes.

An example of the new IAM policies can be found here:

- Master Nodes: https://github.com/kubernetes/kops/blob/master/pkg/model/iam/tests/iam_builder_master_strict.json
- Compute Nodes: https://github.com/kubernetes/kops/blob/master/pkg/model/iam/tests/iam_builder_node_strict.json

On provisioning a new cluster with Kops v1.8.0 or above, by default you will be using the new stricter IAM policies. Upgrading an existing cluster will use the legacy IAM privileges to reduce risk of potential regression.

In order to update your cluster to use the strict IAM privileges, add the following within your Cluster Spec:
```yaml
iam:
  legacy: false
```

Following this, run a cluster update to have the changes take effect:

```
kops update cluster ${CLUSTER_NAME} --yes
```

The Strict IAM flag by default will not grant nodes access to the AWS EC2 Container Registry (ECR), as can be seen by the above example policy documents. To grant access to ECR, update your Cluster Spec with the following and then perform a cluster update:
```yaml
iam:
  allowContainerRegistry: true
  legacy: false
```

Adding ECR permissions will extend the IAM policy documents as below:
- Master Nodes: https://github.com/kubernetes/kops/blob/master/pkg/model/iam/tests/iam_builder_master_strict_ecr.json
- Compute Nodes: https://github.com/kubernetes/kops/blob/master/pkg/model/iam/tests/iam_builder_node_strict_ecr.json

The additional permissions are:
```json
{
  "Sid": "kopsK8sECR",
  "Effect": "Allow",
  "Action": [
    "ecr:BatchCheckLayerAvailability",
    "ecr:BatchGetImage",
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

## Adding External Policies

At times you may want to attach policies shared to you by another AWS account or that are maintained by an outside application. You can specify managed policies through the `policyOverrides` spec field.

Policy Overrides are specified by their ARN on AWS and are grouped by their role type. See the example below:

```yaml
spec:
  externalPolicies:
    node:
    - aws:arn:iam:123456789000:policy:test-policy
    master:
    - aws:arn:iam:123456789000:policy:test-policy
    bastion:
    - aws:arn:iam:123456789000:policy:test-policy
```

External Policy attachments are treated declaritively. Any policies declared will be attached to the role, any policies not specified will be detached _after_ new policies are attached. This does not replace or affect built in Kops policies in any way.

It's important to note that externalPolicies will only handle the attachment and detachment of policies, not creation, modification, or deletion.

## Adding Additional Policies

Sometimes you may need to extend the kops IAM roles to add additional policies. You can do this
through the `additionalPolicies` spec field. For instance, let's say you want
to add DynamoDB and Elasticsearch permissions to your nodes.

Edit your cluster via `kops edit cluster ${CLUSTER_NAME}` and add the following to the spec:

```
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

```
metadata:
  creationTimestamp: "2016-06-27T14:23:34Z"
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

```
kops update cluster ${CLUSTER_NAME} --yes
```

You can have an additional policy for each kops role (node, master, bastion). For instance, if you wanted to apply one set of additional permissions to the master instances, and another to the nodes, you could do the following:

```
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

Rather than having Kops create and manage IAM roles and instance profiles, it is possible to use an existing instance profile. This is useful in organizations where security policies prevent tools from creating their own IAM roles and policies.
Kops will still output any differences in the IAM Inline Policy for each IAM Role.
This is convenient for determining policy changes that need to be made when upgrading Kops.
**Using IAM Managed Policies will not output these differences, it is up to the user to track expected changes to policies.**

*NOTE: Currently Kops only supports using existing instance profiles for every instance group in the cluster, not a mix of existing and managed instance profiles.
This is due to the lifecycle overrides being used to prevent creation of the IAM-related resources.*

To do this, get a list of instance group names for the cluster:

```
kops get ig --name ${CLUSTER_NAME}
```

And update every instance group's spec with the desired instance profile ARNs:

```
kops edit ig --name ${CLUSTER_NAME} ${INSTANCE_GROUP_NAME}
```

Adding the following `iam` section to the spec:

```yaml
spec:
  iam:
    profile: arn:aws:iam::1234567890108:instance-profile/kops-custom-node-role
```

Now run a cluster update to create the new launch configuration, using [lifecycle overrides](./cli/kops_update_cluster.md#options) to prevent IAM-related resources from being created:

```
kops update cluster ${CLUSTER_NAME} --yes --lifecycle-overrides IAMRole=ExistsAndWarnIfChanges,IAMRolePolicy=ExistsAndWarnIfChanges,IAMInstanceProfileRole=ExistsAndWarnIfChanges
```

*Every time `kops update cluster` is run, it must include the above `--lifecycle-overrides` unless a non-`security` phase is specified.*

Finally, perform a rolling update in order to replace EC2 instances in the ASG with the new launch configuration:

```
kops rolling-update cluster ${CLUSTER_NAME} --yes
```
