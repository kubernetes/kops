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

echo "CLOUD_PROVIDER=${CLOUD_PROVIDER}"

REPORT_DIR="${ARTIFACTS:-$(pwd)/_artifacts}/aws-ebs-csi-driver/"

export KOPS_FEATURE_FLAGS="SpecOverrideFlag,${KOPS_FEATURE_FLAGS:-}"
REPO_ROOT=$(git rev-parse --show-toplevel);

KOPS="${REPO_ROOT}/bazel-bin/cmd/kops/linux-amd64/kops"

KUBETEST2="kubetest2 kops -v=2 --cloud-provider=${CLOUD_PROVIDER} --cluster-name=${CLUSTER_NAME:-}"
KUBETEST2="${KUBETEST2} --admin-access=${ADMIN_ACCESS:-}"

export GO111MODULE=on

cd "${REPO_ROOT}/tests/e2e"
go install sigs.k8s.io/kubetest2
go install ./kubetest2-kops
go install ./kubetest2-tester-kops

${KUBETEST2} --build --kops-root="${REPO_ROOT}" --stage-location="${STAGE_LOCATION:-}" --kops-binary-path="${KOPS}"

# Always tear-down the cluster when we're done
function finish {
  ${KUBETEST2} --kops-binary-path="${KOPS}" --down || echo "kubetest2 down failed"
}
trap finish EXIT

${KUBETEST2} \
		--up \
		--kops-binary-path="${KOPS}" \
		--kubernetes-version="1.21.0" \
		--create-args="--networking calico --override=cluster.spec.cloudConfig.awsEBSCSIDriver.enabled=true"


ZONE=$(${KOPS} get ig -o json | jq -r '[.[] | select(.spec.role=="Node") | .spec.subnets[0]][0]')

cd "$(mktemp -dt kops.XXXXXXXXX)"
go get github.com/onsi/ginkgo/ginkgo

git clone --branch v1.0.0 https://github.com/kubernetes-sigs/aws-ebs-csi-driver.git .
cd tests/e2e-kubernetes/

ginkgo --nodes=25 ./... -- -ginkgo.skip="\[Disruptive\]" -report-dir="${REPORT_DIR}" -gce-zone="${ZONE}"
