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
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

if [ -z "${KOPS_VERSION:-}" ] || [ -z "${K8S_VERSION:-}" ]; then
  >&2 echo "must set KOPS_VERSION and K8S_VERSION env vars"
  exit 1
fi

KOPS=$(kops-download-release "${KOPS_VERSION}")

${KUBETEST2} \
	--up \
	--kubernetes-version="${K8S_VERSION}" \
	--kops-binary-path="${KOPS}" \
	--create-args="--networking calico"


"${KOPS}" validate cluster

#"${KOPS}" export kubecfg --name "${CLUSTER_NAME}" --admin

if [[ -n ${KOPS_SKIP_E2E:-} ]]; then
  exit
fi

# shellcheck disable=SC2086
${KUBETEST2} \
    --cloud-provider="${CLOUD_PROVIDER}" \
    --kops-binary-path="${KOPS}" \
    --test=kops \
    -- \
    --parallel 25 \
	--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler"
