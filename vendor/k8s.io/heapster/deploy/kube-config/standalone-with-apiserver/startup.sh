#!/bin/bash

# Copyright 2016 The Kubernetes Authors.
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


# DEVELOPMENT ONLY
#
# This script helps build the environment for the Heapster API server.
# It is best to have the service running first:
#   ${KUBE_ROOT}/cluster/kubectl.sh --namespace=${HEAPSTER_NAMESPACE} \
#       create -f "${HEAPSTER_ROOT}/deploy/kube-config/standalone-with-apiserver/heapster-service.yaml",
#
# Once it is assigned the external IP, run this script to generate the needed certificates,
# basic auth config and kubeconfig, which will be stored in
# ${HEAPSTER_ROOT}/deploy/kube-config/standalone-with-apiserver/tmp/heapster/kubeconfig.
#
# It will also create the following objects in the ${HEAPSTER_NAMESPACE} namespace:
#   -  "heapster-apiserver-kubeconfig" secret
#   -  "heapster-apiserver-secrets" secret
#   -  deployment "heapster-apiserver"
#

# set these to the roots of kubernetes and heapster repos
KUBE_ROOT=$HOME/go/src/k8s.io/kubernetes
HEAPSTER_ROOT=$HOME/go/src/k8s.io/heapster

HEAPSTER_NAMESPACE=${HEAPSTER_NAMESPACE:-"kube-system"}
HEAPSTER_DEPLOYMENT_NAME=${HEAPSTER_DEPLOYMENT_NAME:-"heapster-apiserver"}

manifests_root="${HEAPSTER_ROOT}/deploy/kube-config/standalone-with-apiserver"
host_kubectl="${KUBE_ROOT}/cluster/kubectl.sh --namespace=${HEAPSTER_NAMESPACE}"

# If not yet created, create the service
#$host_kubectl create -f "${manifests_root}/heapster-service.yaml"
HEAPSTER_API_HOST="$($host_kubectl get -o=jsonpath svc/${HEAPSTER_DEPLOYMENT_NAME} --template '{.status.loadBalancer.ingress[*].ip}')"

source "${manifests_root}/common.sh"

### certificates
CONTEXT="heapster-apiserver"
MASTER_NAME="${HEAPSTER_DEPLOYMENT_NAME}" KUBE_TEMP="${HEAPSTER_ROOT}/deploy/kube-config/standalone-with-apiserver/tmp" create-apiserver-certs ${HEAPSTER_API_HOST}
echo "Generated certs"

### basic auth
KUBE_TEMP="${manifests_root}/tmp"
KUBECONFIG="${KUBE_TEMP}/heapster/kubeconfig"
create-auth-config
echo "Generated basic auth"

### kubeconfig file
create-heapster-kubeconfig

### create kubeconfig secret
KUBECONFIG="${KUBE_TEMP}/heapster/kubeconfig"
$host_kubectl create secret generic heapster-apiserver-kubeconfig --from-file="${KUBECONFIG}"

### fill the template with generated values
template="go run ${KUBE_ROOT}/federation/cluster/template.go"
$template "${manifests_root}/heapster-apiserver-secrets.template" > "${manifests_root}/heapster-apiserver-secrets.yaml"

# create secret and deployment
$host_kubectl create -f "${manifests_root}/heapster-apiserver-secrets.yaml"
$host_kubectl create -f "${manifests_root}/heapster-deployment.yaml"
