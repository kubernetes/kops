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

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

if [[ -z "${CLOUD_PROVIDER-}" ]]; then
    export CLOUD_PROVIDER="aws"
fi

echo "CLOUD_PROVIDER=${CLOUD_PROVIDER}"
echo "CLUSTER_NAME=${CLUSTER_NAME-}"

if [[ -z "${WORKSPACE-}" ]]; then
    export WORKSPACE
    WORKSPACE=$(mktemp -dt kops.XXXXXXXXX)
fi

if [[ -z "${WORKSPACE-}" ]]; then
    export WORKSPACE
    WORKSPACE=$(mktemp -dt kops.XXXXXXXXX)
fi

if [[ -z "${NETWORKING-}" ]]; then
    export NETWORKING="calico"
fi

export KOPS_BASE_URL
export KOPS
export CHANNELS

export KOPS_FEATURE_FLAGS="SpecOverrideFlag"
export KOPS_RUN_TOO_NEW_VERSION=1

if [[ -z "${DISCOVERY_STORE-}" ]]; then 
    DISCOVERY_STORE="${KOPS_STATE_STORE-}"
fi

export GO111MODULE=on

if [[ -z "${AWS_SSH_PRIVATE_KEY_FILE-}" ]]; then
    export AWS_SSH_PRIVATE_KEY_FILE="${HOME}/.ssh/id_rsa"
fi
if [[ -z "${AWS_SSH_PUBLIC_KEY_FILE-}" ]]; then
    export AWS_SSH_PUBLIC_KEY_FILE="${HOME}/.ssh/id_rsa.pub"
fi

KUBETEST2="kubetest2 kops -v=2 --cloud-provider=${CLOUD_PROVIDER} --cluster-name=${CLUSTER_NAME:-} --kops-root=${REPO_ROOT}"
KUBETEST2="${KUBETEST2} --admin-access=${ADMIN_ACCESS:-} --env=KOPS_FEATURE_FLAGS=${KOPS_FEATURE_FLAGS}"

if [[ -n "${GCP_PROJECT-}" ]]; then
  KUBETEST2="${KUBETEST2} --gcp-project=${GCP_PROJECT}"
fi

# Always tear-down the cluster when we're done
function kops-finish {
    # shellcheck disable=SC2153
    ${KUBETEST2} --kops-binary-path="${KOPS}" --down || echo "kubetest2 down failed"
}
trap kops-finish EXIT

make test-e2e-install

function kops-download-release() {
    local kops
    kops=$(mktemp -t kops.XXXXXXXXX)
    wget -qO "${kops}" "https://github.com/kubernetes/kops/releases/download/${1}/kops-$(go env GOOS)-$(go env GOARCH)"
    chmod +x "${kops}"
    echo "${kops}"
}

function kops-download-from-base() {
    local kops
    kops=$(mktemp -t kops.XXXXXXXXX)
    wget -qO "${kops}" "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/kops"
    chmod +x "${kops}"
    echo "${kops}"
}

function kops-channels-download-from-base() {
    local channels
    channels=$(mktemp -t channels.XXXXXXXXX)
    wget -qO "${channels}" "$KOPS_BASE_URL/$(go env GOOS)/$(go env GOARCH)/channels"
    chmod +x "${channels}"
    echo "${channels}"
}

function kops-base-from-marker() {
    if [[ "${1}" =~ ^https: ]]; then
        curl -s "${1}"
    elif [[ "${1}" == "latest" ]]; then
        curl -s "https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt"
    else
        curl -s "https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/release-${1}/latest-ci.txt"
    fi
}

# This function will download the latest kops if in a periodic job, otherwise build from the current tree
function kops-acquire-latest() {
    if [[ "${JOB_TYPE-}" == "periodic" ]]; then
        KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt)"
        KOPS=$(kops-download-from-base)
        CHANNELS=$(kops-channels-download-from-base)
    else
         if [[ -n "${KOPS_BASE_URL-}" ]]; then
            KOPS_BASE_URL=""
         fi
         $KUBETEST2 --build
         KOPS="${REPO_ROOT}/.build/dist/linux/amd64/kops"
         CHANNELS="${REPO_ROOT}/.build/dist/linux/amd64/channels"
         KOPS_BASE_URL=$(cat "${REPO_ROOT}/.kubetest2/kops-base-url")
         export KOPS_BASE_URL
         echo "KOPS_BASE_URL=$KOPS_BASE_URL"
    fi
}

function kops-up() {
    local create_args
    create_args="--networking ${NETWORKING} ${OVERRIDES-}"
    if [[ -n "${ZONES-}" ]]; then
        create_args="${create_args} --zones=${ZONES}"
    fi
    if [[ -z "${K8S_VERSION-}" ]]; then
        K8S_VERSION="$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"
    fi

    if [[ ${KOPS_IRSA-} = true ]]; then
        create_args="${create_args} --discovery-store=${DISCOVERY_STORE}/${CLUSTER_NAME}/discovery"
    fi

    ${KUBETEST2} \
        --up \
        --kops-binary-path="${KOPS}" \
        --kubernetes-version="${K8S_VERSION}" \
        --create-args="${create_args}" \
        --control-plane-size="${KOPS_CONTROL_PLANE_SIZE:-1}" \
        --template-path="${KOPS_TEMPLATE-}"
}