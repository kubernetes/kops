#!/usr/bin/env bash

# Copyright 2025 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

REPO_ROOT=$(git rev-parse --show-toplevel);
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

kops-acquire-latest

NETWORKING=cilium
OVERRIDES="${OVERRIDES-} --instance-manager=karpenter"
OVERRIDES="${OVERRIDES} --control-plane-size=c6g.large"
OVERRIDES="${OVERRIDES} --set=cluster.spec.karpenter.featureGates=StaticCapacity=true"

kops-up

# Get the node instance group user data script from the kOps state store
USER_DATA=$(aws s3 cp "${KOPS_STATE_STORE-}/${CLUSTER_NAME}/igconfig/node/nodes/nodeupscript.sh" -)
# Indent the user data script for embedding in the EC2NodeClass
USER_DATA=${USER_DATA//$'\n'/$'\n    '}

# Create a EC2NodeClass for Karpenter
kubectl apply -f - <<YAML
apiVersion: karpenter.k8s.aws/v1
kind: EC2NodeClass
metadata:
  name: default
spec:
  amiFamily: Custom
  amiSelectorTerms:
    - ssmParameter: /aws/service/canonical/ubuntu/server/24.04/stable/current/arm64/hvm/ebs-gp3/ami-id
  associatePublicIPAddress: true
  tags:
    KubernetesCluster: ${CLUSTER_NAME}
    kops.k8s.io/instancegroup: nodes
    k8s.io/role/node: "1"
  subnetSelectorTerms:
    - tags:
        KubernetesCluster: ${CLUSTER_NAME}
  securityGroupSelectorTerms:
    - tags:
        KubernetesCluster: ${CLUSTER_NAME}
        Name: nodes.${CLUSTER_NAME}
  instanceProfile: nodes.${CLUSTER_NAME}
  userData: |
    ${USER_DATA}
YAML

# Create a NodePool for Karpenter
# Effectively disable consolidation for 30 minutes to avoid flakes in the tests
kubectl apply -f - <<YAML
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: default
spec:
  template:
    spec:
      requirements:
        - key: node.kubernetes.io/instance-type
          operator: In
          values: ["m6g.large"]
        - key: karpenter.sh/capacity-type
          operator: In
          values: ["on-demand"]
      nodeClassRef:
        group: karpenter.k8s.aws
        kind: EC2NodeClass
        name: default
  replicas: 4
  disruption:
    consolidationPolicy: WhenEmpty
    consolidateAfter: 30m
YAML

# Wait for the nodes to start being provisioned
sleep 30
# Wait for the nodes to be ready
"${KOPS}" validate cluster --wait=10m

# Run the tests
cp "${KOPS}" "${WORKSPACE}/kops"
${KUBETEST2} \
  --test=kops \
  --kops-binary-path="${KOPS}" \
  -- \
  --test-package-version="${K8S_VERSION}" \
  --focus-regex="\[Conformance\]" \
  --parallel 20
