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
source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

export KOPS_BASE_URL
KOPS_BASE_URL="$(curl -s https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt)"
KOPS=$(kops-download-from-base)

ARGS="--override=cluster.spec.networking.cilium.hubble.enabled=true --override=cluster.spec.certManager.enabled=true"

if [[ $1 == "kube-proxy" ]]; then
    ARGS="${ARGS} --override=cluster.spec.networking.cilium.enableNodePort=false --override=cluster.spec.kubeProxy.enabled=true"
# This test requires private topology, which kubetest2 does not support.
#elif [[ $1 == "eni"]]
#    ARGS="${ARGS} --override=cluster.spec.cilium.ipam=eni --override=cluster.spec.cilium.disable-masquerade"
#    ARGS="${ARGS} --topology private"
elif [[ $1 == "node-local-dns" ]]; then
    ARGS="${ARGS} --override=cluster.spec.kubeDNS.provider=CoreDNS --override=cluster.spec.kubeDNS.nodeLocalDNS.enabled=true"
fi

${KUBETEST2} \
    --up \
    --kubernetes-version="1.21.0" \
    --kops-binary-path="${KOPS}" \
    --create-args="--networking cilium $ARGS"

kubectl port-forward -n kube-system deployment/hubble-relay 4245:4245 &

wget -qO- https://github.com/cilium/cilium-cli/releases/download/v0.7/cilium-linux-amd64.tar.gz | tar xz -C "${WORKSPACE}"

cilium connectivity test --all-flows
