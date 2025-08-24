# Karpenter

[Karpenter](https://karpenter.sh) is a Kubernetes-native capacity manager that directly provisions Nodes and underlying instances based on Pod requirements. On AWS, kOps supports managing an InstanceGroup with either Karpenter or an AWS Auto Scaling Group (ASG).

## Prerequisites

Managed Karpenter requires kOps 1.34+ and that [IAM Roles for Service Accounts (IRSA)](/cluster_spec#service-account-issuer-discovery-and-aws-iam-roles-for-service-accounts-irsa) be enabled for the cluster.

## Installing

### New clusters

```sh
export NAME="my-cluster.example.com"
export REGION="us-east-1"
export ZONE="us-east-1a"

kops create cluster --name ${NAME} \
  --state=s3://my-state-store \
  --discovery-store=s3://my-discovery-store \
  --cloud=aws \
  --networking=cilium \
  --zones=${ZONE} \
  --instance-manager=karpenter \
  --yes
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
export USER_DATA=$(aws s3 cp s3://my-state-store/${NAME}/igconfig/node/nodes/nodeupscript.sh -)

cat <<EOF | kubectl apply -f -
apiVersion: karpenter.k8s.aws/v1
kind: EC2NodeClass
metadata:
  name: default
spec:
  amiFamily: Custom
  amiSelectorTerms:
    - ssmParameter: /aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id 
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
$(echo "$USER_DATA" | sed 's/^/    /')
EOF

cat <<EOF | kubectl apply -f -
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
      expireAfter: 24h
  limits:
    cpu: 4
  disruption:
    consolidationPolicy: WhenEmptyOrUnderutilized
    consolidateAfter: 1m
EOF
```

## Karpenter-managed InstanceGroups

A Karpenter-managed InstanceGroup controls the bootstrap script. kOps will ensure the correct AWS security groups, subnets and permissions.
`EC2NodeClass` and `NodePool` objects must be created by the operator.

## Known limitations

* **Upgrade is not supported** from the previous version of managed Karpenter.
* Control plane nodes must be provisioned with an ASG.
* All `EC2NodeClass` objects must have the `spec.amiFamily` set to `Custom`.
* `spec.instanceStorePolicy` configuration is not supported in `EC2NodeClass`. 
* `spec.kubelet`, `spec.taints` and `spec.labels` configuration are not supported in `EC2NodeClass`, but they can be configured in the `Cluster` or `InstanceGroup` spec.
