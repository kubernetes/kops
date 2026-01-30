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

make test-e2e-install

KUBETEST2_ARGS=()
KUBETEST2_ARGS+=("-v=2")

if [[ "${JOB_TYPE}" == "presubmit" && "${REPO_OWNER}/${REPO_NAME}" == "kubernetes/kops" ]]; then
  KUBETEST2_ARGS+=("--build")
  KUBETEST2_ARGS+=("--kops-binary-path=${GOPATH}/src/k8s.io/kops/.build/dist/linux/$(go env GOARCH)/kops")
else
  KUBETEST2_ARGS+=("--kops-version-marker=${KOPS_VERSION_MARKER:-https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/master/latest-ci.txt}")
fi

kubetest2 kops \
    --up --down \
    "${KUBETEST2_ARGS[@]}" \
    --cloud-provider=aws \
    --create-args="--set=cluster.spec.metricsServer.enabled=true --set=cluster.spec.certManager.enabled=true --master-size=m6g.large --node-size=m6g.large" \
    --kubernetes-version=https://dl.k8s.io/release/stable.txt \
    --test=exec -- "${REPO_ROOT}/tests/e2e/scenarios/metrics-server/test.sh"
