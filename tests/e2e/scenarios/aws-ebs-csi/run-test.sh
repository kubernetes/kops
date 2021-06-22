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

OVERRIDES="${OVERRIDES-} --override=cluster.spec.cloudConfig.awsEBSCSIDriver.enabled=true"
OVERRIDES="$OVERRIDES --override=cluster.spec.snapshotController.enabled=true"
OVERRIDES="$OVERRIDES --override=cluster.spec.certManager.enabled=true"

kops-up

ZONE=$(${KOPS} get ig -o json | jq -r '[.[] | select(.spec.role=="Node") | .spec.subnets[0]][0]')
REPORT_DIR="${ARTIFACTS:-$(pwd)/_artifacts}/aws-ebs-csi-driver/"

# shellcheck disable=SC2164
cd "$(mktemp -dt kops.XXXXXXXXX)"
go get github.com/onsi/ginkgo/ginkgo

git clone --branch v1.1.0 https://github.com/kubernetes-sigs/aws-ebs-csi-driver.git .

# shellcheck disable=SC2164
cd tests/e2e-kubernetes/

# Skipping disruptive and flakes caused by https://github.com/kubernetes-sigs/aws-ebs-csi-driver/issues/911
ginkgo --nodes=25 ./... -- -cluster-tag="${CLUSTER_NAME}" -ginkgo.skip="\[Disruptive\]|should.check.snapshot.fields" -report-dir="${REPORT_DIR}" -gce-zone="${ZONE}"
