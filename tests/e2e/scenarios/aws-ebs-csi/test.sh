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

# Get the zone from an instance group
ZONE=$(kubectl get nodes -o jsonpath='{.items[0].metadata.labels.topology\.kubernetes\.io/zone}')

REPORT_DIR="${ARTIFACTS:-$(pwd)/_artifacts}/aws-ebs-csi-driver/"
mkdir -p "${REPORT_DIR}"

# Clone the aws-ebs-csi-driver repo at the version deployed in the cluster
CSI_VERSION=$(kubectl get deployment -n kube-system ebs-csi-controller -o jsonpath='{.spec.template.spec.containers[?(@.name=="ebs-plugin")].image}' | cut -d':' -f2-)
TEMPDIR=$(mktemp -dt kops.XXXXXXXXX)
cd "${TEMPDIR}"

go install github.com/onsi/ginkgo/v2/ginkgo@latest

CLONE_ARGS=
if [ -n "$CSI_VERSION" ]; then
    CLONE_ARGS="-b ${CSI_VERSION}"
fi
# shellcheck disable=SC2086
git clone ${CLONE_ARGS} https://github.com/kubernetes-sigs/aws-ebs-csi-driver.git .

cd tests/e2e-kubernetes/

ginkgo --nodes=25 ./... -- -cluster-tag="${CLUSTER_NAME}" -ginkgo.skip="\[Disruptive\]" -report-dir="${REPORT_DIR}" -gce-zone="${ZONE}"
