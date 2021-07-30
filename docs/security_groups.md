# Security Groups

## Use existing AWS Security Groups
**Note: Use this at your own risk, when existing SGs are used kOps will NOT ensure they are properly configured.**

Rather than having kOps create and manage IAM Security Groups, it is possible to use an existing one. This is useful in organizations where security policies prevent tools from creating their own Security Groups.
kOps will still output any differences in the managed and your own Security Groups.
This is convenient for determining policy changes that need to be made when upgrading kOps.
**Using Managed Security Groups will not output these differences, it is up to the user to track expected changes to policies.**

NOTE: 

- *Currently kOps only supports using existing Security Groups for every instance group and Load Balancer in the Cluster, not a mix of existing and managed Security Groups.
This is due to the lifecycle overrides being used to prevent creation of the Security Groups related resources.*
- *kOps will add necessary rules to the security group specified in `securityGroupOverride`.*

To do this first specify the Security Groups for the ELB (if you are using a LB) and Instance Groups
Example:
```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  name: mycluster.example.com
spec:
  api:
    loadBalancer:
      securityGroupOverride: sg-abcd1234

.
.
.

apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: mycluster.example.com
  name: master-us-test-1a
spec:
  securityGroupOverride: sg-1234dcba

```

Now run a cluster update to create the new LaunchTemplateVersion, using [lifecycle overrides](./cli/kops_update_cluster.md#options) to prevent Security Group resources from being created:

```shell
kops update cluster ${CLUSTER_NAME} --yes --lifecycle-overrides SecurityGroup=ExistsAndWarnIfChanges,SecurityGroupRule=ExistsAndWarnIfChanges
```

*Every time `kops update cluster` is run, it must include the above `--lifecycle-overrides`.*

Then perform a rolling update in order to replace EC2 instances in the ASG with the new LaunchTemplateVersion:

```shell
kops rolling-update cluster ${CLUSTER_NAME} --yes
```
