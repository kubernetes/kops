# Security Groups

## Use existing AWS Security Groups
**Note: Use this at your own risk, when existing SGs are used Kops will NOT ensure they are properly configured.**

Rather than having Kops create and manage IAM Security Groups, it is possible to use an existing one. This is useful in organizations where security policies prevent tools from creating their own Security Groups.
Kops will still output any differences in the managed and your own Security Groups.
This is convenient for determining policy changes that need to be made when upgrading Kops.
**Using Managed Security Groups will not output these differences, it is up to the user to track expected changes to policies.**

NOTE: 

- *Currently Kops only supports using existing Security Groups for every instance group and Load Balancer in the Cluster, not a mix of existing and managed Security Groups.
This is due to the lifecycle overrides being used to prevent creation of the Security Groups related resources.*
- *Kops will add necessary rules to the security group specified in `securityGroupOverride`.*

To do this first specify the Security Groups for the ELB (if you are using a LB) and Instance Groups
Example:
```yaml
apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  creationTimestamp: "2016-12-10T22:42:27Z"
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
  creationTimestamp: "2017-01-01T00:00:00Z"
  labels:
    kops.k8s.io/cluster: mycluster.example.com
  name: master-us-test-1a
spec:
  securityGroupOverride: sg-1234dcba

```

Now run a cluster update to create the new launch configuration, using [lifecycle overrides](./cli/kops_update_cluster.md#options) to prevent Security Group resources from being created:

```
kops update cluster ${CLUSTER_NAME} --yes --lifecycle-overrides SecurityGroup=ExistsAndWarnIfChanges,SecurityGroupRule=ExistsAndWarnIfChanges
```

*Every time `kops update cluster` is ran, it must include the above `--lifecycle-overrides`.*

Then perform a rolling update in order to replace EC2 instances in the ASG with the new launch configuration:

```
kops rolling-update cluster ${CLUSTER_NAME} --yes
```
