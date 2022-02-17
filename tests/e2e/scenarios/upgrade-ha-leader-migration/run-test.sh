#!/usr/bin/env bash

# Copyright 2022 The Kubernetes Authors.
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

REPO_ROOT=$(git rev-parse --show-toplevel)
SCENARIO_ROOT="${REPO_ROOT}/tests/e2e/scenarios/upgrade-ha-leader-migration"
KOPS_FEATURE_FLAGS=SpecOverrideFlag
export KOPS_FEATURE_FLAGS
(cd "$REPO_ROOT" && make test-e2e-install)

KUBETEST2="kubetest2 kops -v=2 --cloud-provider=${CLOUD_PROVIDER} --kops-root=${REPO_ROOT}"

if [[ -n "${GCP_PROJECT-}" ]]; then
	KUBETEST2="${KUBETEST2} --gcp-project=${GCP_PROJECT}"
fi

if [[ -n "${CLUSTER_NAME-}" ]]; then
	KUBETEST2="${KUBETEST2} --cluster-name=${CLUSTER_NAME}"
fi

# for periodic job, let kubetest2 fetch latest kops. Otherwise let kubetest2 build.
if [[ "${JOB_TYPE-}" == "periodic" ]]; then
	KUBETEST2="${KUBETEST2} --kops-version-marker=https://storage.googleapis.com/kops-ci/bin/latest-ci-updown-green.txt"
else
	KUBETEST2="${KUBETEST2} --build"
fi

OVERRIDES=("--channel=alpha" "--node-count=1" "--master-count=3")

case "${CLOUD_PROVIDER}" in
gce)
	export KOPS_FEATURE_FLAGS="${KOPS_FEATURE_FLAGS},AlphaAllowGCE"
	OVERRIDES+=(
		"--zones=us-central1-a,us-central1-b,us-central1-c"
		"--master-zones=us-central1-a,us-central1-b,us-central1-c"
		"--gce-service-account=default" # see test-infra#24749
	)
	;;
*) ;;

esac

# the migration in this test case is KCM to KCM+CCM, which should happen
# during the upgrade from 1.23 to 1.24
K8S_VERSION_A=$(curl https://storage.googleapis.com/kubernetes-release/release/stable-1.23.txt)
export K8S_VERSION_A
K8S_VERSION_B=$(curl https://storage.googleapis.com/kubernetes-release/release/latest-1.24.txt)
export K8S_VERSION_B

# install kubetest2-test-exec if needed
if ! command -v kubetest2-tester-exec >/dev/null; then
	go install sigs.k8s.io/kubetest2/kubetest2-tester-exec@latest
fi

# run the test with kubetest2
${KUBETEST2} \
	--up \
	--test=exec \
	--down \
	--kubernetes-version="${K8S_VERSION_A}" \
	--create-args="${OVERRIDES[*]}" \
	-- \
	"${SCENARIO_ROOT}/test-leader-migration.sh"
