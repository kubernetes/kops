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

function haveds() {
    local ds=0
    kubectl get ds -n kube-system aws-node-termination-handler --show-labels || ds=$?
    return $ds
}

# Verify we start with a DaemonSet
if ! haveds; then
  echo "Expected aws-node-termination-handler to exist"
  exit 1
fi

# Download latest kOps to upgrade the cluster
KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci-updown-green.txt)"
KOPS=$(mktemp -t kops.XXXXXXXXX)
wget -qO "${KOPS}" "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x "${KOPS}"

# Get cluster name from current context
CLUSTER_NAME=$(kubectl config view --minify -o jsonpath='{.clusters[0].name}')

# Switch to queue mode. This should remove the DS and install a Deployment instead
"${KOPS}" edit cluster "${CLUSTER_NAME}" "--set=cluster.spec.cloudProvider.aws.nodeTerminationHandler.enableSQSTerminationDraining=true"

"${KOPS}" update cluster
"${KOPS}" update cluster --yes

# Rolling-upgrade is needed so we get the new channels binary that supports prune
"${KOPS}" rolling-update cluster --instance-group-roles=master --yes

# just make sure pods are ready
"${KOPS}" validate cluster --wait=5m

# We should no longer have a daemonset called aws-node-termination-handler
if haveds; then
  echo "Expected aws-node-termination-handler to have been pruned"
  exit 1
fi

echo "Test passed: aws-node-termination-handler DaemonSet was successfully pruned after upgrade"
