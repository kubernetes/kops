# Description of Keys in `config` and `cluster.spec`

This list is not complete but aims to document any keys that are less than self-explanatory. Our [godoc](https://godoc.org/k8s.io/kops/pkg/apis/kops) reference provides a more detailed list of API values. [ClusterSpec](https://godoc.org/k8s.io/kops/pkg/apis/kops#ClusterSpec), defined as `kind: Cluster` in YAML, and [InstanceGroup](https://godoc.org/k8s.io/kops/pkg/apis/kops#InstanceGroup), defined as `kind: InstanceGroup` in YAML, are the two top-level API values used to describe a cluster.

## spec

### api

This object configures how we expose the API:

* `dns` will allow direct access to master instances, and configure DNS to point directly to the master nodes.
* `loadBalancer` will configure a load balancer (ELB) in front of the master nodes, and configure DNS to point to the ELB.

DNS example:

```yaml
spec:
  api:
    dns: {}
```


When configuring a LoadBalancer, you can also choose to have a public ELB or an internal (VPC only) ELB.  The `type`
field should be `Public` or `Internal`.

Additionally, you can increase idle timeout of the load balancer by setting its `idleTimeoutSeconds`. The default idle timeout is 5 minutes, with a maximum of 3600 seconds (60 minutes) being allowed by AWS.
For more information see [configuring idle timeouts](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html).

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      idleTimeoutSeconds: 300
```

### sshAccess

This array configures the CIDRs that are able to ssh into nodes. On AWS this is manifested as inbound security group rules on the `nodes` and `master` security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  sshAccess:
    - 12.34.56.78/32
```

### apiAccess

This array configures the CIDRs that are able to access the kubernetes API. On AWS this is manifested as inbound security group rules on the ELB or master security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  apiAccess:
    - 12.34.56.78/32
```

### cluster.spec Subnet Keys

#### id
ID of a subnet to share in an existing VPC.

#### egress
The resource identifier (ID) of something in your existing VPC that you would like to use as "egress" to the outside world.

This feature was originally envisioned to allow re-use of NAT Gateways. In this case, the usage is as follows. Although NAT gateways are "public"-facing resources, in the Cluster spec, you must specify them in the private subnet section. One way to think about this is that you are specifying "egress", which is the default route out from this private subnet.

```
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    egress: nat-987654321
    type: Private
    zone: us-east-1a
  - cidr: 10.20.32.0/21
    name: utility-us-east-1a
    id: subnet-12345
    type: Utility
    zone: us-east-1a
```

### kubeAPIServer

This block contains configuration for the `kube-apiserver`.

#### oidc flags for Open ID Connect Tokens

Read more about this here: https://kubernetes.io/docs/admin/authentication/#openid-connect-tokens

```yaml
spec:
  kubeAPIServer:
    oidcIssuerURL: https://your-oidc-provider.svc.cluster.local
    oidcClientID: kubernetes
    oidcUsernameClaim: sub
    oidcGroupsClaim: user_roles
    oidcCAFile: /etc/kubernetes/ssl/kc-ca.pem
```

#### audit logging

Read more about this here: https://kubernetes.io/docs/admin/audit

```yaml
spec:
  kubeAPIServer:
    auditLogPath: /var/log/kube-apiserver-audit.log
    auditLogMaxAge: 10
    auditLogMaxBackups: 1
    auditLogMaxSize: 100
```

#### runtimeConfig

Keys and values here are translated into `--runtime-config` values for `kube-apiserver`, separated by commas.

Use this to enable alpha features, for example:

```yaml
spec:
  kubeAPIServer:
    runtimeConfig:
      batch/v2alpha1: "true"
      apps/v1alpha1: "true"
```

Will result in the flag `--runtime-config=batch/v2alpha1=true,apps/v1alpha1=true`. Note that `kube-apiserver` accepts `true` as a value for switch-like flags.

### kubelet

This block contains configurations for `kubelet`.  See https://kubernetes.io/docs/admin/kubelet/

NOTE: Where the corresponding configuration value can be empty, fields can be set to empty in the spec, and an empty string will be passed as the configuration value.
 ```yaml
 spec:
   kubelet:
     resolvConf: ""
```

Will result in the flag `--resolv-conf=` being built.

####  Feature Gates

```yaml
spec:
  kubelet:
    featureGates:
      ExperimentalCriticalPodAnnotation: "true"
      AllowExtTrafficLocalEndpoints: "false"
```

Will result in the flag `--feature-gates=ExperimentalCriticalPodAnnotation=true,AllowExtTrafficLocalEndpoints=false`

####  Compute Resources Reservation

```yaml
spec:
  kubelet:
    kubeReserved:
        cpu: "100m"
        memory: "100Mi"
        storage: "1Gi"
    kubeReservedCgroup: "/kube-reserved"
    systemReserved:
        cpu: "100m"
        memory: "100Mi"
        storage: "1Gi"
    systemReservedCgroup: "/system-reserved"
    enforceNodeAllocatable: "pods,system-reserved,kube-reserved"
```

Will result in the flag `--kube-reserved=cpu=100m,memory=100Mi,storage=1Gi --kube-reserved-cgroup=/kube-reserved --system-reserved=cpu=100mi,memory=100Mi,storage=1Gi --system-reserved-cgroup=/system-reserved --enforce-node-allocatable=pods,system-reserved,kube-reserved`

Learn [more about reserving compute resources](https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/).

### networkID

On AWS, this is the id of the VPC the cluster is created in. If creating a cluster from scratch, this field does not need to be specified at create time; `kops` will create a `VPC` for you.

```yaml
spec:
  networkID: vpc-abcdefg1
```

More information about running in an existing VPC is [here](run_in_existing_vpc.md).

### hooks

Hooks allow the execution of a container before the installation of Kubneretes on every node in a cluster.  For intance you can install nvidia drivers for using GPUs.

```
spec:
  # many sections removed
  hooks:
  - execContainer:
      image: kopeio/nvidia-bootstrap:1.6
```

Install Ceph

```
spec:
  # many sections removed
  hooks:
  - execContainer:
      command:
      - sh
      - -c
      - chroot /rootfs apt-get update && chroot /rootfs apt-get install -y ceph-common
      image: busybox
```

### cloudConfig

If you are using aws as `cloudProvider`, you can disable authorization of ELB security group to Kubernetes Nodes security group. In other words, it will not add security group rule.
This can be usefull to avoid AWS limit: 50 rules per security group.
```yaml
spec:
  cloudConfig:
    disableSecurityGroupIngress: true
```

### registryMirrors

If you have a bunch of Docker instances (physicsal or vm) running, each time one of them pulls an image that is not present on the host, it will fetch it from the internet (DockerHub). By caching these images, you can keep the traffic within your local network and avoid egress bandwidth usage.
This setting benefits not only cluster provisioning but also image pulling.

@see [Cache-Mirror Dockerhub For Speed](https://hackernoon.com/mirror-cache-dockerhub-locally-for-speed-f4eebd21a5ca)
@see [Configure the Docker daemon](https://docs.docker.com/registry/recipes/mirror/#configure-the-docker-daemon).

```yaml
spec:
  docker:
    registryMirrors:
    - https://registry.example.com
```

#### WARNING: this works only for Kubernetes version above 1.7.0.

For avoid to create security group per each elb, you can specify security group id, that will be assigned to your LoadBalancer. It must be security group id, not name. Also, security group must be empty, because Kubernetes will add rules per ports that are specified in service file.
This can be usefull to avoid AWS limits: 500 security groups per region and 50 rules per security group.

```yaml
spec:
  cloudConfig:
    elbSecurityGroup: sg-123445678
```
