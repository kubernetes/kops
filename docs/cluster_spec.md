# The `Cluster` resource

The `Cluster` resource contains the specification of the cluster itself.

The complete list of keys can be found at the [Cluster](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#ClusterSpec) reference page.

On this page, we will expand on the more important configuration keys.

The documentation for the optional addons can be found on the [addons page](/addons)

## api

This object configures how we expose the API:

* `dns` will allow direct access to master instances, and configure DNS to point directly to the master nodes.
* `loadBalancer` will configure a load balancer in front of the master nodes and configure DNS to point to the it.

DNS example:

```yaml
spec:
  api:
    dns: {}
```


When configuring a LoadBalancer, you can also choose to have a public load balancer or an internal (VPC only) load balancer. The `type` field should be `Public` or `Internal`.

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

Additionally, you can increase idle timeout of the load balancer by setting its `idleTimeoutSeconds`. The default idle timeout is 5 minutes, with a maximum of 3600 seconds (60 minutes) being allowed by AWS. Note this value is ignored for load balancer Class `Network`.
For more information see [configuring idle timeouts](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/config-idle-timeout.html).

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      idleTimeoutSeconds: 300
```

You can use a valid SSL Certificate for your API Server Load Balancer. Currently, only AWS is supported.

Also, you can change listener's [security policy](https://docs.aws.amazon.com/sdk-for-go/api/service/elbv2/#CreateListenerInput) by `sslPolicy`. Currently, only AWS Network Load Balancer is supported.

Note that when using `sslCertificate`, client certificate authentication, such as with the credentials generated via `kOps export kubecfg`, will not work through the load balancer. As of kOps 1.19, a `kubecfg` that bypasses the load balancer may be created with the `--internal` flag to `kops update cluster` or `kOps export kubecfg`. Security groups may need to be opened to allow access from the clients to the master instances' port TCP/443, for example by using the `additionalSecurityGroups` field on the master instance groups.

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      sslCertificate: arn:aws:acm:<region>:<accountId>:certificate/<uuid>
      sslPolicy: ELBSecurityPolicy-TLS-1-2-2017-01
```

*Openstack only*
As of kOps 1.12.0 it is possible to use the load balancer internally by setting the `useForInternalApi: true`.
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

### Load Balancer Class

**AWS only**

{{ kops_feature_table(kops_added_default='1.19') }}

You can choose to have a Network Load Balancer instead of a Classic Load Balancer. The `class` field should be either `Network` or `Classic` (default).

**Note**: changing the class of load balancer in an existing cluster is a disruptive operation for the control plane. Until the masters have gone through a rolling update, new connections to the apiserver will fail due to the old masters' TLS certificates containing the old load balancer's IP addresses.
```yaml
spec:
  api:
    loadBalancer:
      class : Network
      type: Public
```

### Load Balancer Subnet configuration

**AWS only**

By default, kops will try to choose one suitable subnet per availability zone and use these for the API load balancer.
Depending on the `type`, kops will choose from either `Private` or `Public` subnets. If this default logic is not
suitable for you (e.g. because you have a more granular separation between subnets), you can explicitly configure
the to-be-use subnets:

```yaml
spec:
  api:
    loadBalancer:
      type: Public
      subnets:
        - name: subnet-a
        - name: subnet-b
        - name: subnet-c
````

It is only allowed to add more subnets and forbidden to remove existing ones. This is due to limitations on AWS
ELBs and NLBs.

If the `type` is `Internal` and the `class` is `Network`, you can also specify a static private IPv4 address per subnet:
```yaml
spec:
  api:
    loadBalancer:
      type: Internal
      subnets:
        - name: subnet-a
          privateIPv4Address: 172.16.1.10
```

The specified IPv4 addresses must be part of the subnets CIDR. They can not be changed after initial deployment.

If the `type` is `Public` and the `class` is `Network`, you can also specify an Elastic IP allocationID to bind a fixed public IP address per subnet. Pleae note only IPv4 addresses have been tested:
```yaml
spec:
  api:
    loadBalancer:
      type: Public
      subnets:
        - name: utility-subnet-a
          allocationId: eipalloc-222ghi789
```

The specified Allocation ID's must already be created manually or external infrastructure as code, eg Terraform. You will need to place the loadBalanacer in the utility subnets for external connectivity.

If you made a mistake or need to change subnets for any other reason, you're currently forced to manually delete the
underlying ELB/NLB and re-run `kops update`.

## etcdClusters

### The default etcd configuration

kOps will default to v3 using TLS by default. etcd provisioning and upgrades are handled by etcd-manager. By default, the spec looks like this:

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

The etcd version used by kOps follows the recommended etcd version for the given kubernetes version. It is possible to override this by adding the `version` key to each of the etcd clusters.

By default, the Volumes created for the etcd clusters are `gp3` and 20GB each. The volume size, type (`gp2`, `gp3`, `io1`, `io2`), iops( for `io1`, `io2`, `gp3`) and throughput (`gp3`) can be configured via their parameters.

As of kOps 1.12.0 it is also possible to modify the requests for your etcd cluster members using the `cpuRequest` and `memoryRequest` parameters.

```yaml
etcdClusters:
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
    volumeType: gp3
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
{{ kops_feature_table(kops_added_default='1.18') }}

You can expose /metrics endpoint for the etcd instances and control their type (`basic` or `extensive`) by defining env vars:

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

*Note:* If you are running multiple etcd clusters you need to expose the metrics on different ports for each cluster as etcd is running as a service on the master nodes.

### etcd backups interval
{{ kops_feature_table(kops_added_default='1.24.1') }}

You can set the interval between backups using the `backupInterval` parameter:

```yaml
etcdClusters:
- etcdMembers:
  - instanceGroup: master-us-east-1a
    name: a
  name: main
  manager:
    backupInterval: 1h
```

### etcd backups retention
{{ kops_feature_table(kops_added_default='1.18') }}

You can set the retention duration for the hourly and daily backups by defining env vars:

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

{{ kops_feature_table(kops_added_default='1.23') }}

In AWS, instead of listing all CIDRs, it is possible to specify a pre-existing [AWS Prefix List](https://docs.aws.amazon.com/vpc/latest/userguide/managed-prefix-lists.html) ID.

## kubernetesApiAccess

This array configures the CIDRs that are able to access the kubernetes API. On AWS this is manifested as inbound security group rules on the ELB or master security groups.

Use this key to restrict cluster access to an office ip address range, for example.

```yaml
spec:
  kubernetesApiAccess:
    - 12.34.56.78/32
```

{{ kops_feature_table(kops_added_default='1.23') }}

In AWS, instead of listing all CIDRs, it is possible to specify a pre-existing [AWS Prefix List](https://docs.aws.amazon.com/vpc/latest/userguide/managed-prefix-lists.html) ID.

## cluster.spec Subnet Keys

### id
ID of a subnet to share in an existing VPC.

### egress
The resource identifier (ID) of something in your existing VPC that you would like to use as "egress" to the outside world.

This feature was originally envisioned to allow re-use of NAT gateways. In this case, the usage is as follows. Although NAT gateways are "public"-facing resources, in the Cluster spec, you must specify them in the private subnet section. One way to think about this is that you are specifying "egress", which is the default route out from this private subnet.

```yaml
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

In the case that you don't want to use an existing NAT gateway, but still want to use a pre-allocated elastic IP, kOps 1.19.0 introduced the possibility to specify an elastic IP as egress and kOps will create a NAT gateway that uses it.

```yaml
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    egress: eipalloc-0123456789abcdef0
    type: Private
    zone: us-east-1a
```

Specifying an existing AWS Transit gateways is also supported as of kOps 1.20.0:

```yaml
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    egress: tgw-0123456789abcdef0
    type: Private
    zone: us-east-1a
```

In the case that you don't use NAT gateways or internet gateways, kOps 1.12.0 introduced the "External" flag for egress to force kOps to ignore egress for the subnet. This can be useful when other tools are used to manage egress for the subnet such as virtual private gateways. Please note that your cluster may need to have access to the internet upon creation, so egress must be available upon initializing a cluster. This is intended for use when egress is managed external to kOps, typically with an existing cluster.

```yaml
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

```yaml
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    publicIP: 203.93.148.142
    type: Private
    zone: us-east-1a
```

### additionalRoutes

{{ kops_feature_table(kops_added_default='1.24') }}

Add routes in the route tables of the subnet. Targets of routes can be an instance, a peering connection, a NAT gateway, a transit gateway, an internet gateway or an egress-only internet gateway.
Currently, only AWS is supported.

```yaml
spec:
  subnets:
  - cidr: 10.20.64.0/21
    name: us-east-1a
    type: Private
    zone: us-east-1a
    additionalRoutes:
    - cidr: 10.21.0.0/16
      target: vpc-abcdef
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

### Audit Logging

Read more about this here: https://kubernetes.io/docs/tasks/debug-application-cluster/audit/

```yaml
spec:
  kubeAPIServer:
    auditLogMaxAge: 10
    auditLogMaxBackups: 1
    auditLogMaxSize: 100
    auditLogPath: /var/log/kube-apiserver-audit.log
    auditPolicyFile: /srv/kubernetes/kube-apiserver/audit-policy-config.yaml
  fileAssets:
  - name: audit-policy-config
    path: /srv/kubernetes/kube-apiserver/audit-policy-config.yaml
    roles:
    - Master
    content: |
      apiVersion: audit.k8s.io/v1
      kind: Policy
      rules:
      - level: Metadata
```

**Note**: The auditPolicyFile is needed. If the flag is omitted, no events are logged.

You could use the [fileAssets](https://github.com/kubernetes/kops/blob/master/docs/cluster_spec.md#fileassets) feature to push an advanced audit policy file on the master nodes.

Example policy file can be found [here](https://raw.githubusercontent.com/kubernetes/website/master/content/en/examples/audit/audit-policy.yaml)

### Audit Webhook Backend

Webhook backend sends audit events to a remote API, which is assumed to be the same API as `kube-apiserver` exposes.

```yaml
spec:
  kubeAPIServer:
    auditWebhookBatchMaxWait: 5s
    auditWebhookConfigFile: /srv/kubernetes/kube-apiserver/audit-webhook-config.yaml
  fileAssets:
  - name: audit-webhook-config
    path: /srv/kubernetes/kube-apiserver/audit-webhook-config.yaml
    roles:
    - Master
    content: |
      apiVersion: v1
      kind: Config
      clusters:
      - name: server
        cluster:
          server: https://my-webhook-receiver
      contexts:
      - context:
          cluster: server
          user: ""
        name: default-context
      current-context: default-context
      preferences: {}
      users: []
```

**Note**: The audit logging config is also needed. If it is omitted, no events are shipped.

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

### Request Timeout
{{ kops_feature_table(kops_added_default='1.19') }}

The duration a handler must keep a request open before timing it out and can be overridden by other flags for specific types of requests.
Note that you must fill empty units of time with zeros. (default 1m0s)

```yaml
spec:
  kubeAPIServer:
    requestTimeout: 3m0s
```

### Profiling
{{ kops_feature_table(kops_added_default='1.18') }}

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

### Customize client-ca file

This value is passed as `--client-ca-file` for `kube-apiserver`. (default: `/srv/kubernetes/ca.crt`)

```yaml
spec:
  kubeAPIServer:
    clientCAFile: /srv/kubernetes/client-ca.crt
```
There are certain cases that the user may want to use a customized client CA file other than the default one generated for Kubernetes. In that case, the user can use this flag to specify the client-ca file to use.

To prepare the customized client-ca file on master nodes, the user can either use the [fileAssets](https://kops.sigs.k8s.io/cluster_spec/#fileassets) feature to push an client-ca file, or embed the customized client-ca file in the master AMI.

In the case that the user would use a customized client-ca file, it is common that the kubernetes CA (`/srv/kubernetes/ca/crt`) need to be appended to the end of the client-ca file. One way to append the ca.crt to the end of the customized client-ca file is to write an [kop-hook](https://kops.sigs.k8s.io/cluster_spec/#hooks) to do the append logic.

Kops has a [CA rotation](operations/rotate-secrets.md) feature, which refreshes the Kubernetes certificate files, including the ca.crt. If a customized client-ca file is used, when kOps cert rotation happens, the user is responsible for updating the ca.crt in the customized client-ca file. The refresh ca.crt logic can also be achieved by writing a kops hook.

See also [Kubernetes certificates](https://kubernetes.io/docs/concepts/cluster-administration/certificates/)

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

### Taint based Evictions

There are two parameters related to taint based evictions. These parameters indicate default value of the `tolerationSeconds` for `notReady:NoExecute` and `unreachable:NoExecute`.

```yaml
spec:
  kubeAPIServer:
    defaultNotReadyTolerationSeconds: 600
    defaultUnreachableTolerationSeconds: 600
```

### LogFormat

Choose between log format. Permitted formats: "json", "text". Default: "text".

```yaml
spec:
  kubeAPIServer:
    logFormat: json
```

## externalDns

This block contains configuration options for your `external-DNS` provider.

```yaml
spec:
  externalDns:
    watchIngress: true
```

Default kOps behavior is false. `watchIngress: true` uses the default _dns-controller_ behavior which is to watch the ingress controller for changes. Set this option at risk of interrupting Service updates in some cases.

The default external-DNS provider is the kOps `dns-controller`.

You can use [external-dns](https://github.com/kubernetes-sigs/external-dns/) as provider instead by adding the following:

```yaml
spec:
  externalDns:
    provider: external-dns
```

Note that you if you have dns-controller installed, you need to remove this deployment before updating the cluster with the new configuration.

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

```yaml
spec:
  kubelet:
    cpuCFSQuota: false
```

### Configure CPU CFS Period
Configure CPU CFS quota period value (cpu.cfs_period_us). Example:

```yaml
spec:
  kubelet:
    cpuCFSQuotaPeriod: "100ms"
```

This change requires `CustomCPUCFSQuotaPeriod` [feature gate](#feature-gates).

### Enable Custom metrics support
To use custom metrics in kubernetes as per [custom metrics doc](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-custom-metrics)
we have to set the flag `--enable-custom-metrics` to `true` on all the kubelets. We can specify that in the `kubelet` spec in our cluster.yml.

```yaml
spec:
  kubelet:
    enableCustomMetrics: true
```

### Setting kubelet CPU management policies
kOps 1.12.0 added support for enabling cpu management policies in kubernetes as per [cpu management doc](https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#cpu-management-policies)
we have to set the flag `--cpu-manager-policy` to the appropriate value on all the kubelets. This must be specified in the `kubelet` spec in our cluster.yml.

```yaml
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

kOps will set this for you based off the Operating System in use:
- ContainerOS: `/home/kubernetes/flexvolume/`
- Flatcar: `/var/lib/kubelet/volumeplugins/`
- Default (in-line with upstream k8s): `/usr/libexec/kubernetes/kubelet-plugins/volume/exec/`

If you wish to override this value, it can be done so with the following addition to the kubelet spec:
```yaml
spec:
  kubelet:
    volumePluginDirectory: /provide/a/writable/path/here
```

### Protect Kernel Defaults
{{ kops_feature_table(kops_added_default='1.18', k8s_min='1.4') }}

Default kubelet behaviour for kernel tuning. If set, kubelet errors if any of kernel tunables is different than kubelet defaults.

```yaml
spec:
  kubelet:
    protectKernelDefaults: true
```

### Housekeeping Interval
{{ kops_feature_table(kops_added_default='1.19', k8s_min='1.2') }}

The interval between container housekeepings defaults to `10s`. This can be too small or too high for some use cases and can be modified with the following addition to the kubelet spec.

```yaml
spec:
  kubelet:
    housekeepingInterval: 30s
```

### Pod PIDs Limit
{{ kops_feature_table(kops_added_default='1.22', k8s_min='1.20') }}

`podPidsLimit` allows to configure the maximum number of pids (process ids) in any pod.
[Read more](https://kubernetes.io/docs/concepts/policy/pid-limiting/) in Kubernetes documentation.

```yaml
spec:
  kubelet:
    podPidsLimit: 1024
```

### Event QPS
{{ kops_feature_table(kops_added_default='1.19') }}

The limit event creations per second in kubelet. Default value is `0` which means unlimited event creations.

```yaml
spec:
  kubelet:
    eventQPS: 0
```

### Event Burst
{{ kops_feature_table(kops_added_default='1.19') }}

Maximum size of a bursty event records, temporarily allows event records to burst to this number, while still not exceeding EventQPS. Only used if EventQPS > 0.

```yaml
spec:
  kubelet:
    eventBurst: 10
```

### LogFormat

Choose between log format. Permitted formats: "json", "text". Default: "text".

```yaml
spec:
  kubelet:
    logFormat: json
```

### Graceful Node Shutdown

{{ kops_feature_table(kops_added_default='1.23', k8s_min='1.21') }}

Graceful node shutdown allows kubelet to prevent instance shutdown until Pods have been safely terminated or a timeout has been reached.

For all CNIs except `amazonaws`, kOps will try to add a 30 second timeout for 30 seconds where the first 20 seconds is reserved for normal Pods and the last 10 seconds for critical Pods. When using `amazonaws` this feature is disabled, as it leads to [leaking ENIs](https://github.com/aws/amazon-vpc-cni-k8s/issues/1223).

This configuration can be changed as follows:

```yaml
spec:
  kubelet:
    shutdownGracePeriod: 60s
    shutdownGracePeriodCriticalPods: 20s
```

Note that Kubelet will fail to install the shutdown inhibtor on systems where logind is configured with an `InhibitDelayMaxSeconds` lower than `shutdownGracePeriod`. On Ubuntu, this setting is 30 seconds.

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

### LogFormat

Choose between log format. Permitted formats: "json", "text". Default: "text".

```yaml
spec:
  kubeScheduler:
    logFormat: json
```

## kubeDNS

This block contains configurations for [CoreDNS](https://coredns.io/).

For Kubernetes version >= 1.20, `CoreDNS` will be installed as the default DNS server.

 ```yaml
 spec:
   kubeDNS:
     provider: CoreDNS
```
OR
```yaml
spec:
   kubeDNS:
```

Specifying KubeDNS will install kube-dns as the default service discovery instead of [CoreDNS](https://coredns.io/).

 ```yaml
 spec:
   kubeDNS:
     provider: KubeDNS
```

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

For larger clusters you may need to set custom resource requests and limits. For the CoreDNS provider you can set

- memoryLimit
- cpuRequest
- memoryRequest

This will override the default limit value for memory of 170Mi and default request values for memory and cpu of 70Mi and 100m.

Example:
```
  kubeDNS:
    memoryLimit: 2Gi
    cpuRequest: 300m
    memoryRequest: 700Mi
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
    horizontalPodAutoscalerInitialReadinessDelay: 30s
    horizontalPodAutoscalerCpuInitializationPeriod: 5m
    horizontalPodAutoscalerTolerance: 0.1
    experimentalClusterSigningDuration: 8760h0m0s
    enableProfiling: false
```

For more details on `horizontalPodAutoscaler` flags see the [official HPA docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) and the [kOps guides on how to set it up](horizontal_pod_autoscaling.md).

### LogFormat

Choose between log format. Permitted formats: "json", "text". Default: "text".

```yaml
spec:
  kubeControllerManager:
    logFormat: json
```

##  Feature Gates

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

##  Compute Resources Reservation

In a scenario where node has 32Gi of memory, 16 CPUs and 100Gi of ephemeral storage, resource reservation could be set as in the following example:

```yaml
spec:
  kubelet:
    kubeReserved:
        cpu: "1"
        memory: "2Gi"
        ephemeral-storage: "1Gi"
    kubeReservedCgroup: "/kube-reserved"
    kubeletCgroups: "/kube-reserved"
    runtimeCgroups: "/kube-reserved"
    systemReserved:
        cpu: "500m"
        memory: "1Gi"
        ephemeral-storage: "1Gi"
    systemReservedCgroup: "/system-reserved"
    enforceNodeAllocatable: "pods,system-reserved,kube-reserved"
```

The above will result in the flags `--kube-reserved=cpu=1,memory=2Gi,ephemeral-storage=1Gi --kube-reserved-cgroup=/kube-reserved --kubelet-cgroups=/kube-reserved --runtime-cgroups=/kube-reserved --system-reserved=cpu=500m,memory=1Gi,ephemeral-storage=1Gi --system-reserved-cgroup=/system-reserved --enforce-node-allocatable=pods,system-reserved,kube-reserved` being added to the kubelet.

Learn more about reserving compute resources [here](https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/) and [here](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/).

## networkID

On AWS, this is the id of the VPC the cluster is created in. If creating a cluster from scratch, this field does not need to be specified at create time; `kops` will create a `VPC` for you.

```yaml
spec:
  networkID: vpc-abcdefg1
```

More information about running in an existing VPC is [here](run_in_existing_vpc.md).

## hooks

Hooks allow for the execution of an action before the installation of Kubernetes on every node in a cluster. For instance you can install Nvidia drivers for using GPUs. This hooks can be in the form of container images or manifest files (systemd units). Hooks can be placed in either the cluster spec, meaning they will be globally deployed, or they can be placed into the instanceGroup specification. Note: service names on the instanceGroup which overlap with the cluster spec take precedence and ignore the cluster spec definition, i.e. if you have a unit file 'myunit.service' in cluster and then one in the instanceGroup, only the instanceGroup is applied.

When creating a systemd unit hook using the `manifest` field, the hook system will construct a systemd unit file for you. It creates the `[Unit]` section, adding an automated description and setting `Before` and `Requires` values based on the `before` and `requires` fields. The value of the `manifest` field is used as the `[Service]` section of the unit file. To override this behavior, and instead specify the entire unit file yourself, you may specify `useRawManifest: true`. In this case, the contents of the `manifest` field will be used as a systemd unit, unmodified. The `before` and `requires` fields may not be used together with `useRawManifest`.

```yaml
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

```yaml
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

```yaml
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

FileAssets permit you to place inline file content into the Cluster and [Instance Group](instance_groups.md) specifications. This is useful for deploying additional files that Kubernetes components require, such as audit logging or admission controller configurations.

```yaml
spec:
  fileAssets:
  - name: iptable-restore
    # Note if path is not specified, the default is /srv/kubernetes/assets/<name>
    path: /var/lib/iptables/rules-save
    # Note if roles are not specified, the default is all roles
    roles: [Master,Node,Bastion] # a list of roles to apply the asset to
    content: |
      some file content
```

### mode

{{ kops_feature_table(kops_added_default='1.24') }}

Optionally, `mode` allows you to specify a file's mode and permission bits.

**NOTE**: If not specified, the default is `"0440"`, which matches the behaviour of older versions of kOps.

```yaml
spec:
  fileAssets:
  - name: my-script
    path: /usr/local/bin/my-script
    mode: "0550"
    content: |
      #! /usr/bin/env bash
      ...
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

### manageStorageClasses
{{ kops_feature_table(kops_added_default='1.20') }}


By default kops will create `StorageClass` resources with some opinionated settings specific to cloud provider on which the cluster is installed. One of those storage classes will be defined as default applying the annotation `storageclass.kubernetes.io/is-default-class: "true"`. This may not always be a desirable behaviour and some cluster admins rather prefer to have more control of storage classes and manage them outside of kops. When set to `false`, kOps will no longer create any `StorageClass` objects. Any such objects that kOps created in the past are left as is, and kOps will no longer reconcile them against future changes.

The existing `spec.cloudConfig.openstack.blockStorage.createStorageClass` field remains in place. However, if both that and the new `spec.cloudConfig.manageStorageClasses` field are populated, they must agree: It is invalid both to disable management of `StorageClass` objects globally but to enable them for OpenStack and, conversely, to enable management globally but disable it for OpenStack.

```yaml
spec:
  cloudConfig:
    manageStorageClasses: false
```

## containerRuntime
{{ kops_feature_table(kops_added_default='1.18', k8s_min='1.11') }}

As of Kubernetes 1.20, the default [container runtime](https://kubernetes.io/docs/setup/production-environment/container-runtimes) is containerd. Previously, the default container runtime was Docker.

Docker can still be used as container runtime with Kubernetes 1.20+,  but be aware that Kubernetes is [deprecating](https://kubernetes.io/blog/2020/12/02/dont-panic-kubernetes-and-docker) support for it and will be removed in Kubernetes 1.22.

```yaml
spec:
  containerRuntime: containerd
```

## containerd

### Configuration

It is possible to override the [containerd](https://github.com/containerd/containerd/blob/master/README.md) daemon options for all the nodes in the cluster. See the [API docs](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#ContainerdConfig) for the full list of options.
Overriding the configuration of containerd has to be done with care as the default config may change with new releases and can lead to incompatibilities.

```yaml
spec:
  containerd:
    version: 1.4.4
    logLevel: info
    configOverride: ""
```

### Custom Packages

kOps uses the `.tar.gz` packages for installing containerd on any supported OS. This makes it easy to use a custom build or pre-release packages, by specifying its URL and sha256:

```yaml
spec:
  containerd:
    packages:
      urlAmd64: https://github.com/containerd/containerd/releases/download/v1.4.4/cri-containerd-cni-1.4.4-linux-amd64.tar.gz
      hashAmd64: 96641849cb78a0a119223a427dfdc1ade88412ef791a14193212c8c8e29d447b
```

The format of the custom package must be identical to the official packages:

```bash
tar tf cri-containerd-cni-1.4.4-linux-amd64.tar.gz
    usr/local/bin/containerd
    usr/local/bin/containerd-shim
    usr/local/bin/containerd-shim-runc-v1
    usr/local/bin/containerd-shim-runc-v2
    usr/local/bin/crictl
    usr/local/bin/critest
    usr/local/bin/ctr
    usr/local/sbin/runc
```

### Runc Version and Packages
{{ kops_feature_table(kops_added_default='1.24.2') }}

kOps uses the binaries from https://github.com/opencontainers/runc for installing runc on any supported OS. This makes it easy to specify the desired release version:

```yaml
spec:
  containerd:
    runc:
      version: 1.1.2
```

It also makes it possible to use a newer version than the kOps binary, pre-release packages, or even a custom build, by specifying its URL and sha256:

```yaml
spec:
  containerd:
    runc:
      version: 1.100.0
      packages:
        urlAmd64: https://cdn.example.com/k8s/runc/releases/download/v1.100.0/runc.amd64
        hashAmd64: ab1c67fbcbdddbe481e48a55cf0ef9a86b38b166b5079e0010737fd87d7454bb
```

### Registry Mirrors
{{ kops_feature_table(kops_added_default='1.19') }}

If you have many instances running, each time one of them pulls an image that is not present on the host, it will fetch it from the internet. By caching these images, you can keep the traffic within your local network and avoid egress bandwidth usage.

See [Image Registry](https://github.com/containerd/containerd/blob/master/docs/cri/registry.md#configure-registry-endpoint) docs for more info.

```yaml
spec:
  containerd:
    registryMirrors:
      docker.io:
      - https://registry-1.docker.io
      "*":
      - http://HostIP2:Port2
```

## Docker

It is possible to override Docker daemon options for all masters and nodes in the cluster. See the [API docs](https://pkg.go.dev/k8s.io/kops/pkg/apis/kops#DockerConfig) for the full list of options.

### Registry Mirrors

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

### Storage

The Docker [Storage Driver](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-storage-driver) can be specified in order to override the default. Be sure the driver you choose is supported by your operating system and docker version.

```yaml
docker:
  storage: devicemapper
  storageOpts:
    - "dm.thinpooldev=/dev/mapper/thin-pool"
    - "dm.use_deferred_deletion=true"
    - "dm.use_deferred_removal=true"
```

### Networking

In order for containers started with `docker run` instead of Kubernetes to have network and internet access you need to enable the necessary [iptables](https://docs.docker.com/network/iptables/) rules:

```yaml
docker:
  ipMasq: true
  ipTables: true
```

### Custom Packages

kOps uses the `.tgz` (static) packages for installing Docker on any supported OS. This makes it easy to use a custom build or pre-release packages, by specifying its URL and sha256:

```yaml
spec:
  containerd:
    packages:
      urlAmd64: https://download.docker.com/linux/static/stable/x86_64/docker-20.10.1.tgz
      hashAmd64: 8790f3b94ee07ca69a9fdbd1310cbffc729af0a07e5bf9f34a79df1e13d2e50e
```

The format of the custom package must be identical to the official packages:

```bash
tar tf docker-20.10.1.tgz
    docker/containerd
    docker/containerd-shim
    docker/containerd-shim-runc-v2
    docker/ctr
    docker/docker
    docker/docker-init
    docker/docker-proxy
    docker/dockerd
    docker/runc
```

## sshKeyName

In some cases, it may be desirable to use an existing AWS SSH key instead of allowing kOps to create a new one.
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

In some use-cases you may wish to augment the target output with extra options.  `target` supports a minimal amount of options you can do this with.  Currently only the terraform target supports this, but if other use cases present themselves, kOps may eventually support more.

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

The container registry enables kOps / kubernetes to pull containers from a managed registry.
This is useful when pulling containers from the internet is not an option, eg. because the
deployment is offline / internet restricted or because of special requirements that apply
for deployed artifacts, eg. auditing of containers.

For a use case example, see [How to use kOps in AWS China Region](https://github.com/kubernetes/kops/blob/master/docs/aws-china.md)

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
{{ kops_feature_table(kops_added_default='1.17') }}

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

## cgroupDriver

As of Kubernetes 1.20, kOps will default the cgroup driver of the kubelet and the container runtime to use systemd as the default cgroup driver
as opposed to cgroup fs.

It is important to ensure that the kubelet and the container runtime are using the same cgroup driver. Below are examples showing
how to set the cgroup driver for kubelet and the container runtime.


Setting kubelet to use cgroupfs
```yaml
spec:
  kubelet:
    cgroupDriver: cgroupfs
```

Setting Docker to use cgroupfs
```yaml
spec:
  docker:
    execOpt:
      - native.cgroupdriver=cgroupfs
```

In the case of containerd, the cgroup-driver is dependent on the cgroup driver of kubelet. To use cgroupfs, just update the
cgroupDriver of kubelet to use cgroupfs.

## NTP

The installation and the configuration of NTP can be skipped by setting `managed` to `false`.

```yaml
spec:
  ntp:
    managed: false
```

## Service Account Issuer Discovery and AWS IAM Roles for Service Accounts (IRSA)

{{ kops_feature_table(kops_added_default='1.21') }}

**Warning**: Enabling the following configuration on an existing cluster can be disruptive due to the control plane provisioning tokens with different issuers. The symptom is that Pods are unable to authenticate to the Kubernetes API. To resolve this, delete Service Account token secrets that exists in the cluster and kill all pods unable to authenticate.

kOps can publish the Kubernetes service account token issuer and configure AWS to trust it
to authenticate Kubernetes service accounts:

```yaml
spec:
  serviceAccountIssuerDiscovery:
    discoveryStore: s3://publicly-readable-store
    enableAWSOIDCProvider: true
```

The `discoveryStore` option causes kOps to publish an OIDC-compatible discovery document
to a path in an S3 bucket. This would ordinarily be a different bucket than the state store.
kOps will automatically configure `spec.kubeAPIServer.serviceAccountIssuer` and default
`spec.kubeAPIServer.serviceAccountJWKSURI` to the corresponding
HTTPS URL.

The `enableAWSOIDCProvider` configures AWS to trust the service account issuer to
authenticate service accounts for IAM Roles for Service Accounts (IRSA). In order for this to work,
the service account issuer discovery URL must be publicly readable.

kOps can provision AWS permissions for use by service accounts:

```yaml
spec:
  iam:
    serviceAccountExternalPermissions:
      - name: someServiceAccount
        namespace: someNamespace
        aws:
          policyARNs:
            - arn:aws:iam::000000000000:policy/somePolicy
      - name: anotherServiceAccount
        namespace: anotherNamespace
        aws:
          inlinePolicy: |-
            [
              {
                "Effect": "Allow",
                "Action": "s3:ListAllMyBuckets",
                "Resource": "*"
              }
            ]
```

To configure Pods to assume the given IAM roles, enable the [Pod Identity Webhook](/addons/#pod-identity-webhook). Without this webhook, you need to modify your Pod specs yourself for your Pod to assume the defined roles.
