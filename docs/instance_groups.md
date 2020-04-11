# The `InstanceGroup` resource

The `InstanceGroup` resource represents a group of similar machines typically provisioned in the same availability zone. On AWS, instance groups map directly to an autoscale group.

The complete list of keys can be found at the [InstanceGroup](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#InstanceGroupSpec) reference page. 

You can also find concrete use cases for the configurations on the [Instance Group operations page](instance_groups.md)

On this page, we will expand on the more important configuration keys.

## cloudLabels

If you need to add tags on auto scaling groups or instances (propagate ASG tags), you can add it in the instance group specs with `cloudLabels`. Cloud Labels defined at the cluster spec level will also be inherited.

```YAML
spec:
  cloudLabels:
    billing: infra
    environment: dev
```

## suspendProcess

Autoscaling groups automatically include multiple [scaling processes](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-suspend-resume-processes.html#process-types)
that keep our ASGs healthy.  In some cases, you may want to disable certain scaling activities.

An example of this is if you are running multiple AZs in an ASG while using a Kubernetes Autoscaler.
The autoscaler will remove specific instances that are not being used.  In some cases, the `AZRebalance` process
will rescale the ASG without warning.

```YAML
spec:
  suspendProcesses:
  - AZRebalance
```


## instanceProtection

Autoscaling groups may scale up or down automatically to balance types of instances, regions, etc.
[Instance protection](https://docs.aws.amazon.com/autoscaling/ec2/userguide/as-instance-termination.html#instance-protection) prevents the ASG from being scaled in.

```YAML
spec:
  instanceProtection: true
```

## externalLoadBalancers

Instance groups can be linked to up to 10 load balancers. When attached, any instance launched will
automatically register itself to the load balancer. For example, if you can create an instance group
dedicated to running an ingress controller exposed on a
[NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport), you can
manually create a load balancer and link it to the instance group. Traffic to the load balancer will now
automatically go to one of the nodes.

You can specify either `loadBalancerName` to link the instance group to an AWS Classic ELB or you can
specify `targetGroupArn` to link the instance group to a target group, which are used by Application
load balancers and Network load balancers.

```YAML
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  labels:
    kops.k8s.io/cluster: k8s.dev.local
  name: ingress
spec:
  machineType: m4.large
  maxSize: 2
  minSize: 2
  role: Node
  externalLoadBalancers:
  - targetGroupArn: arn:aws:elasticloadbalancing:eu-west-1:123456789012:targetgroup/my-ingress-target-group/0123456789abcdef
  - loadBalancerName: my-elb-classic-load-balancer
```

## detailedInstanceMonitoring

Detailed monitoring will cause the monitoring data to be available every 1 minute instead of every 5 minutes. [Enabling Detailed Monitoring](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-cloudwatch-new.html). In production environments you may want to consider to enable detailed monitoring for quicker troubleshooting.

**Note: that enabling detailed monitoring is a subject for [charge](https://aws.amazon.com/cloudwatch)**

```YAML
spec:
  detailedInstanceMonitoring: true
```

## additionalUserData

Kops utilizes cloud-init to initialize and setup a host at boot time. However in certain cases you may already be leveraging certain features of cloud-init in your infrastructure and would like to continue doing so. More information on cloud-init can be found [here](http://cloudinit.readthedocs.io/en/latest/)

Additional user-data can be passed to the host provisioning by setting the `additionalUserData` field. A list of valid user-data content-types can be found [here](http://cloudinit.readthedocs.io/en/latest/topics/format.html#mime-multi-part-archive)

Example:

```YAML
spec:
  additionalUserData:
  - name: myscript.sh
    type: text/x-shellscript
    content: |
      #!/bin/sh
      echo "Hello World.  The time is now $(date -R)!" | tee /root/output.txt
  - name: local_repo.txt
    type: text/cloud-config
    content: |
      #cloud-config
      apt:
        primary:
          - arches: [default]
            uri: http://local-mirror.mydomain
            search:
              - http://local-mirror.mydomain
              - http://archive.ubuntu.com
```

## sysctlParameters

To add custom kernel runtime parameters to your instance group, specify the
`sysctlParameters` field as an array of strings. Each string must take the form
of `variable=value` the way it would appear in sysctl.conf (see also
`sysctl(8)` manpage).

Unlike a simple file asset, specifying kernel runtime parameters in this manner
would correctly invoke `sysctl --system` automatically for you to apply said
parameters.

For example:

```YAML
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  name: nodes
spec:
  sysctlParameters:
    - fs.pipe-user-pages-soft=524288
    - net.ipv4.tcp_keepalive_time=200
```

which would end up in a drop-in file on nodes of the instance group in question.

## mixedInstancePolicy (AWS Only)

### Instances

Instances is a list of instance types which we are willing to run in the EC2 fleet

### onDemandAllocationStrategy

Indicates how to allocate instance types to fulfill On-Demand capacity

### onDemandBase

OnDemandBase is the minimum amount of the Auto Scaling group's capacity that must be
fulfilled by On-Demand Instances. This base portion is provisioned first as your group scales.

### onDemandAboveBase

OnDemandAboveBase controls the percentages of On-Demand Instances and Spot Instances for your
additional capacity beyond OnDemandBase. The range is 0–100. The default value is 100. If you
leave this parameter set to 100, the percentages are 100% for On-Demand Instances and 0% for
Spot Instances.

### spotAllocationStrategy

SpotAllocationStrategy diversifies your Spot capacity across multiple instance types to
find the best pricing. Higher Spot availability may result from a larger number of
instance types to choose from https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-fleet.html#spot-fleet-allocation-strategy

### spotInstancePools

SpotInstancePools is the number of Spot pools to use to allocate your Spot capacity (defaults to 2)
pools are determined from the different instance types in the Overrides array of LaunchTemplate