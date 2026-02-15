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
TEST_ROOT="${REPO_ROOT}/tests/e2e/scenarios/addon-resource-tracking"

export KOPS_FEATURE_FLAGS="SpecOverrideFlag"

make test-e2e-install

# Download kOps 1.29.2 to create the initial cluster
export KOPS_BASE_URL="https://artifacts.k8s.io/binaries/kops/1.29.2"
KOPS_BIN=$(mktemp -t kops.XXXXXXXXX)
wget -qO "${KOPS_BIN}" "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x "${KOPS_BIN}"

# Create cluster with nodeTerminationHandler enabled (DaemonSet mode)
CREATE_ARGS="--networking calico"
CREATE_ARGS="${CREATE_ARGS} --set=cluster.spec.cloudProvider.aws.nodeTerminationHandler.enabled=true"
CREATE_ARGS="${CREATE_ARGS} --set=cluster.spec.cloudProvider.aws.nodeTerminationHandler.enableSQSTerminationDraining=false"

KUBETEST2_ARGS=()
KUBETEST2_ARGS+=("-v=2")
KUBETEST2_ARGS+=("--env=KOPS_FEATURE_FLAGS=${KOPS_FEATURE_FLAGS}")
KUBETEST2_ARGS+=("--kops-binary-path=${KOPS_BIN}")

kubetest2 kops \
    --up --down \
    "${KUBETEST2_ARGS[@]}" \
    --cloud-provider=aws \
    --create-args="${CREATE_ARGS}" \
    --kubernetes-version="1.29.8" \
    --test=exec -- "${TEST_ROOT}/test.sh"
