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
set -o xtrace

echo "CLOUD_PROVIDER=${CLOUD_PROVIDER}"

FIRST_VERSION=$1
K8S_VERSION=$2

if [ -z "$FIRST_VERSION" ] || [ -z "$K8S_VERSION" ]; then
  >&2 echo "Usage: '$0 <first-kops-version> <k8s-version>'"
  exit 1
fi

FIRST_KOPS=$(mktemp -t kops.XXXXXXXXX)
wget -qO "${FIRST_KOPS}" "https://github.com/kubernetes/kops/releases/download/$FIRST_VERSION/kops-$(go env GOOS)-$(go env GOARCH)"
chmod +x "${FIRST_KOPS}"

export KOPS_FEATURE_FLAGS="SpecOverrideFlag,${KOPS_FEATURE_FLAGS:-}"
REPO_ROOT=$(git rev-parse --show-toplevel);

SECOND_KOPS="${REPO_ROOT}/bazel-bin/cmd/kops/linux-amd64/kops"

KUBETEST2="kubetest2 kops -v=2 --cloud-provider=${CLOUD_PROVIDER} --cluster-name=${CLUSTER_NAME:-}"
KUBETEST2="${KUBETEST2} --admin-access=${ADMIN_ACCESS:-}"

export GO111MODULE=on

cd "${REPO_ROOT}/tests/e2e"
go install sigs.k8s.io/kubetest2
go install ./kubetest2-kops
go install ./kubetest2-tester-kops

${KUBETEST2} --build --kops-root="${REPO_ROOT}" --stage-location="${STAGE_LOCATION:-}" --kops-binary-path="${SECOND_KOPS}"

# Always tear-down the cluster when we're done
function finish {
  ${KUBETEST2} --kops-binary-path="${SECOND_KOPS}" --down || echo "kubetest2 down failed"
}
trap finish EXIT

${KUBETEST2} \
		--up \
		--kops-binary-path="${FIRST_KOPS}" \
		--kubernetes-version="${K8S_VERSION}" \
		--create-args="--networking calico"

"${SECOND_KOPS}" update cluster
"${SECOND_KOPS}" update cluster --admin --yes
"${SECOND_KOPS}" update cluster

"${SECOND_KOPS}" rolling-update cluster
"${SECOND_KOPS}" rolling-update cluster --yes --validation-timeout 30m

"${SECOND_KOPS}" validate cluster

${KUBETEST2} \
		--cloud-provider="${CLOUD_PROVIDER}" \
		--kops-binary-path="${SECOND_KOPS}" \
		--test=kops \
		-- \
		--test-package-version="${K8S_VERSION}" \
		--parallel 25 \
		--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler|TCP.CLOSE_WAIT|Projected.configMap.optional.updates"
