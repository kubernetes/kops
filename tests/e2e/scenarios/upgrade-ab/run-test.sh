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

if [ -z "${KOPS_VERSION_A-}" ] || [ -z "${K8S_VERSION_A-}" ] || [ -z "${KOPS_VERSION_B-}" ] || [ -z "${K8S_VERSION_B-}" ]; then
  >&2 echo "must set all of KOPS_VERSION_A, K8S_VERSION_A, KOPS_VERSION_B, K8S_VERSION_B env vars"
  exit 1
fi

TEST_PACKAGE_VERSION="${K8S_VERSION_B}"

if [[ "$K8S_VERSION_A" == "latest" ]]; then
  K8S_VERSION_A=$(curl https://storage.googleapis.com/kubernetes-release/release/latest.txt)
fi
if [[ "$K8S_VERSION_B" == "latest" ]]; then
  K8S_VERSION_B=$(curl https://storage.googleapis.com/kubernetes-release/release/latest.txt)
  TEST_PACKAGE_MARKER="latest.txt"
fi
if [[ "$K8S_VERSION_A" == "stable" ]]; then
  K8S_VERSION_A=$(curl https://storage.googleapis.com/kubernetes-release/release/stable.txt)
fi
if [[ "$K8S_VERSION_B" == "stable" ]]; then
  K8S_VERSION_B=$(curl https://storage.googleapis.com/kubernetes-release/release/stable.txt)
  TEST_PACKAGE_MARKER="stable.txt"
fi
if [[ "$K8S_VERSION_A" == "ci" ]]; then
  K8S_VERSION_A=https://storage.googleapis.com/k8s-release-dev/ci/$(curl https://storage.googleapis.com/k8s-release-dev/ci/latest.txt)
fi
if [[ "$K8S_VERSION_B" == "ci" ]]; then
  K8S_VERSION_B=https://storage.googleapis.com/k8s-release-dev/ci/$(curl https://storage.googleapis.com/k8s-release-dev/ci/latest.txt)
  TEST_PACKAGE_MARKER="latest.txt"
  TEST_PACKAGE_DIR="ci"
  TEST_PACKAGE_BUCKET="k8s-release-dev"
fi

export KOPS_BASE_URL

echo "Cleaning up any leaked resources from previous cluster"
# For KOPS_VERSION_B, the value "latest" means build of the tree
if [[ "${KOPS_VERSION_B}" == "latest" ]]; then
  kops-acquire-latest
  KOPS_BASE_URL_B="${KOPS_BASE_URL}"
  KOPS_B="${KOPS}"
else
  KOPS_BASE_URL=$(kops-base-from-marker "${KOPS_VERSION_B}")
  KOPS_BASE_URL_B="${KOPS_BASE_URL}"
  KOPS_B=$(kops-download-from-base)
  CHANNELS=$(kops-channels-download-from-base)
fi

${KUBETEST2} \
    --down \
    --kops-binary-path="${KOPS_B}" || echo "kubetest2 down failed"

# First kOps version may be a released version. If so, it is prefixed with v
if [[ "${KOPS_VERSION_A:0:1}" == "v" ]]; then
  KOPS_BASE_URL=""
  KOPS_A=$(kops-download-release "$KOPS_VERSION_A")
  KOPS="${KOPS_A}"
else
  KOPS_BASE_URL=$(kops-base-from-marker "${KOPS_VERSION_A}")
  KOPS_A=$(kops-download-from-base)
  KOPS="${KOPS_A}"
fi



${KUBETEST2} \
    --up \
    --kops-binary-path="${KOPS_A}" \
    --kubernetes-version="${K8S_VERSION_A}" \
    --control-plane-size="${KOPS_CONTROL_PLANE_SIZE:-1}" \
    --template-path="${KOPS_TEMPLATE:-}" \
    --create-args="--networking calico"

# Export kubeconfig-a
KUBECONFIG_A=$(mktemp -t kops.XXXXXXXXX)
"${KOPS_A}" export kubecfg --name "${CLUSTER_NAME}" --admin --kubeconfig "${KUBECONFIG_A}"

# Verify kubeconfig-a
kubectl get nodes -owide --kubeconfig="${KUBECONFIG_A}"

KOPS_BASE_URL="${KOPS_BASE_URL_B}"

KOPS="${KOPS_B}"

if [[ "${KOPS_VERSION_B}" =~ 1.2[01] ]]; then
  "${KOPS_B}" set cluster "${CLUSTER_NAME}" "cluster.spec.kubernetesVersion=${K8S_VERSION_B}"
else
  "${KOPS_B}" edit cluster "${CLUSTER_NAME}" "--set=cluster.spec.kubernetesVersion=${K8S_VERSION_B}"
fi

"${KOPS_B}" update cluster
"${KOPS_B}" update cluster --admin --yes
# Verify no additional changes
"${KOPS_B}" update cluster

# Verify kubeconfig-a still works
kubectl get nodes -owide --kubeconfig "${KUBECONFIG_A}"

# Sleep to ensure channels has done its thing
sleep 120s

${CHANNELS} apply channel "$KOPS_STATE_STORE"/"${CLUSTER_NAME}"/addons/bootstrap-channel.yaml --yes -v4

"${KOPS_B}" rolling-update cluster
"${KOPS_B}" rolling-update cluster --yes --validation-timeout 30m

"${KOPS_B}" validate cluster

# Verify kubeconfig-a still works
kubectl get nodes -owide --kubeconfig="${KUBECONFIG_A}"

cp "${KOPS_B}" "${WORKSPACE}/kops"

"${KOPS_B}" export kubecfg --name "${CLUSTER_NAME}" --admin

if [[ -n ${KOPS_SKIP_E2E:-} ]]; then
  exit
fi


test_package_args="--parallel 25"

if [[ -n ${TEST_PACKAGE_MARKER-} ]]; then
  test_package_args+=" --test-package-marker=${TEST_PACKAGE_MARKER}"
  if [[ -n ${TEST_PACKAGE_DIR-} ]]; then
    test_package_args+=" --test-package-dir=${TEST_PACKAGE_DIR-}"
  fi
  if [[ -n ${TEST_PACKAGE_BUCKET-} ]]; then
    test_package_args+=" --test-package-bucket=${TEST_PACKAGE_BUCKET-}"
  fi
else
  test_package_args+=" --test-package-version=${TEST_PACKAGE_VERSION}"
fi

# shellcheck disable=SC2086
${KUBETEST2} \
    --cloud-provider="${CLOUD_PROVIDER}" \
    --kops-binary-path="${KOPS}" \
    --test=kops \
    -- \
    $test_package_args \
    --parallel 25
