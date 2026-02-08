#!/usr/bin/env bash

# Copyright 2026 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

# Get the cluster name from the KUBECONFIG context
CLUSTER_NAME=$(kubectl config view --minify -o jsonpath='{.clusters[0].name}')

# Get KOPS_STATE_STORE from kubeconfig server URL (the bucket name is embedded)
# Alternatively, we can get it from environment if kubetest2 passes it
if [[ -z "${KOPS_STATE_STORE:-}" ]]; then
  echo "KOPS_STATE_STORE must be set"
  exit 1
fi

# Get the node instance group user data script from the kOps state store
USER_DATA=$(aws s3 cp "${KOPS_STATE_STORE}/${CLUSTER_NAME}/igconfig/node/nodes/nodeupscript.sh" -)
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

# Download kops to validate cluster
KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci-updown-green.txt)"
KOPS=$(mktemp -t kops.XXXXXXXXX)
wget -qO "${KOPS}" "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x "${KOPS}"

# Wait for the nodes to be ready
"${KOPS}" validate cluster --wait=10m

if [[ -z "${K8S_VERSION:-}" ]]; then
  K8S_VERSION="$(curl -s -L https://dl.k8s.io/release/stable.txt)"
fi

# Download test binaries
BINDIR=$(mktemp -d)
wget -qO- "https://dl.k8s.io/${K8S_VERSION}/kubernetes-test-linux-amd64.tar.gz" | tar xz -C "${BINDIR}" --strip-components=3 kubernetes/test/bin/e2e.test kubernetes/test/bin/ginkgo

# Run conformance tests
"${BINDIR}/ginkgo" \
    --nodes=20 \
    --focus="\[Conformance\]" \
    --no-color \
    "${BINDIR}/e2e.test" \
    -- \
    --provider=skeleton \
    --kubeconfig="${KUBECONFIG:-${HOME}/.kube/config}" \
    --report-dir="${ARTIFACTS:-/tmp}"
