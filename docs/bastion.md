# Bastion in Kops

Bastion provide an external facing point of entry into a network containing private network instances. This host can provide a single point of fortification or audit and can be started and stopped to enable or disable inbound SSH communication from the Internet, some call bastion as the "jump server".

* [More information on bastion from aws](http://docs.aws.amazon.com/quickstart/latest/linux-bastion/architecture.html)
* [More information on bastion from gce](https://cloud.google.com/solutions/connecting-securely#bastion)

## AWS

### Enable/Disable bastion

To enable a bastion instance group, a user will need to set the `--bastion` flag on cluster create

```yaml
kops create cluster --topology private --networking $provider --bastion $NAME
```

To add a bastion instance group to a pre-existing cluster, create a new instance group with the `--role Bastion` flag and one or more subnets (e.g. `utility-us-east-2a,utility-us-east-2b`). 
```yaml
kops create instancegroup bastions --role Bastion --subnet $SUBNET
```

### Configure the bastion instance group

You can edit the bastion instance group to make changes. By default the name of the bastion instance group will be `bastions` and you can specify the name of the cluster with `--name` as in:

```yaml
kops edit ig bastions --name $KOPS_NAME
```

You should now be able to edit and configure your bastion instance group.

```yaml
apiVersion: kops.k8s.io/v1alpha2
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

The default bastion name is `bastion.$NAME` as in

```yaml
bastion.mycluster.example.com
```

Unless a user is using `--dns-zone` which will inherently use the `bastion-$ZONE` syntax.

You can define a custom bastion CNAME by editing the main cluster config `kops edit cluster $NAME` and modifying the following block

```yaml
spec:
  topology:
    bastion:
      bastionPublicName: bastion.mycluster.example.com
```

### Additional security groups to ELB
If you want to add security groups to the bastion ELB

```yaml
spec:
  topology:
    bastion:
      bastionPublicName: bastion.mycluster.example.com
      loadBalancer:
        additionalSecurityGroups:
        - "sg-***"
```

### Access when using gossip

When using [gossip mode](gossip.md), there is no DNS zone where we can configure a
CNAME for the bastion. Because bastions are fronted with a load
balancer, you can instead use the endpoint of the load balancer to
reach your bastion.

On AWS, an easy way to find this DNS name is with kops toolbox:

```
kops toolbox dump -ojson | grep 'bastion.*elb.amazonaws.com'
```

### Using SSH agent to access your bastion

Verify your local agent is configured correctly

```
$ ssh-add -L
ssh-rsa <PUBLIC_RSA_HASH> /Users/kris/.ssh/id_rsa
```

If that command returns no results, add the key to `ssh-agent`

```
ssh-add ~/.ssh/id_rsa
```

Check if the key is now added using `ssh-add -L`

SSH into the bastion, then into a master

```
ssh -A admin@<bastion_elb_a_record>
ssh admin@<master_ip>
```

### Changing your ELB idle timeout

The bastion is accessed via an AWS ELB. The ELB is required to gain secure access into the private network and connect the user to the ASG that the bastion lives in. Kops will by default set the bastion ELB idle timeout to 5 minutes. This is important for SSH connections to the bastion that you plan to keep open.

You can increase the ELB idle timeout by editing the main cluster config `kops edit cluster $NAME` and modifying the following block

```yaml
spec:
  topology:
    bastion:
      idleTimeoutSeconds: 1200
```

Where the maximum value is 3600 seconds (60 minutes) allowed by AWS. For more information see [configuring idle timeouts](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html).

### Using the bastion

Once your cluster is setup and you need to SSH into the bastion you can access a cluster resource using the following steps

```bash
# Verify you have an SSH agent running. This should match whatever you built your cluster with.
ssh-add -l
# If you need to add the key to your agent:
ssh-add path/to/private/key

# Now you can SSH into the bastion
ssh -A admin@<bastion-ELB-address>

# Where <bastion-ELB-address> is usually bastion.$clustername (bastion.example.kubernetes.cluster) unless otherwise specified

```

Now that you can successfully SSH into the bastion with a forwarded SSH agent. You can SSH into any of your cluster resources using their local IP address. You can get their local IP address from the cloud console.
