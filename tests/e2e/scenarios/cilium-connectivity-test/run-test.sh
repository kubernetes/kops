#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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
TEST_ROOT="${REPO_ROOT}/tests/e2e/scenarios/cilium-connectivity-test"

make test-e2e-install

KUBETEST2_ARGS=()
KUBETEST2_ARGS+=("-v=2")

if [[ "${JOB_TYPE}" == "presubmit" && "${REPO_OWNER}/${REPO_NAME}" == "kubernetes/kops" ]]; then
  KUBETEST2_ARGS+=("--build")
  KUBETEST2_ARGS+=("--kops-binary-path=${GOPATH}/src/k8s.io/kops/.build/dist/linux/$(go env GOARCH)/kops")
else
  KUBETEST2_ARGS+=("--kops-version-marker=${KOPS_VERSION_MARKER:-https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci.txt}")
fi

CREATE_ARGS="--networking cilium --set=cluster.spec.networking.cilium.hubble.enabled=true --set=cluster.spec.certManager.enabled=true"

if [[ "${1:-}" == "kube-proxy" ]]; then
    CREATE_ARGS="${CREATE_ARGS} --set=cluster.spec.networking.cilium.enableNodePort=false --set=cluster.spec.kubeProxy.enabled=true"
# This test requires private topology, which kubetest2 does not support.
#elif [[ "${1:-}" == "eni"]]
#    CREATE_ARGS="${CREATE_ARGS} --set=cluster.spec.cilium.ipam=eni --set=cluster.spec.cilium.disable-masquerade"
#    CREATE_ARGS="${CREATE_ARGS} --topology private"
elif [[ "${1:-}" == "node-local-dns" ]]; then
    CREATE_ARGS="${CREATE_ARGS} --set=cluster.spec.kubeDNS.provider=CoreDNS --set=cluster.spec.kubeDNS.nodeLocalDNS.enabled=true"
fi

kubetest2 kops \
    --up --down \
    "${KUBETEST2_ARGS[@]}" \
    --cloud-provider=aws \
    --create-args="${CREATE_ARGS}" \
    --kubernetes-version="https://dl.k8s.io/release/stable.txt" \
    --test=exec -- "${TEST_ROOT}/test.sh"
