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

kops-acquire-latest

OVERRIDES="${OVERRIDES} --master-size=t4g.medium --node-size=t4g.medium"

kops-up

REPORT_DIR="${ARTIFACTS:-$(pwd)/_artifacts}/keypair-rotation"
mkdir -p "${REPORT_DIR}"

${KOPS} create keypair all
${KOPS} update cluster --yes
${KOPS} rolling-update cluster --yes --validate-count=10

KUBECFG_CREATE=$(mktemp -t kubeconfig.XXXXXXXXX)
${KOPS} export kubecfg --admin --kubeconfig="${KUBECFG_CREATE}"
kubectl --kubeconfig="${KUBECFG_CREATE}" config view > "${REPORT_DIR}/create.kubeconfig"

# Confirm the first kubeconfig still works
${KOPS} validate cluster --wait=10m --count=3

export KUBECONFIG="${KUBECFG_CREATE}"
${KOPS} promote keypair all
${KOPS} update cluster --yes
${KOPS} rolling-update cluster --yes --validate-count=10

KUBECFG_PROMOTE=$(mktemp -t kubeconfig.XXXXXXXXX)
${KOPS} export kubecfg --admin --kubeconfig="${KUBECFG_PROMOTE}"
kubectl --kubeconfig="${KUBECFG_PROMOTE}" config view > "${REPORT_DIR}/promote.kubeconfig"

export KUBECONFIG="${KUBECFG_PROMOTE}"
${KOPS} validate cluster --wait=10m --count=3

${KOPS} distrust keypair all
${KOPS} update cluster --yes
${KOPS} rolling-update cluster --yes --validate-count=10

KUBECFG_DISTRUST=$(mktemp -t kubeconfig.XXXXXXXXX)
${KOPS} export kubecfg --admin --kubeconfig="${KUBECFG_DISTRUST}"
kubectl --kubeconfig="${KUBECFG_DISTRUST}" config view > "${REPORT_DIR}/distrust.kubeconfig"

CA=$(kubectl --kubeconfig="${KUBECFG_DISTRUST}" config view --raw -o jsonpath="{.clusters[0].cluster.certificate-authority-data}" | base64 --decode)
if [ "$(echo "${CA}" | grep -c "BEGIN CERTIFICATE")" != "1" ]; then
    >&2 echo unexpected number of CA certificates in kubeconfig
    exit 1
fi

export KUBECONFIG="${KUBECFG_DISTRUST}"
${KOPS} validate cluster --wait=10m --count=3
