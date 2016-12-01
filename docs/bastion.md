# Bastion in Kops

Bastion provide an external facing point of entry into a network containing private network instances. This host can provide a single point of fortification or audit and can be started and stopped to enable or disable inbound SSH communication from the Internet, some call bastion as the "jump server".

Note: Bastion will get setup for the cluster(by default) only when `--topology="private"`.

* [More information on bastion from aws](http://docs.aws.amazon.com/quickstart/latest/linux-bastion/architecture.html)
* [More information on bastion from gce](https://cloud.google.com/solutions/connecting-securely#bastion)

## AWS

### Specify instance type of bastion

Instance types in AWS comprise varying combinations of CPU, memory, storage, and networking capacity and give you the flexibility to choose the appropriate mix of resources for your applications.

- Bastion Instance type can be modified using `kops edit cluster`
- Defaults to `t2.medium`
#### TODO: add the example below for configuration
```

```
[More information](https://aws.amazon.com/ec2/instance-types/)


### Turn on/off bastion

To turn on/off bastion host setup completely.
- **Defaults** to `false` if the topology selected is `public`
- **Defaults** to `true` if the topology selected is `private`

```
    kops create cluster --bastion=[true|false]
```

### Reach bastion from outside of vpc using a name

- **Default:** CNAME for the bastion is only created when the user explicitly define it using `kops edit cluster`
- **Configure:** Bastion friendly CNAME can be configured using `kops edit cluster`
#### TODO: add the example below for configuration
```

```

### High idle timeout for bastion ASG's ELB. (Configurable LoadBalancer Attributes)

By default, elastic load balancing sets the idle timeout to `60` seconds.
- **Default:** Bastion ELB in kops will have `120` seconds as their default timeout.
- **Configure:** This value can be configured using `kops edit cluster`
#### TODO: add the example below for configuration
```

```
[More information](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html)

### Improve access to bastion instances

Current: `ssh to the bastion: ssh -i ~/.ssh/id_rsa admin@api.mydomain.com`
