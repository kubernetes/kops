# Karpenter

[Karpenter](https://karpenter.sh) is a Kubernetes-native capacity manager that directly provisions Nodes and underlying instances based on Pod requirements. On AWS, kOps supports managing an InstanceGroup with either Karpenter or an AWS Auto Scaling Group (ASG).

Karpenter is a fairly new project, and it is still not determined how Karpenter should work with kOps. Because of this, Karpenter is behind the `Karpenter` feature flag.

## Installing

Enable the Karpenter feature flag:

```sh
export KOPS_FEATURE_FLAGS="Karpenter"
```

Karpenter requires that external permissions for ServiceAccounts be enabled for the cluster. See [AWS IAM roles for ServiceAccounts documentation](/cluster_spec#service-account-issuer-discovery-and-aws-iam-roles-for-service-accounts-irsa) for how to enable this. 

### Existing clusters

On existing clusters, you can create a Karpenter InstanceGroup by adding the following to its InstanceGroup spec:

```yaml
spec:
  manager: Karpenter
```

You also need to enable the Karpenter addon in the cluster spec:

```yaml
spec:
  karpenter:
    enabled: true
```

### New clusters

On new clusters, you can simply add the `--instance-manager=karpenter` flag:

```sh
kops create cluster --name mycluster.example.com --cloud aws --networking=amazonvpc --zones=eu-central-1a,eu-central-1b --master-count=3 --yes --discovery-store=s3://discovery-store/
```

## Karpenter-managed InstanceGroups

A Karpenter-managed InstanceGroup controls a corresponding Karpenter Provisioner resource. kOps will ensure that the Provisioner is configured with the correct AWS security groups, subnets, and launch templates. Just like with ASG-managed InstanceGroups, you can add labels and taints to Nodes and kOps will ensure those are added accordingly.

Note that not all features of InstanceGroups are supported.

## Subnets

By default, kOps will tag subnets with `kops.k8s.io/instance-group/<intancegroup>: "true"` for each InstanceGroup the subnet is assigned to. If you enable manual tagging of subnets, you have to ensure these tags are added, if not Karpenter will fail to provision any instances.

## Instance Types

If you do not specify a mixed instances policy, only the instance type specified by `spec.machineType` will be used. With Karpenter, one typically wants a wider range of instances to choose from. kOps supports both providing a list of instance types through `spec.mixedInstancesPolicy.instances` and providing instance type requirements through `spec.mixedInstancesPolicy.instanceRequirements`. See (/instance_groups)[InstanceGroup documentation] for more details.

## Known limitations

### Karpenter-managed Launch Templates

On EKS, Karpener creates its own launch templates for Provisioners. These launch templates will not work with a kOps cluster for a number of reasons. Most importantly, they do not use supported AMIs and they do not install and configure nodeup, the instance-side kOps component. The Karpenter features that require Karpenter to directly manage launch templates will not be available on kOps.

### Unmanaged Provisioner resources

As mentioned above, kOps will manage a Provisioner resource per InstanceGroup. It is technically possible to create Provsioner resources directly, but you have to ensure that you configure Provisioners according to kOps requirements. As mentioned above, Karpenter-managed launch templates do not work and you have to maintain your own kOps-compatible launch templates.

### Other minor limitations

* Control plane nodes must be provisioned with an ASG, not Karpenter.
* Provisioners will unconditionally use spot instances
* Provisioners will unconditionally include burstable instance groups such as the T3 instance family.
* kOps will not allow mixing arm64 and amd64 instances in the same Provider.