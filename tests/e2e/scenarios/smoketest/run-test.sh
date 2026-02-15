#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
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
TEST_ROOT="${REPO_ROOT}/tests/e2e/scenarios/smoketest"

make test-e2e-install

if [ -z "${KOPS_VERSION:-}" ] || [ -z "${K8S_VERSION:-}" ]; then
  >&2 echo "must set KOPS_VERSION and K8S_VERSION env vars"
  exit 1
fi

# Download the specified kOps release
KOPS_BIN=$(mktemp -t kops.XXXXXXXXX)
wget -qO "${KOPS_BIN}" "https://github.com/kubernetes/kops/releases/download/${KOPS_VERSION}/kops-$(go env GOOS)-$(go env GOARCH)"
chmod +x "${KOPS_BIN}"

KUBETEST2_ARGS=()
KUBETEST2_ARGS+=("-v=2")
KUBETEST2_ARGS+=("--kops-binary-path=${KOPS_BIN}")

kubetest2 kops \
    --up --down \
    "${KUBETEST2_ARGS[@]}" \
    --cloud-provider=aws \
    --create-args="--networking calico" \
    --kubernetes-version="${K8S_VERSION}" \
    --env=K8S_VERSION="${K8S_VERSION}" \
    --env=KOPS_SKIP_E2E="${KOPS_SKIP_E2E:-}" \
    --test=exec -- "${TEST_ROOT}/test.sh"
