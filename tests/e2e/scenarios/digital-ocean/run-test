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
set -o xtrace

export KOPS_FEATURE_FLAGS="SpecOverrideFlag,${KOPS_FEATURE_FLAGS:-}"
REPO_ROOT=$(git rev-parse --show-toplevel);
PATH=$REPO_ROOT/.bazel-bin/cmd/kops/$(go env GOOS)-$(go env GOARCH):$PATH

KUBETEST2_COMMON_ARGS="-v=2 --cloud-provider=digitalocean --cluster-name=e2e-test-do.k8s.local --kops-binary-path=${REPO_ROOT}/.bazel-bin/cmd/kops/linux-amd64/kops"
KUBETEST2_COMMON_ARGS="${KUBETEST2_COMMON_ARGS} --admin-access=${ADMIN_ACCESS:-}"

export GO111MODULE=on
go get sigs.k8s.io/kubetest2/...@latest

cd ${REPO_ROOT}/tests/e2e
go install ./kubetest2-kops
go install ./kubetest2-tester-kops

kubetest2 kops ${KUBETEST2_COMMON_ARGS} --build --kops-root=${REPO_ROOT} --stage-location=${STAGE_LOCATION:-}

kubetest2 kops ${KUBETEST2_COMMON_ARGS} \
		-v 6 \
		--up --down \
		--env S3_ENDPOINT=sfo3.digitaloceanspaces.com \
		--env JOB_NAME=pull-kops-e2e-kubernetes-do-kubetest2 \
		--create-args "--networking=cilium --api-loadbalancer-type=public --node-count=2 --master-count=3" \
		--kops-version-marker=https://storage.googleapis.com/k8s-staging-kops/kops/releases/markers/release-1.20/latest-ci.txt \
		--kubernetes-version=https://storage.googleapis.com/kubernetes-release/release/stable-1.20.txt \
		--test=kops \
		-- \
		--ginkgo-args="--debug" \
		--test-package-marker=stable-1.20.txt \
		--parallel 25 \
		--skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler|Services.*functioning.*NodePort|Services.*rejected.*endpoints|Services.*NodePort.*listening.*same.*port|TCP.CLOSE_WAIT|should.*run.*through.*the.*lifecycle.*of.*Pods.*and.*PodStatus"
