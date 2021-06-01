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

KUBETEST2="kubetest2 kops -v=2 --cloud-provider=${CLOUD_PROVIDER} --cluster-name=${CLUSTER_NAME:-}"
KUBETEST2="${KUBETEST2} --admin-access=${ADMIN_ACCESS:-}"

export GO111MODULE=on

cd "${REPO_ROOT}/tests/e2e"
go install sigs.k8s.io/kubetest2
go install ./kubetest2-kops
go install ./kubetest2-tester-kops

# Always tear-down the cluster when we're done
function finish {
  ${KUBETEST2} --down || echo "kubetest2 down failed"
}
trap finish EXIT


export KUBE_TEST_REPO_LIST="${REPO_ROOT}/tests/e2e/scenarios/aws-ebs-csi/repos.yaml"

${KUBETEST2} \
	--up --down \
	--cloud-provider=aws \
	--create-args="--node-size=m6g.large --master-size=m6g.large --image='099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-arm64-server-20210510' --channel=alpha --networking=calico --container-runtime=containerd" \
	--kops-version-marker=https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt \
	--kubernetes-version=https://storage.googleapis.com/kubernetes-release/release/stable-1.21.txt \
	--test=kops \
	-- \
	--ginkgo-args="--debug" \
	--test-args="-test.timeout=60m -num-nodes=0" \
	--test-package-marker=stable-1.21.txt \
	--parallel=25 \
	--focus-regex="In-tree.Volumes.*nfs" \
	--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler"
