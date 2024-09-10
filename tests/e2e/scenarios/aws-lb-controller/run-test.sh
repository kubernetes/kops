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


# shellcheck disable=SC2034
NETWORKING="amazonvpc"

OVERRIDES="${OVERRIDES-} --set=cluster.spec.cloudProvider.aws.loadBalancerController.enabled=true"
OVERRIDES="${OVERRIDES} --set=cluster.spec.certManager.enabled=true"
OVERRIDES="${OVERRIDES} --master-size=t3.medium --node-size=t3.medium" # Use amd64 because LBC's E2E suite uses single-arch amd64 test images
OVERRIDES="${OVERRIDES} --image=${INSTANCE_IMAGE:-099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-20240906}"

# shellcheck disable=SC2034
ZONES="eu-west-1a,eu-west-1b,eu-west-1c"

kops-up

KUBECONFIG=$(mktemp -t kops.XXXXXXXXX)
export KUBECONFIG
"${KOPS}" export kubecfg --name "${CLUSTER_NAME}" --admin --kubeconfig "${KUBECONFIG}"

VPC=$(${KOPS} toolbox dump -o json | jq -r .vpc.id)

ZONE=$(${KOPS} get ig -o json | jq -r '[.[] | select(.spec.role=="Node") | .spec.subnets[0]][0]')

REGION=${ZONE%?}

REPORT_DIR="${ARTIFACTS:-$(pwd)/_artifacts}/aws-lb-controller"

# shellcheck disable=SC2164
cd "$(mktemp -dt kops.XXXXXXXXX)"
go install github.com/onsi/ginkgo/ginkgo@latest

LBC_VERSION=$(kubectl get deployment -n kube-system aws-load-balancer-controller -o jsonpath='{.spec.template.spec.containers[?(@.name=="controller")].image}' | cut -d':' -f2-)
CLONE_ARGS=
if [ -n "$LBC_VERSION" ]; then
    CLONE_ARGS="-b ${LBC_VERSION}"
fi
# shellcheck disable=SC2086
git clone ${CLONE_ARGS} https://github.com/kubernetes-sigs/aws-load-balancer-controller .

mkdir -p "${REPORT_DIR}"

ginkgo -v -r test/e2e/ingress -- \
    -cluster-name="${CLUSTER_NAME}" \
    -aws-region="${REGION}" \
    -aws-vpc-id="$VPC" \
    -ginkgo.junit-report="${REPORT_DIR}/junit-ingress.xml"

ginkgo -v -r test/e2e/service -- \
    -cluster-name="${CLUSTER_NAME}" \
    -aws-region="${REGION}" \
    -aws-vpc-id="$VPC" \
    -ginkgo.junit-report="${REPORT_DIR}/junit-service.xml"
