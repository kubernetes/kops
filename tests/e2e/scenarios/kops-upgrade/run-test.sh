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

KUBETEST2="kubetest2 kops -v=2 --cloud-provider=${CLOUD_PROVIDER} --cluster-name=${CLUSTER_NAME:-}"
KUBETEST2="${KUBETEST2} --admin-access=${ADMIN_ACCESS:-}"

export GO111MODULE=on

make test-e2e-install

KOPS="${FIRST_KOPS}"

# Always tear-down the cluster when we're done
function finish {
  ${KUBETEST2} --kops-binary-path="${KOPS}" --down || echo "kubetest2 down failed"
}
trap finish EXIT


${KUBETEST2} \
	--up \
	--kubernetes-version="${K8S_VERSION}" \
	--kops-binary-path="${FIRST_KOPS}" \
	--create-args="--networking calico"

export KOPS_BASE_URL
KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt)"


SECOND_KOPS=$(mktemp -t kops.XXXXXXXXX)
wget -qO "${SECOND_KOPS}" "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
chmod +x "${SECOND_KOPS}"

KOPS="${SECOND_KOPS}"

"${SECOND_KOPS}" update cluster
"${SECOND_KOPS}" update cluster --admin --yes
"${SECOND_KOPS}" update cluster

"${SECOND_KOPS}" rolling-update cluster
"${SECOND_KOPS}" rolling-update cluster --yes --validation-timeout 30m

"${SECOND_KOPS}" validate cluster

cp "${SECOND_KOPS}" "${WORKSPACE}/kops"

${KUBETEST2} \
		--test=kops \
		-- \
		--test-package-version="${K8S_VERSION}" \
		--parallel 25 \
		--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler|TCP.CLOSE_WAIT|Projected.configMap.optional.updates"
