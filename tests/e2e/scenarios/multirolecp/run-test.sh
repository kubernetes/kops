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

# Brings up a GCE cluster whose control plane is split across instance groups, two of which carry
# more than one role, and checks that each component came up where it was placed.
#
# The topology is declared as YAML rather than through create-cluster flags, which can only
# express a count and a size per role. The splitkcp scenario covers the flag-driven shape.

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(git rev-parse --show-toplevel)
TEST_ROOT="${REPO_ROOT}/tests/e2e/scenarios/multirolecp"
cd "${REPO_ROOT}/"

# Composite roles need ExperimentalRoles; the APIServer role needs APIServerNodes.
export KOPS_FEATURE_FLAGS=+APIServerNodes,+ExperimentalRoles

make test-e2e-install

KUBETEST2_ARGS=()
KUBETEST2_ARGS+=("-v=2")

if [[ "${JOB_TYPE:-}" == "presubmit" && "${REPO_OWNER:-}/${REPO_NAME:-}" == "kubernetes/kops" ]]; then
  KUBETEST2_ARGS+=("--build")
  KUBETEST2_ARGS+=("--kops-binary-path=${GOPATH}/src/k8s.io/kops/.build/dist/linux/$(go env GOARCH)/kops")
else
  KUBETEST2_ARGS+=("--kops-version-marker=${KOPS_VERSION_MARKER:-https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci.txt}")
fi

kubetest2 kops \
    --up --down \
    "${KUBETEST2_ARGS[@]}" \
    --cloud-provider=gce \
    --create-args="--zones=us-west1-a --gce-service-account=default" \
    --template-path="${TEST_ROOT}/cluster.yaml.tmpl" \
    --kubernetes-version=https://dl.k8s.io/release/stable.txt \
    --test=exec -- "${TEST_ROOT}/test.sh"
