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
  manager: Karpenter
```

### EC2NodeClass and NodePool

```sh
USER_DATA=$(aws s3 cp ${KOPS_STATE_STORE}/${NAME}/igconfig/node/nodes/nodeupscript.sh -)
USER_DATA=${USER_DATA//$'\n'/$'\n    '}

kubectl apply -f - <<YAML
apiVersion: karpenter.k8s.aws/v1
kind: EC2NodeClass
metadata:
  name: default
spec:
  amiFamily: Custom
  amiSelectorTerms:
    - ssmParameter: /aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id 
    - ssmParameter: /aws/service/canonical/ubuntu/server/24.04/stable/current/arm64/hvm/ebs-gp3/ami-id 
  associatePublicIPAddress: true
  tags:
    KubernetesCluster: ${NAME}
    kops.k8s.io/instancegroup: nodes
    k8s.io/role/node: "1"
  subnetSelectorTerms:
    - tags:
        KubernetesCluster: ${NAME}
  securityGroupSelectorTerms:
    - tags:
        KubernetesCluster: ${NAME}
        Name: nodes.${NAME}
  instanceProfile: nodes.${NAME}
  userData: |
    ${USER_DATA}
YAML

kubectl apply -f - <<YAML
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: default
spec:
  template:
    spec:
      requirements:
        - key: kubernetes.io/arch
          operator: In
          values: ["amd64", "arm64"]
        - key: kubernetes.io/os
          operator: In
          values: ["linux"]
        - key: karpenter.sh/capacity-type
          operator: In
          values: ["on-demand", "spot"]
      nodeClassRef:
        group: karpenter.k8s.aws
        kind: EC2NodeClass
        name: default
YAML
```

## Karpenter-managed InstanceGroups

A Karpenter-managed InstanceGroup controls the bootstrap script. kOps will ensure the correct AWS security groups, subnets and permissions.
`EC2NodeClass` and `NodePool` objects must be created by the cluster operator.

## Known limitations

* **Upgrade is not supported** from the previous version of managed Karpenter.
* Control plane nodes must be provisioned with an ASG.
* All `EC2NodeClass` objects must have the `spec.amiFamily` set to `Custom`.
* `spec.instanceStorePolicy` configuration is not supported in `EC2NodeClass`. 
* `spec.kubelet`, `spec.taints` and `spec.labels` configuration are not supported in `EC2NodeClass`, but they can be configured in the `Cluster` or `InstanceGroup` spec.
