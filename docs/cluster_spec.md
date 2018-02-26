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

Also, you can add precreated additional security groups to the load balancer by setting `additionalSecurityGroups`.

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      additionalSecurityGroups:
      - sg-xxxxxxxx
      - sg-xxxxxxxx
```

Additionally, you can increase idle timeout of the load balancer by setting its `idleTimeoutSeconds`. The default idle timeout is 5 minutes, with a maximum of 3600 seconds (60 minutes) being allowed by AWS.
For more information see [configuring idle timeouts](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html).

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      idleTimeoutSeconds: 300
```
Additionally, you can increase idle timeout of the load balancer by setting its `idleTimeoutSeconds`. The default idle timeout is 5 minutes, with a maximum of 3600 seconds (60 minutes) being allowed by AWS.
For more information see [configuring idle timeouts](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html).

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      idleTimeoutSeconds: 300
```

### etcdClusters v3 & tls

Although kops doesn't presently default to etcd3, it is possible to turn on both v3 and TLS authentication for communication amongst cluster members. These options may be enabled via the cluster spec (manifests only i.e. no command line options as yet). An upfront warning; at present no upgrade path exists for migrating from v2 to v3 so **DO NOT** try to enable this on a v2 running cluster as it must be done on cluster creation. The below example snippet assumes a HA cluster of three masters.

```yaml
etcdClusters:
- etcdMembers:
  - instanceGroup: master0-az0
    name: a-1
  - instanceGroup: master1-az0
    name: a-2
  - instanceGroup: master0-az1
    name: b-1
  enableEtcdTLS: true
  name: main
  version: 3.0.17
- etcdMembers:
  - instanceGroup: master0-az0
    name: a-1
  - instanceGroup: master1-az0
    name: a-2
  - instanceGroup: master0-az1
    name: b-1
  enableEtcdTLS: true
  name: events
  version: 3.0.17
```

### sshAccess

This array configures the CIDRs that are able to ssh into nodes. On AWS this is manifested as inbound security group rules on the `nodes` and `master` security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  sshAccess:
    - 12.34.56.78/32
```

### kubernetesApiAccess

This array configures the CIDRs that are able to access the kubernetes API. On AWS this is manifested as inbound security group rules on the ELB or master security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  kubernetesApiAccess:
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
    oidcUsernamePrefix: "oidc:"
    oidcGroupsClaim: user_roles
    oidcGroupsPrefix: "oidc:"
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
    auditPolicyFile: /srv/kubernetes/audit.yaml
```

**Note**: The auditPolicyFile is needed. If the flag is omitted, no events are logged.

You could use the [fileAssets](https://github.com/kubernetes/kops/blob/master/docs/cluster_spec.md#fileassets)  feature to push an advanced audit policy file on the master nodes.

Example policy file can be found [here]( https://raw.githubusercontent.com/kubernetes/website/master/docs/tasks/debug-application-cluster/audit-policy.yaml)

#### Max Requests Inflight 

The maximum number of non-mutating requests in flight at a given time. When the server exceeds this, it rejects requests. Zero for no limit. (default 400)

```yaml
spec:
  kubeAPIServer:
    maxRequestsInflight: 1000
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

#### serviceNodePortRange

This value is passed as `--service-node-port-range` for `kube-apiserver`.

```yaml
spec:
  kubeAPIServer:
    serviceNodePortRange: 30000-33000
```

### externalDns

This block contains configuration options for your `external-DNS` provider.
The current external-DNS provider is the kops `dns-controller`, which can set up DNS records for Kubernetes resources.
`dns-controller` is scheduled to be phased out and replaced with `external-dns`.

```yaml
spec:
  externalDns:
    watchIngress: true
```

Default _kops_ behavior is false. `watchIngress: true` uses the default _dns-controller_ behavior which is to watch the ingress controller for changes. Set this option at risk of interrupting Service updates in some cases.

### kubelet

This block contains configurations for `kubelet`.  See https://kubernetes.io/docs/admin/kubelet/

NOTE: Where the corresponding configuration value can be empty, fields can be set to empty in the spec, and an empty string will be passed as the configuration value.
 ```yaml
 spec:
   kubelet:
     resolvConf: ""
```

Will result in the flag `--resolv-conf=` being built.

#### Enable Custom metrics support
To use custom metrics in kubernetes as per [custom metrics doc](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-custom-metrics)
we have to set the flag `--enable-custom-metrics` to `true` on all the kubelets. We can specify that in the `kubelet` spec in our cluster.yml.

```
spec:
  kubelet:
    enableCustomMetrics: true
```

### kubeScheduler

This block contains configurations for `kube-scheduler`.  See https://kubernetes.io/docs/admin/kube-scheduler/

 ```yaml
 spec:
   kubeScheduler:
     usePolicyConfigMap: true
```

Will make kube-scheduler use the scheduler policy from configmap "scheduler-policy" in namespace kube-system.

Note that as of Kubernetes 1.8.0 kube-scheduler does not reload its configuration from configmap automatically. You will need to ssh into the master instance and restart the Docker container manually.

### kubeControllerManager
This block contains configurations for the `controller-manager`.

```yaml
spec:
  kubeControllerManager:
    horizontalPodAutoscalerSyncPeriod: 15s
    horizontalPodAutoscalerDownscaleDelay: 5m0s
    horizontalPodAutoscalerUpscaleDelay: 3m0s
```

For more details on `horizontalPodAutoscaler` flags see the [HPA docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).

####  Feature Gates

```yaml
spec:
  kubelet:
    featureGates:
      Accelerators: "true"
      AllowExtTrafficLocalEndpoints: "false"
```

Will result in the flag `--feature-gates=Accelerators=true,AllowExtTrafficLocalEndpoints=false`

NOTE: Feature gate `ExperimentalCriticalPodAnnotation` is enabled by default because some critical components like `kube-proxy` depend on its presence.

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

Hooks allow for the execution of an action before the installation of Kubernetes on every node in a cluster.  For instance you can install Nvidia drivers for using GPUs. This hooks can be in the form of Docker images or manifest files (systemd units). Hooks can be placed in either the cluster spec, meaning they will be globally deployed, or they can be placed into the instanceGroup specification. Note: service names on the instanceGroup which overlap with the cluster spec take precedence and ignore the cluster spec definition, i.e. if you have a unit file 'myunit.service' in cluster and then one in the instanceGroup, only the instanceGroup is applied.

```
spec:
  # many sections removed
  hooks:
  - before:
    - some_service.service
    requires:
    - docker.service
      execContainer:
      image: kopeio/nvidia-bootstrap:1.6
      # these are added as -e to the docker environment
      environment:
        AWS_REGION: eu-west-1
        SOME_VAR: SOME_VALUE

  # or a raw systemd unit
  hooks:
  - name: iptable-restore.service
    roles:
    - Node
    - Master
    before:
    - kubelet.service
    manifest: |
      [Service]
      EnvironmentFile=/etc/environment
      # do some stuff

  # or disable a systemd unit
  hooks:
  - name: update-engine.service
    disabled: true

  # or you could wrap this into a full unit
  hooks:
  - name: disable-update-engine.service
    before:
    - update-engine.service
    manifest: |
      Type=oneshot
      ExecStart=/usr/bin/systemctl stop update-engine.service
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

### fileAssets

FileAssets is an alpha feature which permits you to place inline file content into the cluster and instanceGroup specification. It's desiginated as alpha as you can probably do this via kubernetes daemonsets as an alternative.

```yaml
spec:
  fileAssets:
  - name: iptable-restore
    # Note if not path is specificied the default path it /srv/kubernetes/assets/<name>
    path: /var/lib/iptables/rules-save
    roles: [Master,Node,Bastion] # a list of roles to apply the asset to, zero defaults to all
    content: |
      some file content
```


### cloudConfig

#### disableSecurityGroupIngress
If you are using aws as `cloudProvider`, you can disable authorization of ELB security group to Kubernetes Nodes security group. In other words, it will not add security group rule.
This can be usefull to avoid AWS limit: 50 rules per security group.
```yaml
spec:
  cloudConfig:
    disableSecurityGroupIngress: true
```

#### elbSecurityGroup
*WARNING: this works only for Kubernetes version above 1.7.0.*

To avoid creating a security group per elb, you can specify security group id, that will be assigned to your LoadBalancer. It must be security group id, not name.
`api.loadBalancer.additionalSecurityGroups` must be empty, because Kubernetes will add rules per ports that are specified in service file.
This can be useful to avoid AWS limits: 500 security groups per region and 50 rules per security group.

```yaml
spec:
  cloudConfig:
    elbSecurityGroup: sg-123445678
```

### docker

It is possible to override Docker daemon options for all masters and nodes in the cluster. See the [API docs](https://godoc.org/k8s.io/kops/pkg/apis/kops#DockerConfig) for the full list of options.

#### registryMirrors

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

#### storage

The Docker [Storage Driver](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-storage-driver) can be specified in order to override the default. Be sure the driver you choose is supported by your operating system and docker version.

```yaml
docker:
  storage: devicemapper
  storageOpts:
    - "dm.thinpooldev=/dev/mapper/thin-pool"
    - "dm.use_deferred_deletion=true"
    - "dm.use_deferred_removal=true"
```

### sshKeyName

In some cases, it may be desirable to use an existing AWS SSH key instead of allowing kops to create a new one.
Providing the name of a key already in AWS is an alternative to `--ssh-public-key`.

```yaml
spec:
  sshKeyName: myexistingkey
```

### target

In some use-cases you may wish to augment the target output with extra options.  `target` supports a minimal amount of options you can do this with.  Currently only the terraform target supports this, but if other use cases present themselves, kops may eventually support more.

```yaml
spec:
  target:
    terraform:
      providerExtraConfig:
        alias: foo
```

### Custom Security Groups

This is an aplpha feature.  Currently security group rules may be updated.

If you are unable to have kops manage its own security groups, API values exist to re-use security groups.  The configuration of the security groups is not trivial, so we do recommend having kops create its own sercurity groups.  If you must re-use or utilize other security groups, we recommend having kops export terraform which will provide a starting point for the security groups. All kops cluster manifest values are lists, which
will accept multiple security groups.


#### Prerequisites

In order to re-use security groups, kops must be running inside of a shared VPC.  At this point shared security groups are only available in AWS.  Security groups must be created before kops is run.

`kops` requires both master and node security groups to be defined, in every use case when re-using security groups.  If using an ELB for the API, a security group is aways required.  If using bastions, and ELB for that, security groups is always required.

`kops` creates security groups with rules reference each other, because of this `kops` is unable to re-use a value for say a node security group.

#### API Manifest Examples

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      # API ELB loadbalancer list of security group id exists in the loadBalancer section
      securityGroups:
      - sg-ca130742
  # securityGroups API definition includes the following element
  # these api manifest elements allow for the setting of a list of
  # sg ids
  securityGroups:
    # always required
    masterGroups:
    - sg-df180cab
    # always required
    nodeGroups:
    - sg-2c1e0a5q
    # required if using a bastion with private topology
    bastionGroups:
    - sg-a3110542
    # required if using a bastion with private topology
    bastionELBGroups:
    - sg-dd1004af
```
