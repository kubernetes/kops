#!/usr/bin/env bash

# Copyright 2021 The Kubernetes Authors.
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

export KOPS_FEATURE_FLAGS="SpecOverrideFlag"
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

function haveds() {
	local ds=0
	kubectl get ds -n kube-system aws-node-termination-handler --show-labels || ds=$?
	return $ds
}

# Start a cluster with an old version of channel

export KOPS_BASE_URL
KOPS_BASE_URL="https://artifacts.k8s.io/binaries/kops/1.29.2"
KOPS=$(kops-download-from-base)

# Start with a cluster running nodeTerminationHandler
ARGS="--set=cluster.spec.cloudProvider.aws.nodeTerminationHandler.enabled=true"
ARGS="${ARGS} --set=cluster.spec.cloudProvider.aws.nodeTerminationHandler.enableSQSTerminationDraining=false"

${KUBETEST2} \
    --up \
    --kubernetes-version="1.29.8" \
    --kops-binary-path="${KOPS}" \
    --create-args="$ARGS"


if ! haveds; then
  echo "Expected aws-node-termination-handler to exist"
  exit 1
fi

# Upgrade to a version that should adopt existing resources and apply the change below
kops-acquire-latest

cp "${KOPS}" "${WORKSPACE}/kops"

# Switch to queue mode. This should remove the DS and install a Deployment instead
kops edit cluster "${CLUSTER_NAME}" "--set=cluster.spec.cloudProvider.aws.nodeTerminationHandler.enableSQSTerminationDraining=true"

# allow downgrade is a bug where the version written to VFS is not the same as the running version.
kops update cluster --allow-kops-downgrade
kops update cluster --yes --allow-kops-downgrade

# Rolling-upgrade is needed so we get the new channels binary that supports prune
kops rolling-update cluster --instance-group-roles=master --yes

# just make sure pods are ready
kops validate cluster --wait=5m

# We should no longer have a daemonset called aws-node-termination-handler
if haveds; then
  echo "Expected aws-node-termination-handler to have been pruned"
  exit 1
fi