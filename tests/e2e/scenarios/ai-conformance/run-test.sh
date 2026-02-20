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
SCENARIO_ROOT="${REPO_ROOT}/tests/e2e/scenarios/ai-conformance"

# Check for g6.xlarge availability in the region
echo "Checking availability of g6.xlarge in ${AWS_REGION}..."
(cd "${SCENARIO_ROOT}/tools/check-aws-availability" && go build -o check-aws-availability main.go)
AVAILABILITY=$("${SCENARIO_ROOT}/tools/check-aws-availability/check-aws-availability" -region "${AWS_REGION}" -instance-type g6.xlarge)
if [[ "${AVAILABILITY}" == "false" ]]; then
  echo "Error: g6.xlarge instances are not available in ${AWS_REGION}. Please choose a region with L4 GPU support."
  exit 1
fi
rm -f "${SCENARIO_ROOT}/tools/check-aws-availability/check-aws-availability"


kops-acquire-latest

# Cluster Configuration
# - Networking: Cilium with Gateway API enabled
# - Nodes: c5.large (we need some non-GPU nodes for non-GPU workloads)
# - NVIDIA driver and runtime are managed by GPU Operator (not kOps)
OVERRIDES="${OVERRIDES-} --networking=cilium"
OVERRIDES="${OVERRIDES} --set=cluster.spec.networking.cilium.gatewayAPI.enabled=true"
OVERRIDES="${OVERRIDES} --node-size=c5.large"
OVERRIDES="${OVERRIDES} --node-count=2"
OVERRIDES="${OVERRIDES} --zones=us-east-2a,us-east-2b,us-east-2c"

kops-up

# Now add an instance group for GPU nodes with the appropriate labels for NVIDIA DRA
# TODO: find zones, match images, etc. rather than hard-coding
${KOPS} create --name "${CLUSTER_NAME}" -f - <<EOF
apiVersion: kops.k8s.io/v1alpha2
kind: InstanceGroup
metadata:
  name: gpu-nodes
  labels:
    kops.k8s.io/cluster: ${CLUSTER_NAME}
spec:
  image: 099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20251212
  machineType: g6.xlarge
  maxSize: 3
  minSize: 1
  role: Node
  rootVolumeSize: 48
  subnets:
  - us-east-2a
  - us-east-2b
  - us-east-2c
EOF

${KOPS} update cluster --name "${CLUSTER_NAME}" --yes --admin

echo "----------------------------------------------------------------"
echo "Deploying AI Conformance Components"
echo "----------------------------------------------------------------"

# 0. Gateway API CRDs (Required for Cilium)
echo "Installing Gateway API CRDs v1.2.0..."
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml

# cert-manager: required for KubeRay webhooks
echo "Installing cert-manager..."
kubectl apply --server-side -f https://github.com/cert-manager/cert-manager/releases/download/v1.19.2/cert-manager.yaml

# Setup helm repo for NVIDIA GPU Operator and DRA Driver
helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
helm repo update

# NVIDIA GPU Operator
# Manages the full NVIDIA stack: kernel driver, container toolkit, device plugin.
# The driver is installed into /run/nvidia/driver on each node.
helm upgrade -i nvidia-gpu-operator --wait \
    -n gpu-operator --create-namespace \
    nvidia/gpu-operator \
    --version=v25.10.1 \
    --wait

PATH="$(pwd):$PATH"
export PATH

# NVIDIA DRA Driver
# Uses the driver installed by GPU Operator at /run/nvidia/driver (the default).
echo "Installing NVIDIA DRA Driver..."

cat > values.yaml <<EOF
# The driver daemonset needs a toleration for the nvidia.com/gpu taint
kubeletPlugin:
  tolerations:
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
EOF

helm upgrade -i nvidia-dra-driver-gpu nvidia/nvidia-dra-driver-gpu \
    --version="25.12.0" \
    --create-namespace \
    --namespace nvidia-dra-driver-gpu \
    --set resources.gpus.enabled=true \
    --set nvidiaDriverRoot=/run/nvidia/driver \
    --set gpuResourcesEnabledOverride=true \
    -f values.yaml \
    --wait


# KubeRay
echo "Installing KubeRay Operator..."
kubectl apply --server-side -k "github.com/ray-project/kuberay/ray-operator/config/default-with-webhooks?ref=v1.5.0"

# Kueue
echo "Installing Kueue..."
kubectl apply --server-side -f https://github.com/kubernetes-sigs/kueue/releases/download/v0.14.8/manifests.yaml


echo "----------------------------------------------------------------"
echo "Verifying Cluster and Components"
echo "----------------------------------------------------------------"

# Wait for kOps validation
"${KOPS}" validate cluster --wait=15m

# Verify Components
echo "Verifying NVIDIA Device Plugin..."
#kubectl rollout status daemonset -n kube-system nvidia-device-plugin-daemonset --timeout=5m || echo "Warning: NVIDIA Device Plugin not ready yet"

echo "Verifying Kueue..."
kubectl rollout status deployment -n kueue-system kueue-controller-manager --timeout=5m || echo "Warning: Kueue not ready yet"

echo "Verifying KubeRay..."
kubectl rollout status deployment -n ray-system kuberay-operator --timeout=5m || echo "Warning: KubeRay not ready yet"

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
  devices:
    requests:
    - name: single-gpu
      exactly:
        deviceClassName: gpu.nvidia.com
        allocationMode: ExactCount
        count: 1
---
apiVersion: batch/v1
kind: Job
metadata:
  name: test-gpu-pod
spec:
  template:
    spec:
      restartPolicy: Never
      tolerations:
      - key: "nvidia.com/gpu"
        operator: "Exists"
        effect: "NoSchedule"
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
kubectl wait --for=condition=complete job/test-gpu-pod --timeout=5m || true
kubectl logs job/test-gpu-pod || echo "Failed to get logs"

echo "AI Conformance Environment Setup Complete."

# Now run the actual AI conformance tests
cd "${REPO_ROOT}/tests/e2e/scenarios/ai-conformance/validators"
go test -v ./... -timeout=60m
