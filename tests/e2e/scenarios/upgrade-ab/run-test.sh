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

set -o errexit
set -o nounset
set -o pipefail

# Print all commands
set -o xtrace

echo "CLOUD_PROVIDER=${CLOUD_PROVIDER}"

if [ -z "$KOPS_VERSION_A" ] || [ -z "$K8S_VERSION_A" ] || [ -z "$KOPS_VERSION_B" ] || [ -z "$K8S_VERSION_B" ]; then
  >&2 echo "must set all of KOPS_VERSION_A, K8S_VERSION_A, KOPS_VERSION_B, K8S_VERSION_B env vars"
  exit 1
fi

export KOPS_FEATURE_FLAGS="SpecOverrideFlag,${KOPS_FEATURE_FLAGS:-}"
REPO_ROOT=$(git rev-parse --show-toplevel);

WORKDIR=$(mktemp -d)

KOPS_A=${WORKDIR}/kops-${KOPS_VERSION_A}
wget -qO "${KOPS_A}" "https://github.com/kubernetes/kops/releases/download/$KOPS_VERSION_A/kops-$(go env GOOS)-$(go env GOARCH)"
chmod +x "${KOPS_A}"


KUBETEST2="kubetest2 kops -v=2 --cloud-provider=${CLOUD_PROVIDER} --cluster-name=${CLUSTER_NAME:-}"
KUBETEST2="${KUBETEST2} --admin-access=${ADMIN_ACCESS:-}"

export GO111MODULE=on

cd "${REPO_ROOT}/tests/e2e"
go install sigs.k8s.io/kubetest2
go install ./kubetest2-kops
go install ./kubetest2-tester-kops

KOPS_B=${WORKDIR}/kops-${KOPS_VERSION_B}
if [[ "${KOPS_VERSION_B}" == "source" ]]; then
  ${KUBETEST2} --build --kops-root="${REPO_ROOT}" --stage-location="${STAGE_LOCATION:-}" --kops-binary-path="${KOPS_B}"
else
  wget -O "${KOPS_B}" "https://github.com/kubernetes/kops/releases/download/$KOPS_VERSION_B/kops-$(go env GOOS)-$(go env GOARCH)"
  chmod +x "${KOPS_B}"
fi

# Always tear-down the cluster when we're done
function finish {
  ${KUBETEST2} --kops-binary-path="${KOPS_B}" --down || echo "kubetest2 down failed"
}
trap finish EXIT

${KUBETEST2} \
		--up \
		--kops-binary-path="${KOPS_A}" \
		--kubernetes-version="${K8S_VERSION_A}" \
		--create-args="--networking calico"

# Export kubeconfig-a
KUBECONFIG_A="${WORKDIR}/kubeconfig-a"
# TODO: Add --admin if 1.19 or higher...
# Note: --kubeconfig flag not in 1.18
KUBECONFIG="${KUBECONFIG_A}" "${KOPS_A}" export kubecfg --name "${CLUSTER_NAME}"

# Verify kubeconfig-a
KUBECONFIG="${KUBECONFIG_A}" kubectl get nodes -owide

"${KOPS_B}" set cluster "${CLUSTER_NAME}" "cluster.spec.kubernetesVersion=${K8S_VERSION_B}"

"${KOPS_B}" update cluster
"${KOPS_B}" update cluster --admin --yes
# Verify no additional changes
"${KOPS_B}" update cluster

sleep 300
# Verify kubeconfig-a still works
KUBECONFIG="${KUBECONFIG_A}" kubectl get nodes -owide

"${KOPS_B}" rolling-update cluster
"${KOPS_B}" rolling-update cluster --yes --validation-timeout 30m

"${KOPS_B}" validate cluster

# Verify kubeconfig-a still works
KUBECONFIG="${KUBECONFIG_A}" kubectl get nodes -owide

${KUBETEST2} \
		--cloud-provider="${CLOUD_PROVIDER}" \
		--kops-binary-path="${KOPS_B}" \
		--test=kops \
		-- \
		--test-package-version="${K8S_VERSION_B}" \
		--parallel 25 \
		--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler|TCP.CLOSE_WAIT|Projected.configMap.optional.updates"
