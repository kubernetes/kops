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

> __Note:__ The images for etcd that kops uses are from the Google Cloud Repository. Google doesn't release every version of etcd to the gcr. Check that the version of etcd you want to use is available [at the gcr](https://console.cloud.google.com/gcr/images/google-containers/GLOBAL/etcd?gcrImageListsize=50) before using it in your cluster spec.

By default, the Volumes created for the etcd clusters are `gp2` and 20GB each. The volume size, type and Iops( for `io1`) can be configured via their parameters. Conversion between `gp2` and `io1` is not supported, nor are size changes.

As of Kops 1.12.0 it is also possible to specify the requests for your etcd cluster members using the `cpuRequest` and `memoryRequest` parameters.

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

#### publicIP
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
    oidcRequiredClaim:
    	- "key=value"
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

Example policy file can be found [here](https://raw.githubusercontent.com/kubernetes/website/master/content/en/examples/audit/audit-policy.yaml)

#### bootstrap tokens

Read more about this here: https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens/

```yaml
spec:
  kubeAPIServer:
    enableBootstrapTokenAuth: true
```

By enabling this feature you instructing two things;
- master nodes will bypass the bootstrap token but they _will_ build kubeconfigs with unique usernames in the system:nodes group _(this ensure's the master nodes confirm with the node authorization mode https://kubernetes.io/docs/reference/access-authn-authz/node/)_
- secondly the nodes will be configured to use a bootstrap token located by default at `/var/lib/kubelet/bootstrap-kubeconfig` _(though this can be override in the kubelet spec)_. The nodes will sit the until a bootstrap file is created and once available attempt to provision the node.

**Note** enabling bootstrap tokens does not provision bootstrap tokens for the worker nodes. Under this configuration it is assumed a third-party process is provisioning the tokens on behalf of the worker nodes. For the full setup please read [Node Authorizer Service](https://github.com/kubernetes/kops/blob/master/docs/node_authorization.md)

#### Max Requests Inflight

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

#### Disable Basic Auth

This will disable the passing of the `--basic-auth-file` flag.

```yaml
spec:
  kubeAPIServer:
    disableBasicAuth: true
```

#### targetRamMb

Memory limit for apiserver in MB (used to configure sizes of caches, etc.)

```yaml
spec:
  kubeAPIServer:
    targetRamMb: 4096
```

#### eventTTL

How long API server retains events. Note that you must fill empty units of time with zeros.

```yaml
spec:
  kubeAPIServer:
    eventTTL: 03h0m0s
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

#### Disable CPU CFS Quota
To disable CPU CFS quota enforcement for containers that specify CPU limits (default true) we have to set the flag `--cpu-cfs-quota` to `false`
on all the kubelets. We can specify that in the `kubelet` spec in our cluster.yml.

```
spec:
  kubelet:
    cpuCFSQuota: false
```

#### Configure CPU CFS Period
Configure CPU CFS quota period value (cpu.cfs_period_us). Example:

```
spec:
  kubelet:
    cpuCFSQuotaPeriod: "100ms"
```

#### Enable Custom metrics support
To use custom metrics in kubernetes as per [custom metrics doc](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-custom-metrics)
we have to set the flag `--enable-custom-metrics` to `true` on all the kubelets. We can specify that in the `kubelet` spec in our cluster.yml.

```
spec:
  kubelet:
    enableCustomMetrics: true
```

#### Setting kubelet CPU management policies
Kops 1.12.0 added support for enabling cpu management policies in kubernetes as per [cpu management doc](https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#cpu-management-policies)
we have to set the flag `--cpu-manager-policy` to the appropriate value on all the kubelets. This must be specified in the `kubelet` spec in our cluster.yml.

```
spec:
  kubelet:
    cpuManagerPolicy: static
```

#### Setting kubelet configurations together with the Amazon VPC backend
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

#### Configure a Flex Volume plugin directory
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

### kubeScheduler

This block contains configurations for `kube-scheduler`.  See https://kubernetes.io/docs/admin/kube-scheduler/

 ```yaml
 spec:
   kubeScheduler:
     usePolicyConfigMap: true
```

Will make kube-scheduler use the scheduler policy from configmap "scheduler-policy" in namespace kube-system.

Note that as of Kubernetes 1.8.0 kube-scheduler does not reload its configuration from configmap automatically. You will need to ssh into the master instance and restart the Docker container manually.

### kubeDNS

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

**Note:** If you are using this functionality you will need to be extra vigiliant on version changes of CoreDNS for changes in functionality of the plugins being used etc.

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

### kubeControllerManager
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
```

For more details on `horizontalPodAutoscaler` flags see the [official HPA docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) and the [Kops guides on how to set it up](horizontal_pod_autoscaling.md).

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

Some feature gates also require the `featureGates` setting to be used on other components - e.g. `PodShareProcessNamespace` requires
the feature gate to be enabled on the api server:

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

####  Compute Resources Reservation

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

### networkID

On AWS, this is the id of the VPC the cluster is created in. If creating a cluster from scratch, this field does not need to be specified at create time; `kops` will create a `VPC` for you.

```yaml
spec:
  networkID: vpc-abcdefg1
```

More information about running in an existing VPC is [here](run_in_existing_vpc.md).

### hooks

Hooks allow for the execution of an action before the installation of Kubernetes on every node in a cluster.  For instance you can install Nvidia drivers for using GPUs. This hooks can be in the form of Docker images or manifest files (systemd units). Hooks can be placed in either the cluster spec, meaning they will be globally deployed, or they can be placed into the instanceGroup specification. Note: service names on the instanceGroup which overlap with the cluster spec take precedence and ignore the cluster spec definition, i.e. if you have a unit file 'myunit.service' in cluster and then one in the instanceGroup, only the instanceGroup is applied.

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

### fileAssets

FileAssets is an alpha feature which permits you to place inline file content into the cluster and instanceGroup specification. It's designated as alpha as you can probably do this via kubernetes daemonsets as an alternative.

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


### cloudConfig

#### disableSecurityGroupIngress
If you are using aws as `cloudProvider`, you can disable authorization of ELB security group to Kubernetes Nodes security group. In other words, it will not add security group rule.
This can be useful to avoid AWS limit: 50 rules per security group.
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

#### Skip Install

If you want nodeup to skip the Docker installation tasks, you can do so with:

```yaml
spec:
  docker:
    skipInstall: true
``` 

**NOTE:** When this field is set to `true`, it is entirely up to the user to install and configure Docker.

#### Static Binary Archive

Kops also has the option to install from a static binary archive. This can be useful if you cannot download the Docker
binaries from download.docker.com, and have a heterogeneous cluster.

```yaml
spec:
  docker:
    dockerVersion: "18.06.3"
    staticBinaryUrl: https://dockermirror.example.com/linux/static/stable/x86_64/docker-18.06.3-ce.tgz
    staticBinaryHash: 79a983a41275bd7a660a42935504f5c1ff7a5858
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

### useHostCertificates

Self-signed certificates towards Cloud APIs. In some cases Cloud APIs do have self-signed certificates.

```yaml
spec:
  useHostCertificates: true
```

#### Optional step: add root certificates to instancegroups root ca bundle

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


### target

In some use-cases you may wish to augment the target output with extra options.  `target` supports a minimal amount of options you can do this with.  Currently only the terraform target supports this, but if other use cases present themselves, kops may eventually support more.

```yaml
spec:
  target:
    terraform:
      providerExtraConfig:
        alias: foo
```

### assets

Assets define alernative locations from where to retrieve static files and containers

#### containerRegistry

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


#### containerProxy

The container proxy is designed to acts as a [pull through cache](https://docs.docker.com/registry/recipes/mirror/) for docker container assets.
Basically, what it does is it remaps the Kubernetes image URL to point to you cache so that the docker daemon will pull the image from that location.
If, for example, the containerProxy is set to `proxy.example.com`, the image `k8s.gcr.io/kube-apiserver` will be pulled from `proxy.example.com/kube-apiserver` instead.
Note that the proxy you use has to support this feature for private registries.


```yaml
spec:
  assets:
    containerProxy: proxy.example.com
```
