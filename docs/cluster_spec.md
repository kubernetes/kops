# The `Cluster` resource

The `Cluster` resource contains the specification of the cluster itself.

The complete list of keys can be found at the [Cluster](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#ClusterSpec) reference page. 

On this page, we will expand on the more important configuration keys.

## api

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

You can use a valid SSL Certificate for your API Server Load Balancer. Currently, only AWS is supported:

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      sslCertificate: arn:aws:acm:<region>:<accountId>:certificate/<uuid>
```

*Openstack only*
As of Kops 1.12.0 it is possible to use the load balancer internally by setting the `useForInternalApi: true`.
This will point both `masterPublicName` and `masterInternalName` to the load balancer. You can therefore set both of these to the same value in this configuration.

```yaml
spec:
  api:
    loadBalancer:
      type: Internal
      useForInternalApi: true
```

You can also set the API load balancer to be cross-zone:
```yaml
spec:
  api:
    loadBalancer:
      crossZoneLoadBalancing: true
```

## etcdClusters

### The default etcd configuration

Kops will default to v3 using TLS by default. etcd provisioning and upgrades are handled by etcd-manager. By default, the spec looks like this:

```yaml
etcdClusters:
- etcdMembers:
  - instanceGroup: master0-az0
    name: a-1
  - instanceGroup: master1-az0
    name: a-2
  - instanceGroup: master0-az1
    name: b-1
  name: main
- etcdMembers:
  - instanceGroup: master0-az0
    name: a-1
  - instanceGroup: master1-az0
    name: a-2
  - instanceGroup: master0-az1
    name: b-1
  name: events
```

The etcd version used by kops follows the recommended etcd version for the given kubernetes version. It is possible to override this by adding the `version` key to each of the etcd clusters.

By default, the Volumes created for the etcd clusters are `gp2` and 20GB each. The volume size, type and Iops( for `io1`) can be configured via their parameters. Conversion between `gp2` and `io1` is not supported, nor are size changes.

As of Kops 1.12.0 it is also possible to modify the requests for your etcd cluster members using the `cpuRequest` and `memoryRequest` parameters.

```yaml
etcdClusters:
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
    volumeType: gp2
    volumeSize: 20
  name: main
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
    volumeType: io1
    # WARNING: bear in mind that the Iops to volume size ratio has a maximum of 50 on AWS!
    volumeIops: 100
    volumeSize: 21
  name: events
  cpuRequest: 150m
  memoryRequest: 512Mi
```

### etcd metrics

You cam expose /metrics endpoint for the etcd instances and control their type (`basic` or `extensive`) by defining env vars:

```yaml
etcdClusters:
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
  name: main
  manager:
    env:
    - name: ETCD_LISTEN_METRICS_URLS
      value: http://0.0.0.0:8081
    - name: ETCD_METRICS
      value: basic
```

### etcd backups retention

You can set the retention duration for the hourly and yearly backups by defining env vars:

```yaml
etcdClusters:
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
  name: main
  manager:
    env:
    - name: ETCD_MANAGER_HOURLY_BACKUPS_RETENTION
      value: 7d
    - name: ETCD_MANAGER_DAILY_BACKUPS_RETENTION
      value: 1y
```

## sshAccess

This array configures the CIDRs that are able to ssh into nodes. On AWS this is manifested as inbound security group rules on the `nodes` and `master` security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  sshAccess:
    - 12.34.56.78/32
```

## kubernetesApiAccess

This array configures the CIDRs that are able to access the kubernetes API. On AWS this is manifested as inbound security group rules on the ELB or master security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  kubernetesApiAccess:
    - 12.34.56.78/32
```

## cluster.spec Subnet Keys

### id
ID of a subnet to share in an existing VPC.

### egress
The resource identifier (ID) of something in your existing VPC that you would like to use as "egress" to the outside world.

This feature was originally envisioned to allow re-use of NAT gateways. In this case, the usage is as follows. Although NAT gateways are "public"-facing resources, in the Cluster spec, you must specify them in the private subnet section. One way to think about this is that you are specifying "egress", which is the default route out from this private subnet.

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

In the case that you don't use NAT gateways or internet gateways, Kops 1.12.0 introduced the "External" flag for egress to force kops to ignore egress for the subnet. This can be useful when other tools are used to manage egress for the subnet such as virtual private gateways. Please note that your cluster may need to have access to the internet upon creation, so egress must be available upon initializing a cluster. This is intended for use when egress is managed external to kops, typically with an existing cluster.

```
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    egress: External
    type: Private
    zone: us-east-1a
```

### publicIP
The IP of an existing EIP that you would like to attach to the NAT gateway.

```
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    publicIP: 203.93.148.142
    type: Private
    zone: us-east-1a
```

## kubeAPIServer

This block contains configuration for the `kube-apiserver`.

### oidc flags for Open ID Connect Tokens

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
    oidcRequiredClaim:
    	- "key=value"
```

### audit logging

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

Example policy file can be found [here](https://raw.githubusercontent.com/kubernetes/website/master/content/en/examples/audit/audit-policy.yaml)

### dynamic audit configuration

Read more about this here: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#dynamic-backend

```yaml
spec:
  kubeAPIServer:
    auditDynamicConfiguration: true
```

By enabling this feature you are allowing for auditsinks to be registered with the API server.  For information on audit sinks please read [Audit Sink](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#auditsink-v1alpha1-auditregistration).  This feature is only supported in kubernetes versions greater than 1.13.  Currently, this feature is alpha and requires enabling the feature gate and a runtime config.

**Note** For kubernetes versions greater than 1.13, this is an alpha feature that requires the API auditregistration.k8s.io/v1alpha1 to be enabled as a runtime-config option, and the feature gate DynamicAuditing to be also enabled.  The options --feature-gates=DynamicAuditing=true and --runtime-config=auditregistration.k8s.io/v1alpha1=true must be enabled on the API server in addition to this flag.  See the sections for how to enable feature gates [here](https://github.com/kubernetes/kops/blob/master/docs/cluster_spec.md#feature-gates).  See the section on how to enable alphas APIs in the runtime config [here](https://github.com/kubernetes/kops/blob/master/docs/cluster_spec.md#runtimeconfig).
Also, an audit policy should be provided in the file assets section.  If the flag is omitted, no events are logged.
You could use the [fileAssets](https://github.com/kubernetes/kops/blob/master/docs/cluster_spec.md#fileassets)  feature to push an advanced audit policy file on the master nodes.

Example policy file can be found [here](https://raw.githubusercontent.com/kubernetes/website/master/content/en/examples/audit/audit-policy.yaml)

### bootstrap tokens

Read more about this here: https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens/

```yaml
spec:
  kubeAPIServer:
    enableBootstrapTokenAuth: true
```

By enabling this feature you instructing two things;
- master nodes will bypass the bootstrap token but they _will_ build kubeconfigs with unique usernames in the system:nodes group _(this ensure's the master nodes confirm with the node authorization mode https://kubernetes.io/docs/reference/access-authn-authz/node/)_
- secondly the nodes will be configured to use a bootstrap token located by default at `/var/lib/kubelet/bootstrap-kubeconfig` _(though this can be override in the kubelet spec)_. The nodes will sit the until a bootstrap file is created and once available attempt to provision the node.

**Note** enabling bootstrap tokens does not provision bootstrap tokens for the worker nodes. Under this configuration it is assumed a third-party process is provisioning the tokens on behalf of the worker nodes. For the full setup please read [Node Authorizer Service](node_authorization.md)

### Max Requests Inflight

The maximum number of non-mutating requests in flight at a given time. When the server exceeds this, it rejects requests. Zero for no limit. (default 400)

```yaml
spec:
  kubeAPIServer:
    maxRequestsInflight: 1000
```

The maximum number of mutating requests in flight at a given time. When the server exceeds this, it rejects requests. Zero for no limit. (default 200)

```yaml
spec:
  kubeAPIServer:
    maxMutatingRequestsInflight: 450
```
### Profiling

Profiling via web interface `host:port/debug/pprof/`. (default: true)

```yaml
spec:
  kubeAPIServer:
    enableProfiling: false
```

### runtimeConfig

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

### serviceNodePortRange

This value is passed as `--service-node-port-range` for `kube-apiserver`.

```yaml
spec:
  kubeAPIServer:
    serviceNodePortRange: 30000-33000
```

### Disable Basic Auth

Support for basic authentication was removed in Kubernetes 1.19. For previous versions
of Kubernetes this will disable the passing of the `--basic-auth-file` flag when:

```yaml
spec:
  kubeAPIServer:
    disableBasicAuth: true
```

### targetRamMb

Memory limit for apiserver in MB (used to configure sizes of caches, etc.)

```yaml
spec:
  kubeAPIServer:
    targetRamMb: 4096
```

### eventTTL

How long API server retains events. Note that you must fill empty units of time with zeros.

```yaml
spec:
  kubeAPIServer:
    eventTTL: 03h0m0s
```

## externalDns

This block contains configuration options for your `external-DNS` provider.
The current external-DNS provider is the kops `dns-controller`, which can set up DNS records for Kubernetes resources.
`dns-controller` is scheduled to be phased out and replaced with `external-dns`.

```yaml
spec:
  externalDns:
    watchIngress: true
```

Default _kops_ behavior is false. `watchIngress: true` uses the default _dns-controller_ behavior which is to watch the ingress controller for changes. Set this option at risk of interrupting Service updates in some cases.

## kubelet

This block contains configurations for `kubelet`.  See https://kubernetes.io/docs/admin/kubelet/

NOTE: Where the corresponding configuration value can be empty, fields can be set to empty in the spec, and an empty string will be passed as the configuration value.
 ```yaml
 spec:
   kubelet:
     resolvConf: ""
```

Will result in the flag `--resolv-conf=` being built.

### Disable CPU CFS Quota
To disable CPU CFS quota enforcement for containers that specify CPU limits (default true) we have to set the flag `--cpu-cfs-quota` to `false`
on all the kubelets. We can specify that in the `kubelet` spec in our cluster.yml.

```
spec:
  kubelet:
    cpuCFSQuota: false
```

### Configure CPU CFS Period
Configure CPU CFS quota period value (cpu.cfs_period_us). Example:

```
spec:
  kubelet:
    cpuCFSQuotaPeriod: "100ms"
```

### Enable Custom metrics support
To use custom metrics in kubernetes as per [custom metrics doc](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-custom-metrics)
we have to set the flag `--enable-custom-metrics` to `true` on all the kubelets. We can specify that in the `kubelet` spec in our cluster.yml.

```
spec:
  kubelet:
    enableCustomMetrics: true
```

### Setting kubelet CPU management policies
Kops 1.12.0 added support for enabling cpu management policies in kubernetes as per [cpu management doc](https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#cpu-management-policies)
we have to set the flag `--cpu-manager-policy` to the appropriate value on all the kubelets. This must be specified in the `kubelet` spec in our cluster.yml.

```
spec:
  kubelet:
    cpuManagerPolicy: static
```

### Setting kubelet configurations together with the Amazon VPC backend
Setting kubelet configurations together with the networking Amazon VPC backend requires to also set the `cloudProvider: aws` setting in this block. Example:

```yaml
spec:
  kubelet:
    enableCustomMetrics: true
    cloudProvider: aws
...
...
  cloudProvider: aws
...
...
  networking:
    amazonvpc: {}
```

### Configure a Flex Volume plugin directory
An optional flag can be provided within the KubeletSpec to set a volume plugin directory (must be accessible for read/write operations), which is additionally provided to the Controller Manager and mounted in accordingly.

Kops will set this for you based off the Operating System in use:
- ContainerOS: `/home/kubernetes/flexvolume/`
- CoreOS: `/var/lib/kubelet/volumeplugins/`
- Default (in-line with upstream k8s): `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/`

If you wish to override this value, it can be done so with the following addition to the kubelet spec:
```yaml
spec:
  kubelet:
    volumePluginDirectory: /provide/a/writable/path/here
```

### Protect Kernel Defaults

Default kubelet behaviour for kernel tuning. If set, kubelet errors if any of kernel tunables is different than kubelet defaults.

```yaml
spec:
  kubelet:
    protectKernelDefaults: true
```

## kubeScheduler

This block contains configurations for `kube-scheduler`.  See https://kubernetes.io/docs/admin/kube-scheduler/

```yaml
spec:
  kubeScheduler:
    usePolicyConfigMap: true
    enableProfiling: false
```

Will make kube-scheduler use the scheduler policy from configmap "scheduler-policy" in namespace kube-system.

Note that as of Kubernetes 1.8.0 kube-scheduler does not reload its configuration from configmap automatically. You will need to ssh into the master instance and restart the Docker container manually.

## kubeDNS

This block contains configurations for `kube-dns`.

 ```yaml
 spec:
   kubeDNS:
     provider: KubeDNS
```

Specifying KubeDNS will install kube-dns as the default service discovery.

 ```yaml
 spec:
   kubeDNS:
     provider: CoreDNS
```

This will install [CoreDNS](https://coredns.io/) instead of kube-dns.

If you are using CoreDNS and want to use an entirely custom CoreFile you can do this by specifying the file. This will not work with any other options which interact with the default CoreFile. You can also override the version of the CoreDNS image used to use a different registry or version by specifying `CoreDNSImage`.

**Note:** If you are using this functionality you will need to be extra vigilant on version changes of CoreDNS for changes in functionality of the plugins being used etc.

```yaml
spec:
  kubeDNS:
    provider: CoreDNS
    coreDNSImage: mirror.registry.local/mirrors/coredns:1.3.1
    externalCoreFile: |
      amazonaws.com:53 {
            errors
            log . {
                class denial error
            }
            health :8084
            prometheus :9153
            proxy . 169.254.169.253 {
            }
            cache 30
        }
        .:53 {
            errors
            health :8080
            autopath @kubernetes
            kubernetes cluster.local {
                pods verified
                upstream 169.254.169.253
                fallthrough in-addr.arpa ip6.arpa
            }
            prometheus :9153
            proxy . 169.254.169.253
            cache 300
        }
```

**Note:** If you are upgrading to CoreDNS, kube-dns will be left in place and must be removed manually (you can scale the kube-dns and kube-dns-autoscaler deployments in the `kube-system` namespace to 0 as a starting point). The `kube-dns` Service itself should be left in place, as this retains the ClusterIP and eliminates the possibility of DNS outages in your cluster. If you would like to continue autoscaling, update the `kube-dns-autoscaler` Deployment container command for `--target=Deployment/kube-dns` to be `--target=Deployment/coredns`.

## Node local DNS cache

If you are using CoreDNS, you can enable NodeLocal DNSCache. It is used to improve improve the Cluster DNS performance by running a dns caching agent on cluster nodes as a DaemonSet.

```yaml
spec:
  kubeDNS:
    provider: CoreDNS
    nodeLocalDNS:
      enabled: true
```

If you are using kube-proxy in ipvs mode or Cilium as CNI, you have to set the nodeLocalDNS as ClusterDNS.

```yaml
spec:
  kubelet:
    clusterDNS: 169.254.20.10
  masterKubelet:
    clusterDNS: 169.254.20.10
```

## kubeControllerManager
This block contains configurations for the `controller-manager`.

```yaml
spec:
  kubeControllerManager:
    horizontalPodAutoscalerSyncPeriod: 15s
    horizontalPodAutoscalerDownscaleDelay: 5m0s
    horizontalPodAutoscalerDownscaleStabilization: 5m
    horizontalPodAutoscalerUpscaleDelay: 3m0s
    horizontalPodAutoscalerTolerance: 0.1
    experimentalClusterSigningDuration: 8760h0m0s
    enableProfiling: false
```

For more details on `horizontalPodAutoscaler` flags see the [official HPA docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) and the [Kops guides on how to set it up](horizontal_pod_autoscaling.md).

###  Feature Gates

Feature gates can be configured on the kubelet.

```yaml
spec:
  kubelet:
    featureGates:
      Accelerators: "true"
      AllowExtTrafficLocalEndpoints: "false"
```

The above will result in the flag `--feature-gates=Accelerators=true,AllowExtTrafficLocalEndpoints=false` being added to the kubelet.

Some feature gates also require the `featureGates` setting on other components. For example`PodShareProcessNamespace` requires
the feature gate to be enabled also on the api server:

```yaml
spec:
  kubelet:
    featureGates:
      PodShareProcessNamespace: "true"
  kubeAPIServer:
    featureGates:
      PodShareProcessNamespace: "true"
```

For more information, see the [feature gate documentation](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/)

###  Compute Resources Reservation

```yaml
spec:
  kubelet:
    kubeReserved:
        cpu: "100m"
        memory: "100Mi"
        ephemeral-storage: "1Gi"
    kubeReservedCgroup: "/kube-reserved"
    systemReserved:
        cpu: "100m"
        memory: "100Mi"
        ephemeral-storage: "1Gi"
    systemReservedCgroup: "/system-reserved"
    enforceNodeAllocatable: "pods,system-reserved,kube-reserved"
```

Will result in the flag `--kube-reserved=cpu=100m,memory=100Mi,ephemeral-storage=1Gi --kube-reserved-cgroup=/kube-reserved --system-reserved=cpu=100m,memory=100Mi,ephemeral-storage=1Gi --system-reserved-cgroup=/system-reserved --enforce-node-allocatable=pods,system-reserved,kube-reserved`

Learn [more about reserving compute resources](https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/).

## networkID

On AWS, this is the id of the VPC the cluster is created in. If creating a cluster from scratch, this field does not need to be specified at create time; `kops` will create a `VPC` for you.

```yaml
spec:
  networkID: vpc-abcdefg1
```

More information about running in an existing VPC is [here](run_in_existing_vpc.md).

## hooks

Hooks allow for the execution of an action before the installation of Kubernetes on every node in a cluster. For instance you can install Nvidia drivers for using GPUs. This hooks can be in the form of Docker images or manifest files (systemd units). Hooks can be placed in either the cluster spec, meaning they will be globally deployed, or they can be placed into the instanceGroup specification. Note: service names on the instanceGroup which overlap with the cluster spec take precedence and ignore the cluster spec definition, i.e. if you have a unit file 'myunit.service' in cluster and then one in the instanceGroup, only the instanceGroup is applied.

When creating a systemd unit hook using the `manifest` field, the hook system will construct a systemd unit file for you. It creates the `[Unit]` section, adding an automated description and setting `Before` and `Requires` values based on the `before` and `requires` fields. The value of the `manifest` field is used as the `[Service]` section of the unit file. To override this behavior, and instead specify the entire unit file yourself, you may specify `useRawManifest: true`. In this case, the contents of the `manifest` field will be used as a systemd unit, unmodified. The `before` and `requires` fields may not be used together with `useRawManifest`.

```
spec:
  # many sections removed

  # run a docker container as a hook
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

  # or construct a systemd unit
  hooks:
  - name: iptable-restore.service
    roles:
    - Node
    - Master
    before:
    - kubelet.service
    manifest: |
      EnvironmentFile=/etc/environment
      # do some stuff

  # or use a raw systemd unit
  hooks:
  - name: iptable-restore.service
    roles:
    - Node
    - Master
    useRawManifest: true
    manifest: |
      [Unit]
      Description=Restore iptables rules
      Before=kubelet.service
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

Install cachefilesd

```
spec:
  # many sections removed
  hooks:
  - before:
    - kubelet.service
    manifest: |
      Type=oneshot
      ExecStart=/sbin/modprobe cachefiles
    name: cachefiles.service
  - execContainer:
      command:
      - sh
      - -c
      - chroot /rootfs apt-get update && chroot /rootfs apt-get install -y cachefilesd
        && chroot /rootfs sed -i s/#RUN/RUN/ /etc/default/cachefilesd && chroot /rootfs
        service cachefilesd restart
      image: busybox
```

## fileAssets

FileAssets permits you to place inline file content into the cluster and instanceGroup specification. This is useful for deploying additional configuration files that kubernetes components requires, such as auditlogs or admission controller configurations.

```yaml
spec:
  fileAssets:
  - name: iptable-restore
    # Note if not path is specified the default path it /srv/kubernetes/assets/<name>
    path: /var/lib/iptables/rules-save
    roles: [Master,Node,Bastion] # a list of roles to apply the asset to, zero defaults to all
    content: |
      some file content
```


## cloudConfig

### disableSecurityGroupIngress
If you are using aws as `cloudProvider`, you can disable authorization of ELB security group to Kubernetes Nodes security group. In other words, it will not add security group rule.
This can be useful to avoid AWS limit: 50 rules per security group.
```yaml
spec:
  cloudConfig:
    disableSecurityGroupIngress: true
```

### elbSecurityGroup

To avoid creating a security group per elb, you can specify security group id, that will be assigned to your LoadBalancer. It must be security group id, not name.
`api.loadBalancer.additionalSecurityGroups` must be empty, because Kubernetes will add rules per ports that are specified in service file.
This can be useful to avoid AWS limits: 500 security groups per region and 50 rules per security group.

```yaml
spec:
  cloudConfig:
    elbSecurityGroup: sg-123445678
```

## containerRuntime

Alternative [container runtimes](https://kubernetes.io/docs/setup/production-environment/container-runtimes/) can be used to run Kubernetes. Docker is still the default container runtime, but [containerd](https://kubernetes.io/blog/2018/05/24/kubernetes-containerd-integration-goes-ga/) can also be selected.

```yaml
spec:
  containerRuntime: containerd
```

## containerd

It is possible to override the [containerd](https://github.com/containerd/containerd/blob/master/README.md) daemon options for all the nodes in the cluster. See the [API docs](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#ContainerdConfig) for the full list of options.

```yaml
spec:
  containerd:
    version: 1.3.3
    logLevel: info
    configOverride: ""
```

## docker

It is possible to override Docker daemon options for all masters and nodes in the cluster. See the [API docs](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#DockerConfig) for the full list of options.

### registryMirrors

If you have a bunch of Docker instances (physical or vm) running, each time one of them pulls an image that is not present on the host, it will fetch it from the internet (DockerHub). By caching these images, you can keep the traffic within your local network and avoid egress bandwidth usage.
This setting benefits not only cluster provisioning but also image pulling.

@see [Cache-Mirror Dockerhub For Speed](https://hackernoon.com/mirror-cache-dockerhub-locally-for-speed-f4eebd21a5ca)
@see [Configure the Docker daemon](https://docs.docker.com/registry/recipes/mirror/#configure-the-docker-daemon).

```yaml
spec:
  docker:
    registryMirrors:
    - https://registry.example.com
```

### Skip Install

If you want nodeup to skip the Docker installation tasks, you can do so with:

```yaml
spec:
  docker:
    skipInstall: true
```

**NOTE:** When this field is set to `true`, it is entirely up to the user to install and configure Docker.

### storage

The Docker [Storage Driver](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-storage-driver) can be specified in order to override the default. Be sure the driver you choose is supported by your operating system and docker version.

```yaml
docker:
  storage: devicemapper
  storageOpts:
    - "dm.thinpooldev=/dev/mapper/thin-pool"
    - "dm.use_deferred_deletion=true"
    - "dm.use_deferred_removal=true"
```

## sshKeyName

In some cases, it may be desirable to use an existing AWS SSH key instead of allowing kops to create a new one.
Providing the name of a key already in AWS is an alternative to `--ssh-public-key`.

```yaml
spec:
  sshKeyName: myexistingkey
```

If you want to create your instance without any SSH keys you can set this to an empty string:
```yaml
spec:
  sshKeyName: ""
```

## useHostCertificates

Self-signed certificates towards Cloud APIs. In some cases Cloud APIs do have self-signed certificates.

```yaml
spec:
  useHostCertificates: true
```

### Optional step: add root certificates to instancegroups root ca bundle

```yaml
  additionalUserData:
  - name: cacert.sh
    type: text/x-shellscript
    content: |
      #!/bin/sh
      cat > /usr/local/share/ca-certificates/mycert.crt <<EOF
      -----BEGIN CERTIFICATE-----
snip
      -----END CERTIFICATE-----
      EOF
      update-ca-certificates
```

**NOTE**: `update-ca-certificates` is command for debian/ubuntu. That command is different depending your OS.

## target

In some use-cases you may wish to augment the target output with extra options.  `target` supports a minimal amount of options you can do this with.  Currently only the terraform target supports this, but if other use cases present themselves, kops may eventually support more.

```yaml
spec:
  target:
    terraform:
      providerExtraConfig:
        alias: foo
```

## assets

Assets define alternative locations from where to retrieve static files and containers

### containerRegistry

The container registry enables kops / kubernetes to pull containers from a managed registry.
This is useful when pulling containers from the internet is not an option, eg. because the
deployment is offline / internet restricted or because of special requirements that apply
for deployed artifacts, eg. auditing of containers.

For a use case example, see [How to use kops in AWS China Region](https://github.com/kubernetes/kops/blob/master/docs/aws-china.md)

```yaml
spec:
  assets:
    containerRegistry: example.com/registry
```


### containerProxy

The container proxy is designed to acts as a [pull through cache](https://docs.docker.com/registry/recipes/mirror/) for docker container assets.
Basically, what it does is it remaps the Kubernetes image URL to point to your cache so that the docker daemon will pull the image from that location.
If, for example, the containerProxy is set to `proxy.example.com`, the image `k8s.gcr.io/kube-apiserver` will be pulled from `proxy.example.com/kube-apiserver` instead.
Note that the proxy you use has to support this feature for private registries.


```yaml
spec:
  assets:
    containerProxy: proxy.example.com
```

## sysctlParameters

To add custom kernel runtime parameters to your all instance groups in the
cluster, specify the `sysctlParameters` field as an array of strings. Each
string must take the form of `variable=value` the way it would appear in
sysctl.conf (see also `sysctl(8)` manpage).

You could also use the `sysctlParameters` field on [the instance group](https://github.com/kubernetes/kops/blob/master/docs/instance_groups.md#setting-custom-kernel-runtime-parameters) to specify different parameters for each instance group.

Unlike a simple file asset, specifying kernel runtime parameters in this manner
would correctly invoke `sysctl --system` automatically for you to apply said
parameters.

For example:

```yaml
spec:
  sysctlParameters:
    - fs.pipe-user-pages-soft=524288
    - net.ipv4.tcp_keepalive_time=200
```

which would end up in a drop-in file on all masters and nodes of the cluster.
