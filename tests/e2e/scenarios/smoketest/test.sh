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

if [[ -n "${KOPS_SKIP_E2E:-}" ]]; then
  echo "Skipping e2e tests (KOPS_SKIP_E2E is set)"
  exit 0
fi

if [[ -z "${K8S_VERSION:-}" ]]; then
  echo "K8S_VERSION must be set"
  exit 1
fi

# Download test binaries
BINDIR=$(mktemp -d)
wget -qO- "https://dl.k8s.io/${K8S_VERSION}/kubernetes-test-linux-amd64.tar.gz" | tar xz -C "${BINDIR}" --strip-components=3 kubernetes/test/bin/e2e.test kubernetes/test/bin/ginkgo

# Run e2e tests (skipping slow/serial/disruptive tests)
"${BINDIR}/ginkgo" \
    --nodes=25 \
    --skip="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler" \
    --no-color \
    "${BINDIR}/e2e.test" \
    -- \
    --provider=skeleton \
    --kubeconfig="${KUBECONFIG:-${HOME}/.kube/config}" \
    --report-dir="${ARTIFACTS:-/tmp}"
