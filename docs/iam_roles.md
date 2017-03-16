# IAM Roles

Two IAM roles are created for the cluster: one for the masters, and one for the nodes.

> Work is being done on scoping permissions to the minimum required to setup and maintain cluster.
> Please note that currently all Pods running on your cluster have access to instance IAM role.
> Consider using projects such as [kube2iam](https://github.com/jtblin/kube2iam) to prevent that.

Master permissions:

```
ec2:*
elasticloadbalancing:*
ecr:GetAuthorizationToken
ecr:BatchCheckLayerAvailability
ecr:GetDownloadUrlForLayer
ecr:GetRepositoryPolicy
ecr:DescribeRepositories
ecr:ListImages
ecr:BatchGetImage
route53:ListHostedZones
route53:GetChange
// The following permissions are scoped to AWS Route53 HostedZone used to bootstrap the cluster
// arn:aws:route53:::hostedzone/$hosted_zone_id
route53:ChangeResourceRecordSets, ListResourceRecordSets, GetHostedZone

// The following permissions are only created if you are using etcd volumes with "encrypted: true" and a custom kmsKeyId.
// They are scoped to the kmsKeyId that you are using.
kms:Encrypt
kms:Decrypt
kms:ReEncrypt*
kms:GenerateDataKey*
kms:DescribeKey
kms:CreateGrant
kms:ListGrants
kms:RevokeGrant
```

Node permissions:

```
ec2:Describe*
ecr:GetAuthorizationToken
ecr:BatchCheckLayerAvailability
ecr:GetDownloadUrlForLayer
ecr:GetRepositoryPolicy
ecr:DescribeRepositories
ecr:ListImages
ecr:BatchGetImage
route53:ListHostedZones
route53:GetChange
// The following permissions are scoped to AWS Route53 HostedZone used to bootstrap the cluster
// arn:aws:route53:::hostedzone/$hosted_zone_id
route53:ChangeResourceRecordSets, ListResourceRecordSets, GetHostedZone
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

Now you can update to have the changes take effect:

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

## Reusing Existing Policies

Sometimes you may need to reuse existing IAM roles. You can do this
through the `customPolicies` cluster spec API field.  This setting is highly advanced
and only enabled via CustomPoliciesSupport feature flag.  Setting the wrong role
permissions can impact various components inside of Kubernetes, and cause
unexpected issues.  This feature is in place to support the initial documenting and testing the creation of custom roles. Again, use the existing kops functionality, or reach out
if you want to help!

At this point, we do not have a full definition of the fine grain roles. Please refer
[to](https://github.com/kubernetes/kops/issues/1873) for more information.

Please use this feature wisely! Enable the feature flag by:

```console
$ export KOPS_FEATURE_FLAGS="+CustomPoliciesSupport"
```
Inside the cluster spec define one or two roles specific to the master and
a node.

```yaml
  customPolicies:
    node: "arn:aws:iam::123456789012:role/kops-node"
    master: "arn:aws:iam::123456789012:role/kops-master"
``` 
