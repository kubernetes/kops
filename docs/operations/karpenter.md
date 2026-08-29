# Karpenter

[Karpenter](https://karpenter.sh) is an open-source node lifecycle management project built for Kubernetes.
Adding Karpenter to a Kubernetes cluster can dramatically improve the efficiency and cost of running workloads on that cluster.

On AWS, kOps supports managing an InstanceGroup with either Karpenter or an AWS Auto Scaling Group (ASG).

## Prerequisites

Managed Karpenter requires kOps 1.34+ and that [IAM Roles for Service Accounts (IRSA)](/cluster_spec#service-account-issuer-discovery-and-aws-iam-roles-for-service-accounts-irsa) be enabled for the cluster.

If an older version of Karpenter was installed, it must be uninstalled before installing the new version.

## Installing

### New clusters

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

### EC2NodeClass and NodePool
{{ kops_feature_table(kops_added_default='1.36') }}

kOps generates one `EC2NodeClass` and one `NodePool` for each AWS node InstanceGroup with `spec.manager: Karpenter`.
The generated objects use the InstanceGroup name, are delivered by the `karpenter.sh` addon, and are pruned when the InstanceGroup is removed.

The generated `EC2NodeClass` uses:

* `amiFamily: Custom`
* the InstanceGroup image translated into `amiSelectorTerms`
* the InstanceGroup `spec.rootVolume` translated into `blockDeviceMappings`
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

This generates the appropriate `karpenter.k8s.aws/instance-cpu` and `karpenter.k8s.aws/instance-memory` requirements on the `NodePool`.
`excludedInstanceTypes` entries are mapped to `NotIn` requirements: a `<family>.*` wildcard excludes an instance family, and a bare instance type excludes that type.

When `instanceRequirements` is set, neither `spec.machineType` nor `spec.mixedInstancesPolicy.instances` is required.
If either field is set, its instance types further restrict the generated `NodePool`.

Karpenter NodePools can include both GPU and non-GPU instance types.
When using kOps-managed NVIDIA support, use a dedicated GPU-only InstanceGroup because kOps applies GPU labels and taints to the entire NodePool.

## Karpenter-managed InstanceGroups
{{ kops_feature_table(kops_added_default='1.36') }}

A Karpenter-managed InstanceGroup controls the bootstrap script. kOps ensures the correct AWS security groups, subnets, permissions, and Karpenter resource definitions.

When `minSize` is omitted, kOps generates a dynamic NodePool and Karpenter owns scale-out decisions.
For a static NodePool, set `minSize` to a positive number:

```yaml
spec:
  role: Node
  manager: Karpenter
  minSize: 4
```

For new clusters, `--instance-manager=karpenter --node-count=4` creates the same static configuration.
Zero and negative `minSize` values are rejected.

The Karpenter addon enables `StaticCapacity` by default.
If `cluster.spec.karpenter.featureGates` is customized, it must include `StaticCapacity=true` for static InstanceGroups.
When set, `maxSize` is mapped to `NodePool.spec.limits.nodes`, capping the number of nodes the NodePool may provision.

Karpenter does not allow an existing NodePool to transition between dynamic and static modes.
Delete the generated NodePool before running `kops update cluster` after adding or removing `minSize`.

## Known limitations

* **Upgrade is not supported** from the legacy Karpenter integration (Karpenter v0.x, using the `Provisioner` and `AWSNodeTemplate` resources).
* Karpenter-managed InstanceGroups are only supported on AWS.
* Control plane nodes must be provisioned with an ASG.
* Generated `EC2NodeClass` objects use `spec.amiFamily: Custom`.
* `spec.instanceStorePolicy` configuration is not supported in `EC2NodeClass`.
* `spec.kubelet` settings that affect Karpenter scheduling (`maxPods`, `systemReserved`, `kubeReserved`) are mapped to `EC2NodeClass.spec.kubelet` so Karpenter computes node allocatable capacity correctly. Other `spec.kubelet` settings are applied via the nodeup bootstrap script but are not surfaced to `EC2NodeClass`.
* `spec.rootVolume.optimization` has no `EC2NodeClass` equivalent and is ignored on Karpenter-managed InstanceGroups. Modern instance types are EBS-optimized by default.
* `spec.volumes` and `spec.volumeMounts` (additional, non-root volumes) are not mapped to `EC2NodeClass.spec.blockDeviceMappings`.
* The Karpenter controller policy grants `iam:PassRole` only for the kOps-managed node role. When a Karpenter-managed InstanceGroup uses a custom `spec.iam.profile`, kOps cannot determine the role inside the custom instance profile, so the policy allows passing any role to EC2 instead.
