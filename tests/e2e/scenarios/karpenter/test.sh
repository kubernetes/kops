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

# Wait for the generated static NodePool.
kubectl wait --for=create nodepool/nodes --timeout=5m
test "$(kubectl get nodepool/nodes -o jsonpath='{.spec.replicas}')" = "4"
kubectl wait --for=jsonpath='{.status.nodes}'=4 nodepool/nodes --timeout=15m

# Wait for the cluster to be ready.
"${KOPS}" validate cluster --wait=10m

if [[ -z "${K8S_VERSION:-}" ]]; then
  K8S_VERSION="$(curl -s -L https://dl.k8s.io/release/stable.txt)"
fi

# Download test binaries
BINDIR=$(mktemp -d)
wget -qO- "https://dl.k8s.io/${K8S_VERSION}/kubernetes-test-linux-amd64.tar.gz" | tar xz -C "${BINDIR}" --strip-components=3 kubernetes/test/bin/e2e.test kubernetes/test/bin/ginkgo

# Run conformance tests
"${BINDIR}/ginkgo" \
    --nodes=20 \
    --focus="\[Conformance\]" \
    --no-color \
    "${BINDIR}/e2e.test" \
    -- \
    --provider=skeleton \
    --kubeconfig="${KUBECONFIG:-${HOME}/.kube/config}" \
    --report-dir="${ARTIFACTS:-/tmp}"
