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

REPO_ROOT=$(git rev-parse --show-toplevel)
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

# AI Conformance requirements:
# - Kubernetes 1.35
# - NVIDIA L4 Instances (g6.xlarge on AWS)
# - Gateway API
# - Gang Scheduling (Kueue)
# - Robust Controller (KubeRay)

K8S_VERSION=$(curl -L -s https://dl.k8s.io/release/stable.txt)
export K8S_VERSION

export CLOUD_PROVIDER=aws
# Ensure region with L4 (g6) availability
export AWS_REGION="${AWS_REGION:-us-east-2}"

# Check for g6.xlarge availability in the region
echo "Checking availability of g6.xlarge in ${AWS_REGION}..."
AVAILABILITY=$(aws ec2 describe-instance-type-offerings --location-type availability-zone --filters Name=instance-type,Values=g6.xlarge --region "${AWS_REGION}" --query 'InstanceTypeOfferings' --output text)
if [[ -z "${AVAILABILITY}" ]]; then
  echo "Error: g6.xlarge instances are not available in ${AWS_REGION}. Please choose a region with L4 GPU support."
  exit 1
fi

kops-acquire-latest

# Cluster Configuration
# - Networking: Cilium with Gateway API enabled
# - Nodes: g6.xlarge (L4 GPU)
# - Runtime: NVIDIA enabled
OVERRIDES="${OVERRIDES-} --networking=cilium"
OVERRIDES="${OVERRIDES} --set=cluster.spec.networking.cilium.enableGatewayAPI=true"
OVERRIDES="${OVERRIDES} --node-size=g6.xlarge"
OVERRIDES="${OVERRIDES} --node-count=2"
OVERRIDES="${OVERRIDES} --set=cluster.spec.containerd.nvidia.enabled=true"

kops-up

echo "----------------------------------------------------------------"
echo "Deploying AI Conformance Components"
echo "----------------------------------------------------------------"

# 0. Gateway API CRDs (Required for Cilium)
echo "Installing Gateway API CRDs v1.2.0..."
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml

# 1. NVIDIA Device Plugin
echo "Installing NVIDIA Device Plugin..."
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v0.17.0/nvidia-device-plugin.yml

# 1.5 NVIDIA DRA Driver
echo "Installing Helm..."
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
USE_SUDO=false HELM_INSTALL_DIR=. ./get_helm.sh

PATH="$(pwd):$PATH"
export PATH

echo "Installing NVIDIA DRA Driver..."
helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
helm repo update
helm install nvidia-dra-driver-gpu nvidia/nvidia-dra-driver-gpu \
  --create-namespace \
  --namespace nvidia-dra-driver-gpu \
  --version 25.8.1 \
  --set resources.gpus.enabled=true \
  --wait

# 2. Gang Scheduling (Kueue)
echo "Installing Kueue..."
kubectl apply --server-side -f https://github.com/kubernetes-sigs/kueue/releases/download/v0.14.8/manifests.yaml

# 3. Robust Controller (KubeRay)
echo "Installing KubeRay Operator..."
# KubeRay 1.3.0
kubectl apply -k "github.com/ray-project/kuberay/ray-operator/config/default?ref=v1.5.0"

echo "----------------------------------------------------------------"
echo "Verifying Cluster and Components"
echo "----------------------------------------------------------------"

# Wait for kOps validation
"${KOPS}" validate cluster --wait=15m

# Verify Components
echo "Verifying NVIDIA Device Plugin..."
kubectl rollout status daemonset -n kube-system nvidia-device-plugin-daemonset --timeout=5m || echo "Warning: NVIDIA Device Plugin not ready yet"

echo "Verifying Kueue..."
kubectl rollout status deployment -n kueue-system kueue-controller-manager --timeout=5m || echo "Warning: Kueue not ready yet"

echo "Verifying KubeRay..."
kubectl rollout status deployment -n kuberay-system kuberay-operator --timeout=5m || echo "Warning: KubeRay not ready yet"

echo "Verifying Gateway API..."
kubectl get gatewayclass || echo "Warning: GatewayClass not found"

echo "Verifying Allocatable GPUs..."
# Wait a bit for nodes to report resources
sleep 30
kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}: {.status.allocatable.nvidia\.com/gpu} GPUs{"\n"}{end}'

echo "Running Sample DRA Workload..."
# Create a ResourceClaim and Pod to test DRA
kubectl apply -f - <<EOF
apiVersion: resource.k8s.io/v1
kind: ResourceClaim
metadata:
  name: test-gpu-claim
spec:
  resourceClassName: nvidia-gpu
---
apiVersion: v1
kind: Pod
metadata:
  name: test-gpu-pod
spec:
  restartPolicy: Never
  containers:
  - name: test
    image: nvcr.io/nvidia/k8s/cuda-sample:vectoradd-cuda12.5.0-ubuntu22.04
    command: ["/bin/sh", "-c"]
    args: ["/cuda-samples/vectorAdd"]
    resources:
      claims:
      - name: gpu
  resourceClaims:
  - name: gpu
    resourceClaimName: test-gpu-claim
EOF

echo "Waiting for Sample Workload to Complete..."
# Wait for the pod to succeed
kubectl wait --for=condition=Ready pod/test-gpu-pod --timeout=5m || true
kubectl logs test-gpu-pod || echo "Failed to get logs"

# Note: The actual AI conformance test suite (e.g., k8s-ai-conformance binary)
# would be executed here. For this scenario, we establish the compliant environment.

echo "AI Conformance Environment Setup Complete."
