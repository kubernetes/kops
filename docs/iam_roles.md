# IAM Roles

Two IAM roles are created for the cluster: one for the masters, and one for the nodes.

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


## Adding Additional Policies

Sometimes you may need to extend the kops IAM roles to add additional policies. You can do this
through the `additionalPolicies` spec field. For instance, let's say you want
to add DynamoDB and Elasticsearch permissions to your nodes.

Edit your cluster via `kops edit cluster ${CLUSTER_NAME}` and add the following to the spec:

```
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
