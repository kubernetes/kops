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


set -euo pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}/"

# Enable feature flag for APIServer Nodes support and split control plane
export KOPS_FEATURE_FLAGS=+APIServerNodes,+ExperimentalRoles

# Override some settings
CLUSTER_NAME="splitkcp.k8s.local"

CLOUD_PROVIDER=gce
ZONES="us-west1-a,us-west1-b,us-west1-c" # Currently the zone name gets encoded in the machinedeployment name (maybe we can use labels instead?)

OVERRIDES="${OVERRIDES-} --node-count=2" # We need at least 2 nodes for CoreDNS to validate
OVERRIDES="${OVERRIDES} --control-plane-count=0" # Turning off full control plane nodes and turning on split types
OVERRIDES="${OVERRIDES} --api-server-count=2" # We need at least 1 api-server node to split the control plane
OVERRIDES="${OVERRIDES} --api-server-size=e2-standard-2" #
OVERRIDES="${OVERRIDES} --etcd-count=3" # We need at least 3 etcd nodes to have quorum
OVERRIDES="${OVERRIDES} --etcd-size=e2-standard-2" #
OVERRIDES="${OVERRIDES} --kcm-count=2" # We need at least 1 kube-controller-manager node to split the control plane
OVERRIDES="${OVERRIDES} --kcm-size=e2-standard-2" #
OVERRIDES="${OVERRIDES} --scheduler-count=2" # We need at least 1 kube-controller-manager node to split the control plane
OVERRIDES="${OVERRIDES} --scheduler-size=e2-standard-2" #
OVERRIDES="${OVERRIDES} --networking=kubenet" # :( Need to work out why we need this.
OVERRIDES="${OVERRIDES} --node-size=e2-standard-2" #
OVERRIDES="${OVERRIDES} --gce-service-account=default" # Use default service account because boskos permissions are limited

# Create kOps cluster
source "${REPO_ROOT}/tests/e2e/scenarios/lib/common.sh"

kops-acquire-latest

kops-up

# Export KUBECONFIG; otherwise the precedence for controllers is wrong (?)
KUBECONFIG=$(mktemp -t kops.XXXXXXXXX)
export KUBECONFIG
"${KOPS}" export kubecfg --name "${CLUSTER_NAME}" --admin --kubeconfig "${KUBECONFIG}"


# Install kOps CRDs (Cluster & InstanceGroup)
# Ideally we would install these as part of kops-up, but while we're developing CAPI support it's easier to do it here
kubectl apply --server-side -k "${REPO_ROOT}/k8s"
#kubectl apply --server-side -k "${REPO_ROOT}/clusterapi/config"

# Install extra RBAC for kops-controller CAPI support
kubectl apply --server-side -f "${REPO_ROOT}/clusterapi/examples/kopscontroller.yaml"

# Bounce kops-controller in case it went into backoff before the CRDs were installed
kubectl delete pod -n kube-system -l k8s-app=kops-controller

# Install cert-manager
kubectl apply --server-side -f https://github.com/cert-manager/cert-manager/releases/download/v1.18.2/cert-manager.yaml

kubectl wait --for=condition=Available --timeout=5m -n cert-manager deployment/cert-manager
kubectl wait --for=condition=Available --timeout=5m -n cert-manager deployment/cert-manager-cainjector
kubectl wait --for=condition=Available --timeout=5m -n cert-manager deployment/cert-manager-webhook

# Install cluster-api core and cluster-api-provider-gcp
kubectl apply --server-side -k "${REPO_ROOT}/clusterapi/manifests/cluster-api"
kubectl wait --for=condition=Available --timeout=5m -n capi-system deployment/capi-controller-manager

kubectl apply --server-side -k "${REPO_ROOT}/clusterapi/manifests/cluster-api-provider-gcp"
kubectl wait --for=condition=Available --timeout=5m -n capg-system deployment/capg-controller-manager


# Install extra RBAC for capi-manager loopback connection to cluster (used to check node health etc)
kubectl apply --server-side -f "${REPO_ROOT}/clusterapi/examples/capi-loopback.yaml"

# Debug: log kops-controller
kubectl logs -n kube-system -l k8s-app=kops-controller --follow &

# Print the nodes, machines and gcpmachines
kubectl get nodes -owide
kubectl get machine -A -owide
kubectl get gcpmachine -A -owide

# Print the nodes, machines and gcpmachines again
kubectl get nodes -owide
kubectl get machine -A -owide
kubectl get gcpmachine -A -owide

# CAPI currently creates some firewall rules that otherwise are not cleaned up, and block kops cluster cleanup
function cleanup_capi_leaks() {
  gcloud compute firewall-rules delete allow-clusterapi-k8s-local-cluster --quiet || true
  gcloud compute firewall-rules delete allow-clusterapi-k8s-local-healthchecks --quiet || true

  #gcloud compute networks subnets delete us-east4-clusterapi-k8s-local --region us-east4 --quiet
  #gcloud compute networks delete clusterapi-k8s-local --quiet
}
cleanup_capi_leaks
