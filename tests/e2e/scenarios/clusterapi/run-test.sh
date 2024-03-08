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

# Enable feature flag for CAPI support
export KOPS_FEATURE_FLAGS=ClusterAPI

# Override some settings
CLUSTER_NAME="clusterapi.k8s.local"

CLOUD_PROVIDER=gce
ZONES=us-east4-a # Currently the zone name gets encoded in the machinedeployment name (maybe we can use labels instead?)

OVERRIDES="${OVERRIDES-} --node-count=2" # We need at least 2 nodes for CoreDNS to validate
OVERRIDES="${OVERRIDES} --gce-service-account=default" # Use default service account because boskos permissions are limited

# Create kOps cluster
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

kops-acquire-latest

kops-up

# Export KUBECONFIG; otherwise the precedence for controllers is wrong (?)
KUBECONFIG=$(mktemp -t kops.XXXXXXXXX)
export KUBECONFIG
"${KOPS}" export kubecfg --name "${CLUSTER_NAME}" --admin --kubeconfig "${KUBECONFIG}"


# Install kOps CRDs (Cluster & InstanceGroup)
# Ideally we would install these as part of kops-up, but while we're developing CAPI support it's easier to do it here
kubectl apply --server-side -k "${REPO_ROOT}/k8s"
kubectl apply --server-side -k "${REPO_ROOT}/clusterapi/config"

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
kubectl apply --server-side -k "${REPO_ROOT}/clusterapi/manifests/cluster-api-provider-gcp"

kubectl wait --for=condition=Available --timeout=5m -n capi-system deployment/capi-controller-manager
kubectl wait --for=condition=Available --timeout=5m -n capg-system deployment/capg-controller-manager


# Install extra RBAC for capi-manager loopback connection to cluster (used to check node health etc)
kubectl apply --server-side -f "${REPO_ROOT}/clusterapi/examples/capi-loopback.yaml"

# Create a Cluster API Cluster object
"${KOPS}" get cluster clusterapi.k8s.local -oyaml | kubectl apply --server-side -n kube-system -f -

# Create a MachineDeployment matching our configuration
"${KOPS}" toolbox clusterapi generate machinedeployment \
  --cluster clusterapi.k8s.local \
  --name clusterapi-k8s-local-md-0 \
  --namespace kube-system | kubectl apply --server-side -n kube-system -f -

# Debug: print output from kops-controller
kubectl logs -n kube-system -l k8s-app=kops-controller --follow &

# Wait for the MachineDeployment machines to become ready
kubectl wait --for=condition=Available -n kube-system machinedeployment/clusterapi-k8s-local-md-0-us-east4-a --timeout=10m
kubectl wait --for=condition=MachinesReady -n kube-system machinedeployment/clusterapi-k8s-local-md-0-us-east4-a --timeout=10m
kubectl get -n kube-system machinedeployment/clusterapi-k8s-local-md-0-us-east4-a -oyaml

# Print the nodes, machines and gcpmachines
kubectl get nodes -owide
kubectl get machine -A -owide
kubectl get gcpmachine -A -owide

# Delete the machinedeployment, causing the machines and nodes to be deleted
kubectl delete -n kube-system machinedeployment/clusterapi-k8s-local-md-0-us-east4-a

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
