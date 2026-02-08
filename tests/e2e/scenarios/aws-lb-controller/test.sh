#!/usr/bin/env bash

# Copyright 2026 The Kubernetes Authors.
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

# Get the cluster name from the KUBECONFIG context
CLUSTER_NAME=$(kubectl config view --minify -o jsonpath='{.clusters[0].name}')

# Get VPC ID by looking at the cluster's subnet
VPC=$(kubectl get nodes -o jsonpath='{.items[0].spec.providerID}' | cut -d'/' -f5 | xargs -I{} aws ec2 describe-instances --instance-ids {} --query 'Reservations[0].Instances[0].VpcId' --output text)

# Get the zone from a node
ZONE=$(kubectl get nodes -o jsonpath='{.items[0].metadata.labels.topology\.kubernetes\.io/zone}')
REGION=${ZONE%?}

REPORT_DIR="${ARTIFACTS:-$(pwd)/_artifacts}/aws-lb-controller"
mkdir -p "${REPORT_DIR}"

# Clone the aws-load-balancer-controller repo at the version deployed in the cluster
LBC_VERSION=$(kubectl get deployment -n kube-system aws-load-balancer-controller -o jsonpath='{.spec.template.spec.containers[?(@.name=="controller")].image}' | cut -d':' -f2-)
TEMPDIR=$(mktemp -dt kops.XXXXXXXXX)
cd "${TEMPDIR}"

go install github.com/onsi/ginkgo/v2/ginkgo@latest

CLONE_ARGS=
if [ -n "$LBC_VERSION" ]; then
    CLONE_ARGS="-b ${LBC_VERSION}"
fi
# shellcheck disable=SC2086
git clone ${CLONE_ARGS} https://github.com/kubernetes-sigs/aws-load-balancer-controller .

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
