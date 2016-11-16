# Bastion in Kops

Bastion provide an external facing point of entry into a network containing private network instances. This host can provide a single point of fortification or audit and can be started and stopped to enable or disable inbound SSH communication from the Internet, some call bastion as the "jump server".

Note: Bastion will get setup for the cluster(by default) only when `--topology="private"`.

* [More information on bastion from aws](http://docs.aws.amazon.com/quickstart/latest/linux-bastion/architecture.html)
* [More information on bastion from gce](https://cloud.google.com/solutions/connecting-securely#bastion)

## AWS

### Specify instance type of bastion

Instance types in AWS comprise varying combinations of CPU, memory, storage, and networking capacity and give you the flexibility to choose the appropriate mix of resources for your applications.
```
kops create cluster --bastion-instance-type="t2.large"
```

Bastion instance type will default to `t2.medium`

[More information](https://aws.amazon.com/ec2/instance-types/)


#### Enable/disable bastion, defaults to false
To turn on/off bastion host setup completely.

|   Turn on/off Bastion   |    Example                                | Bastion ASG settings
| ----------------------- |------------------------------------------ | --------------------
|   Enable Bastion        |   `kops create cluster --bastion=true`    | ASG's desired/min/max set to default value 1
|   Disable Bastion       |   `kops create cluster --bastion=false`   | ASG' desired/min/max = 0

#### Reach bastion from outside of vpc using a name

When the cluster is created using below -
```
kops create cluster --bastion-name="bastion" --dns-zone="uswest1.clusters.example.com"
```
This will create a route53 entry for `bastion.uswest1.clusters.example.com` mapping with bastion ASG's ELB. And bastion can be reached using
```
ssh -i <key> ubuntu@bastion.uswest1.clusters.example.com
```

### High idle timeout for bastion ASG's ELB.

By default, elastic load balancing sets the idle timeout to 60 seconds. This value can be configured by the user using `-bastion-elb-idle-timeout=120` for making it 120 seconds.

[More information](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html)

### Improve access to bastion instances

Current: `ssh to the bastion: ssh -i ~/.ssh/id_rsa admin@api.mydomain.com`
