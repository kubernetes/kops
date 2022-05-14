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
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

FIRST_VERSION=$1
K8S_VERSION=$2

if [ -z "$FIRST_VERSION" ] || [ -z "$K8S_VERSION" ]; then
  >&2 echo "Usage: '$0 <first-kops-version> <k8s-version>'"
  exit 1
fi

KOPS=$(kops-download-release "${FIRST_VERSION}")

${KUBETEST2} \
	--up \
	--kubernetes-version="${K8S_VERSION}" \
	--kops-binary-path="${KOPS}" \
	--create-args="--networking calico"

export KOPS_BASE_URL
KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt)"
KOPS=$(kops-download-from-base)

"${KOPS}" update cluster
"${KOPS}" update cluster --admin --yes
"${KOPS}" update cluster

"${KOPS}" rolling-update cluster
"${KOPS}" rolling-update cluster --yes --validation-timeout 30m

"${KOPS}" validate cluster

cp "${KOPS}" "${WORKSPACE}/kops"

${KUBETEST2} \
		--test=kops \
		-- \
		--test-package-version="${K8S_VERSION}" \
		--parallel 25 \
		--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler|nfs|NFS|TCP.CLOSE_WAIT|Projected.configMap.optional.updates"
