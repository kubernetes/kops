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
