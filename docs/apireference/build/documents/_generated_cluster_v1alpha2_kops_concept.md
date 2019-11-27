

-----------
# Cluster v1alpha2 kops

>bdocs-tab:example 

```bdocs-tab:example_yaml

apiVersion: kops.k8s.io/v1alpha2
kind: Cluster
metadata:
  creationTimestamp: "2019-12-10T22:42:27Z"
  name: test.example.com
spec:
  api:
    loadBalancer:
      type: Public
      additionalSecurityGroups:
      - sg-exampleid3
      - sg-exampleid4
  authorization:
    rbac: {}
  kubernetesApiAccess:
  - 0.0.0.0/0
  channel: stable
  cloudProvider: aws
  cloudLabels:
    Owner: John Doe
    foo/bar: fib+baz
  configBase: s3://test.example.com/test.example.com
  etcdClusters:
  - etcdMembers:
    - instanceGroup: master-us-test-1a
      name: us-test-1a
    name: main
    version: 3.0.17
  - etcdMembers:
    - instanceGroup: master-us-test-1a
      name: us-test-1a
    name: events
    version: 3.0.17
  kubeAPIServer:
    serviceNodePortRange: 28000-32767
  kubernetesVersion: 1.8.5
  masterInternalName: api.internal.test.example.com
  masterPublicName: api.test.example.com
  networkCIDR: 172.20.0.0/16
  networkID: vpc-12345678
  networking:
    weave: 
      mtu: 8912
  nodePortAccess:
  - 1.2.3.4/32
  - 10.20.30.0/24
  nonMasqueradeCIDR: 100.64.0.0/10
  sshAccess:
  - 0.0.0.0/0
  subnets:
  - cidr: 172.20.32.0/19
    id: subnet-12345678
    name: us-test-1a
    type: Public
    zone: us-test-1a
  - cidr: 172.20.4.0/22
    name: utility-us-test-1a
    type: Utility
    zone: us-test-1a
    id: subnet-12345679
  topology:
    masters: private
    nodes: private


```


Group        | Version     | Kind
------------ | ---------- | -----------
kops | v1alpha2 | Cluster









<aside class="notice">
Appears In:

<ul> 
<li><a href="#clusterlist-v1alpha2-kops">ClusterList kops/v1alpha2</a></li>
</ul> </aside>

Field        | Description
------------ | -----------
apiVersion <br /> *string*    | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
kind <br /> *string*    | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
metadata <br /> *[ObjectMeta](#objectmeta-v1-meta)*    | 
spec <br /> *[ClusterSpec](#clusterspec-v1alpha2-kops)*    | 


### ClusterSpec v1alpha2 kops

<aside class="notice">
Appears In:

<ul>
<li><a href="#cluster-v1alpha2-kops">Cluster kops/v1alpha2</a></li>
</ul></aside>

Field        | Description
------------ | -----------
additionalNetworkCIDRs <br /> *string array*    | AdditionalNetworkCIDRs is a list of additional CIDR used for the AWS VPC or otherwise allocated to k8s. This is a real CIDR, not the internal k8s network On AWS, it maps to any additional CIDRs added to a VPC.
additionalPolicies <br /> *object*    | Additional policies to add for roles
additionalSans <br /> *string array*    | AdditionalSANs adds additional Subject Alternate Names to apiserver cert that kops generates
addons <br /> *[AddonSpec](#addonspec-v1alpha2-kops) array*    | Additional addons that should be installed on the cluster
api <br /> *[AccessSpec](#accessspec-v1alpha2-kops)*    | API field controls how the API is exposed outside the cluster
assets <br /> *[Assets](#assets-v1alpha2-kops)*    | Alternative locations for files and containers
authentication <br /> *[AuthenticationSpec](#authenticationspec-v1alpha2-kops)*    | Authentication field controls how the cluster is configured for authentication
authorization <br /> *[AuthorizationSpec](#authorizationspec-v1alpha2-kops)*    | Authorization field controls how the cluster is configured for authorization
channel <br /> *string*    | The Channel we are following
cloudConfig <br /> *[CloudConfiguration](#cloudconfiguration-v1alpha2-kops)*    | 
cloudControllerManager <br /> *[CloudControllerManagerConfig](#cloudcontrollermanagerconfig-v1alpha2-kops)*    | 
cloudLabels <br /> *object*    | Tags for AWS resources
cloudProvider <br /> *string*    | The CloudProvider to use (aws or gce)
clusterDNSDomain <br /> *string*    | ClusterDNSDomain is the suffix we use for internal DNS names (normally cluster.local)
configBase <br /> *string*    | ConfigBase is the path where we store configuration for the cluster This might be different that the location when the cluster spec itself is stored, both because this must be accessible to the cluster, and because it might be on a different cloud or storage system (etcd vs S3)
configStore <br /> *string*    | ConfigStore is the VFS path to where the configuration (Cluster, InstanceGroups etc) is stored
dnsZone <br /> *string*    | DNSZone is the DNS zone we should use when configuring DNS This is because some clouds let us define a managed zone foo.bar, and then have kubernetes.dev.foo.bar, without needing to define dev.foo.bar as a hosted zone. DNSZone will probably be a suffix of the MasterPublicName and MasterInternalName Note that DNSZone can either by the host name of the zone (containing dots), or can be an identifier for the zone.
docker <br /> *[DockerConfig](#dockerconfig-v1alpha2-kops)*    | Component configurations
egressProxy <br /> *[EgressProxySpec](#egressproxyspec-v1alpha2-kops)*    | HTTPProxy defines connection information to support use of a private cluster behind an forward HTTP Proxy
encryptionConfig <br /> *boolean*    | EncryptionConfig holds the encryption config
etcdClusters <br /> *[EtcdClusterSpec](#etcdclusterspec-v1alpha2-kops) array*    | EtcdClusters stores the configuration for each cluster
externalDns <br /> *[ExternalDNSConfig](#externaldnsconfig-v1alpha2-kops)*    | 
fileAssets <br /> *[FileAssetSpec](#fileassetspec-v1alpha2-kops) array*    | A collection of files assets for deployed cluster wide
hooks <br /> *[HookSpec](#hookspec-v1alpha2-kops) array*    | Hooks for custom actions e.g. on first installation
iam <br /> *[IAMSpec](#iamspec-v1alpha2-kops)*    | IAM field adds control over the IAM security policies applied to resources
isolateMasters <br /> *boolean*    | IsolateMasters determines whether we should lock down masters so that they are not on the pod network. true is the kube-up behaviour, but it is very surprising: it means that daemonsets only work on the master if they have hostNetwork=true. false is now the default, and it will:  * give the master a normal PodCIDR  * run kube-proxy on the master  * enable debugging handlers on the master, so kubectl logs works
keyStore <br /> *string*    | KeyStore is the VFS path to where SSL keys and certificates are stored
kubeAPIServer <br /> *[KubeAPIServerConfig](#kubeapiserverconfig-v1alpha2-kops)*    | 
kubeControllerManager <br /> *[KubeControllerManagerConfig](#kubecontrollermanagerconfig-v1alpha2-kops)*    | 
kubeDNS <br /> *[KubeDNSConfig](#kubednsconfig-v1alpha2-kops)*    | 
kubeProxy <br /> *[KubeProxyConfig](#kubeproxyconfig-v1alpha2-kops)*    | 
kubeScheduler <br /> *[KubeSchedulerConfig](#kubeschedulerconfig-v1alpha2-kops)*    | 
kubelet <br /> *[KubeletConfigSpec](#kubeletconfigspec-v1alpha2-kops)*    | 
kubernetesApiAccess <br /> *string array*    | KubernetesAPIAccess determines the permitted access to the API endpoints (master HTTPS) Currently only a single CIDR is supported (though a richer grammar could be added in future)
kubernetesVersion <br /> *string*    | The version of kubernetes to install (optional, and can be a "spec" like stable)
masterInternalName <br /> *string*    | MasterInternalName is the internal DNS name for the master nodes
masterKubelet <br /> *[KubeletConfigSpec](#kubeletconfigspec-v1alpha2-kops)*    | 
masterPublicName <br /> *string*    | MasterPublicName is the external DNS name for the master nodes
networkCIDR <br /> *string*    | NetworkCIDR is the CIDR used for the AWS VPC / GCE Network, or otherwise allocated to k8s This is a real CIDR, not the internal k8s network On AWS, it maps to the VPC CIDR.  It is not required on GCE.
networkID <br /> *string*    | NetworkID is an identifier of a network, if we want to reuse/share an existing network (e.g. an AWS VPC)
networking <br /> *[NetworkingSpec](#networkingspec-v1alpha2-kops)*    | Networking configuration
nodePortAccess <br /> *string array*    | NodePortAccess is a list of the CIDRs that can access the node ports range (30000-32767).
nonMasqueradeCIDR <br /> *string*    | MasterIPRange                 string `json:",omitempty"` NonMasqueradeCIDR is the CIDR for the internal k8s network (on which pods & services live) It cannot overlap ServiceClusterIPRange
project <br /> *string*    | Project is the cloud project we should use, required on GCE
secretStore <br /> *string*    | SecretStore is the VFS path to where secrets are stored
serviceClusterIPRange <br /> *string*    | ServiceClusterIPRange is the CIDR, from the internal network, where we allocate IPs for services
sshAccess <br /> *string array*    | SSHAccess determines the permitted access to SSH Currently only a single CIDR is supported (though a richer grammar could be added in future)
sshKeyName <br /> *string*    | SSHKeyName specifies a preexisting SSH key to use
subnets <br /> *[ClusterSubnetSpec](#clustersubnetspec-v1alpha2-kops) array*    | Configuration of subnets we are targeting
target <br /> *[TargetSpec](#targetspec-v1alpha2-kops)*    | Target allows for us to nest extra config for targets such as terraform
topology <br /> *[TopologySpec](#topologyspec-v1alpha2-kops)*    | Topology defines the type of network topology to use on the cluster - default public This is heavily weighted towards AWS for the time being, but should also be agnostic enough to port out to GCE later if needed
updatePolicy <br /> *string*    | UpdatePolicy determines the policy for applying upgrades automatically. Valid values:   'external' do not apply updates automatically - they are applied manually or by an external system   missing: default policy (currently OS security upgrades that do not require a reboot)

### ClusterList v1alpha2 kops



Field        | Description
------------ | -----------
apiVersion <br /> *string*    | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
items <br /> *[Cluster](#cluster-v1alpha2-kops) array*    | 
kind <br /> *string*    | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
metadata <br /> *[ListMeta](#listmeta-v1-meta)*    | 





