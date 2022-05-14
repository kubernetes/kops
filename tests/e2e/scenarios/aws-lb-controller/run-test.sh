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

OVERRIDES="${OVERRIDES-} --set=cluster.spec.awsLoadBalancerController.enabled=true"
OVERRIDES="${OVERRIDES} --set=cluster.spec.certManager.enabled=true"

# shellcheck disable=SC2034
ZONES="eu-west-1a,eu-west-1b,eu-west-1c"

kops-up

VPC=$(${KOPS} toolbox dump -o json | jq -r .vpc.id)

ZONE=$(${KOPS} get ig -o json | jq -r '[.[] | select(.spec.role=="Node") | .spec.subnets[0]][0]')

REGION=${ZONE%?}

REPORT_DIR="${ARTIFACTS:-$(pwd)/_artifacts}/aws-lb-controller"

# shellcheck disable=SC2164
cd "$(mktemp -dt kops.XXXXXXXXX)"
go install github.com/onsi/ginkgo/ginkgo@latest

git clone https://github.com/kubernetes-sigs/aws-load-balancer-controller .

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
