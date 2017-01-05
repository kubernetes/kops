# Bastion in Kops

Bastion provide an external facing point of entry into a network containing private network instances. This host can provide a single point of fortification or audit and can be started and stopped to enable or disable inbound SSH communication from the Internet, some call bastion as the "jump server".

* [More information on bastion from aws](http://docs.aws.amazon.com/quickstart/latest/linux-bastion/architecture.html)
* [More information on bastion from gce](https://cloud.google.com/solutions/connecting-securely#bastion)

## AWS

### Enable/Disable bastion

To enable a bastion instance group, a user will need to set the `--bastion` flag on cluster create

```
kops create cluster --topology private --networking $provider --bastion $NAME
```

### Configure the bastion instance group

You can edit the bastion instance group to make changes. By default the name of the bastion instance group will be `bastions` and you can specify the name of the cluster with `--name` as in:

```
kops edit ig bastions --name $KOPS_NAME
```

You should now be able to edit and configure your bastion instance group.

```
apiVersion: kops/v1alpha2
kind: InstanceGroup
metadata:
  creationTimestamp: "2017-01-05T13:37:07Z"
  name: bastions
spec:
  associatePublicIp: true
  image: kope.io/k8s-1.4-debian-jessie-amd64-hvm-ebs-2016-10-21
  machineType: t2.micro
  maxSize: 1
  minSize: 1
  role: Bastion
  subnets:
  - utility-us-east-2a
```

**Note**: If you want to turn off the bastion server, you must set the instance group `maxSize` and `minSize` fields to `0`.

If you do not want the bastion instance group created at all, simply drop the `--bastion` flag off of your create command. The instance group will never be created.


### Using a public CNAME to access your bastion

By default the bastion instance group will create a public CNAME alias that will point to the bastion ELB. 

The default bastion name is `bastion-$NAME` as in

```
bastion-example.kubernetes.com
```

Unless a user is using `--dns-zone` which will inherently use the `basion-$ZONE` syntax.

You can define a custom bastion CNAME by editing the main cluster config `kops edit cluster $NAME` and modifying the following block

```
spec:
  topology:
    bastion:
      bastionPublicName: bastion-example.kubernetes.com
```


### Changing your ELB idle timeout

The bastion is accessed via an AWS ELB. Kops will by default set the bastion ELB idle timeout to 5 minutes. This is important for SSH connections to the bastion that you plan to keep open.

You can increase the ELB idle timeout by editing the main cluster config `kops edit cluster $NAME` and modifyng the following block

```
spec:
  topology:
    bastion:
      idleTimeoutSeconds: 1200
```

Where the maximum value is 1200 seconds (20 minutes) allowed by AWS. [More information](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html)
