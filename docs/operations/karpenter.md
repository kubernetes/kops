# Karpenter

[Karpenter](https://karpenter.sh) is an open-source node lifecycle management project built for Kubernetes.
Adding Karpenter to a Kubernetes cluster can dramatically improve the efficiency and cost of running workloads on that cluster.

On AWS, kOps supports managing an InstanceGroup with either Karpenter or an AWS Auto Scaling Group (ASG).
On GCE, kOps supports managing an InstanceGroup with either Karpenter or a Managed Instance Group (MIG); see [Karpenter on GCE](#karpenter-on-gce).

## Prerequisites

On AWS, managed Karpenter requires kOps 1.34+ and that [IAM Roles for Service Accounts (IRSA)](/cluster_spec#service-account-issuer-discovery-and-aws-iam-roles-for-service-accounts-irsa) be enabled for the cluster.

If an older version of Karpenter was installed, it must be uninstalled before installing the new version.

## Installing

### New clusters (AWS only)

```sh
export KOPS_STATE_STORE="s3://my-state-store"
export KOPS_DISCOVERY_STORE="s3://my-discovery-store" 
export NAME="my-cluster.example.com"
export ZONES="eu-central-1a"

kops create cluster --name ${NAME} \
  --cloud=aws \
  --instance-manager=karpenter \
  --discovery-store=${KOPS_DISCOVERY_STORE} \
  --zones=${ZONES} \
  --yes

kops validate cluster --name ${NAME} --wait=10m

kops export kubeconfig --name ${NAME} --admin
```

### Existing clusters

The Karpenter addon must be enabled in the cluster spec:

```yaml
spec:
  karpenter:
    enabled: true
```

To create a Karpenter InstanceGroup, set the following in its InstanceGroup spec:

```yaml
spec:
  role: Node
  manager: Karpenter
```

The manager of an existing InstanceGroup cannot be changed: create a new InstanceGroup with the other manager and delete the old one.

### EC2NodeClass and NodePool
{{ kops_feature_table(kops_added_default='1.36') }}

kOps generates one `EC2NodeClass` and one `NodePool` for each AWS node InstanceGroup with `spec.manager: Karpenter`.
The generated objects use the InstanceGroup name, are delivered by the `karpenter.sh` addon, and are pruned when the InstanceGroup is removed.
When `kops update cluster` updates only some InstanceGroups, with `--instance-group` or `--instance-group-roles`, the objects of the other InstanceGroups are kept as the previous update published them, like the rest of their configuration.

The generated `EC2NodeClass` uses:

* `amiFamily: Custom`
* the InstanceGroup image translated into `amiSelectorTerms`
* the InstanceGroup `spec.rootVolume` translated into `blockDeviceMappings`
* the InstanceGroup `spec.instanceMetadata` translated into `metadataOptions`
* the kOps node instance profile
* the kOps node security groups
* the subnets tagged for the InstanceGroup
* the kOps nodeup bootstrap script as `userData`

The generated `NodePool` references that `EC2NodeClass`, sets Linux as a requirement, and includes instance type and capacity type requirements when they are configured on the InstanceGroup.
Safe InstanceGroup node labels and taints are added to the NodePool template.

Supported image selector forms are:

* `ami-*`
* `ssm:<parameter>`
* `<name>`
* `<owner>/<name>`

### Instance requirements
{{ kops_feature_table(kops_added_default='1.37') }}

By default, the generated `NodePool` requires one of the instance types listed in `spec.machineType` and `spec.mixedInstancesPolicy.instances`.
To let Karpenter choose instance types based on a capacity range, set `spec.mixedInstancesPolicy.instanceRequirements`:

```yaml
spec:
  role: Node
  manager: Karpenter
  mixedInstancesPolicy:
    instanceRequirements:
      cpu:
        min: "2"
        max: "16"
      memory:
        min: "2Gi"
        max: "64Gi"
      excludedInstanceTypes:
      - m3.*
      - t3.small
```

This generates the appropriate `instance-cpu` and `instance-memory` requirements on the `NodePool` (`karpenter.k8s.aws/*` on AWS, `karpenter.k8s.gcp/*` on GCE).
`excludedInstanceTypes` entries are mapped to `NotIn` requirements: a `<family>.*` wildcard excludes an instance family (`instance-family`), and a bare instance type excludes that type.

When `instanceRequirements` is set, neither `spec.machineType` nor `spec.mixedInstancesPolicy.instances` is required.
If either field is set, its instance types further restrict the generated `NodePool`.

Karpenter NodePools can include both GPU and non-GPU instance types.
When using kOps-managed NVIDIA support, use a dedicated GPU-only InstanceGroup because kOps applies GPU labels and taints to the entire NodePool.

## Karpenter-managed InstanceGroups
{{ kops_feature_table(kops_added_default='1.36') }}

A Karpenter-managed InstanceGroup controls the bootstrap script. kOps ensures the correct cloud resources (security groups and subnets on AWS, network tags and service accounts on GCE), permissions, and Karpenter resource definitions.

When `minSize` is omitted, kOps generates a dynamic NodePool and Karpenter owns scale-out decisions.
For a static NodePool, set `minSize` to a positive number:

```yaml
spec:
  role: Node
  manager: Karpenter
  minSize: 4
```

For new AWS clusters, `--instance-manager=karpenter --node-count=4` creates the same static configuration.
Zero and negative `minSize` values are rejected.

The Karpenter addon enables `StaticCapacity` by default.
If `cluster.spec.karpenter.featureGates` is customized, it must include `StaticCapacity=true` for static InstanceGroups.
When set, `maxSize` is mapped to `NodePool.spec.limits.nodes`, capping the number of nodes the NodePool may provision.

Karpenter does not allow an existing NodePool to transition between dynamic and static modes.
Delete the generated NodePool before running `kops update cluster` after adding or removing `minSize`.

## Interruption handling

On AWS, kOps provisions the SQS queue, the EventBridge rules and the IAM permissions that Karpenter needs for [interruption handling](https://karpenter.sh/docs/concepts/disruption/#interruption). The queue is shared with the Node Termination Handler, so an existing cluster keeps its queue when Karpenter replaces it.

Clusters without internet egress need a VPC endpoint for SQS. Karpenter starts without it, but processes no interruption events until it can reach the queue.

## Karpenter on GCE
{{ kops_feature_table(kops_added_default='1.38') }}

On GCE, the `karpenter.sh` addon deploys [karpenter-provider-gcp](https://github.com/karpenter-gcp/karpenter-provider-gcp) in
self-hosted provisioning mode: nodes bootstrap through the kOps nodeup script, and the controller uses only the Compute API
(no GKE APIs). The controller runs on the control plane and authenticates with Application Default Credentials from
the instance metadata server.

Enablement is the same as on AWS: set `spec.karpenter.enabled: true` on the cluster, and `spec.manager: Karpenter` with
`spec.role: Node` on the InstanceGroups Karpenter should manage. IRSA is not required on GCE.

kOps generates one `GCENodeClass` and one `NodePool` per Karpenter-managed InstanceGroup. The generated `GCENodeClass` uses:

* the InstanceGroup image (in `<project>/<name>` form, for example `ubuntu-os-cloud/ubuntu-2404-noble-amd64-v20260615`) translated into `imageSelectorTerms`
* a boot disk from the InstanceGroup `rootVolume` size, type, IOPS and throughput, with the same defaults as MIG instance templates
* `kubeletConfiguration` with `maxPods` matching the kubelet configuration (default 110), `systemReserved` and `kubeReserved`
* the InstanceGroup subnetwork; `enablePrivateNodes` follows the subnet type
* `subnetRangeNames` pointing to the cluster's pods secondary IP range on IP-alias clusters
* the kOps node network tag, service account, and instance ownership labels and metadata
* a Shielded VM configuration with vTPM enabled, required by the kOps node authorization
* the kOps nodeup bootstrap script as `startupScript`

A `kops update cluster` that changes an InstanceGroup's nodeup configuration updates its `startupScript`, and Karpenter
then replaces that NodeClass's nodes through drift, the same way AWS nodes roll when `userData` changes.

GCE-specific limitations:

* Requires the karpenter-provider-gcp release with self-hosted provisioning mode
  ([karpenter-gcp/karpenter-provider-gcp#540](https://github.com/karpenter-gcp/karpenter-provider-gcp/pull/540)).
* `kops delete cluster` and `kops toolbox dump` do not discover Karpenter-launched instances, because they are not part
  of any MIG. Delete the workloads or the generated NodePools (`kubectl delete nodepool <name>`) and wait for Karpenter to
  remove their instances before deleting the cluster.
* `kops rolling-update cluster` does not roll Karpenter-launched nodes; node replacement is owned by Karpenter (drift,
  consolidation, disruption budgets).
* `spec.zones` and `spec.associatePublicIP` are not applied to Karpenter-launched instances: Karpenter chooses zones
  within the region, and addressing follows the subnet type. To constrain zones, add a `topology.kubernetes.io/zone`
  requirement or node selector to the workloads.
* kOps-managed SSH public keys are not installed on Karpenter-launched instances; use OS Login or project-level SSH keys
  to access these nodes.
* `kops create cluster --instance-manager=karpenter` is not yet supported for GCE; enable Karpenter on an existing cluster instead.

## Known limitations

* **Upgrade is not supported** from the legacy AWS Karpenter integration (Karpenter v0.x, using the `Provisioner` and `AWSNodeTemplate` resources).
* Karpenter-managed InstanceGroups are only supported on AWS and GCE.
* Control plane InstanceGroups cannot be Karpenter-managed; the control plane uses an ASG on AWS or a MIG on GCE.
* Generated `EC2NodeClass` objects use `spec.amiFamily: Custom`.
* `spec.instanceStorePolicy` configuration is not supported in `EC2NodeClass`.
* `spec.kubelet` settings that affect Karpenter scheduling (`maxPods`, `systemReserved`, `kubeReserved`) are mapped to `EC2NodeClass.spec.kubelet` on AWS and `GCENodeClass.spec.kubeletConfiguration` on GCE, so Karpenter computes node allocatable capacity correctly. Other `spec.kubelet` settings are applied via the nodeup bootstrap script but are not surfaced to the NodeClass.
* `spec.rootVolume.optimization` has no `EC2NodeClass` equivalent and is ignored on Karpenter-managed InstanceGroups. Modern instance types are EBS-optimized by default.
* `spec.volumes` and `spec.volumeMounts` (additional, non-root volumes) are not mapped to `EC2NodeClass.spec.blockDeviceMappings`.
* The Karpenter controller policy grants `iam:PassRole` only for the kOps-managed node role. When a Karpenter-managed InstanceGroup uses a custom `spec.iam.profile`, kOps cannot determine the role inside the custom instance profile, so the policy allows passing any role to EC2 instead.
