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

# This script runs the Leader Migration test suite against an existing replicated
# control plane that is created by kOps.

# Please use kubetest2-tester-exec to invoke this script to properly setup testing
# environment, even though the tester is technically not required.

# Please ensure K8S_VERSION_A and K8S_VERSION_B are set and exported
# so that the test will upgrade the cluster from ${K8S_VERSION_A} to ${K8S_VERSION_B}.

set -xe

REPO_ROOT=$(git rev-parse --show-toplevel)
SCENARIO_ROOT="${REPO_ROOT}/tests/e2e/scenarios/upgrade-ha-leader-migration"
KUBECTL="kubectl -n default" # explicitly set namespace in case the context mess it up

if [[ -z "${KOPS}" ]]; then
	KOPS=$(command -v kops)
fi

echo "KOPS=${KOPS}"

# create the recorder pod
${KUBECTL} create -f "${SCENARIO_ROOT}/resources.yaml"
# build and push the recorder executable
RECORDER=$(mktemp)
(cd "${SCENARIO_ROOT}/cmd/recorder/" && go build -o "${RECORDER}")
${KUBECTL} wait --for=condition=ready pod/recorder
${KUBECTL} cp "${RECORDER}" recorder:/tmp/recorder
rm "${RECORDER}"
${KUBECTL} exec recorder -- /usr/bin/env sh -c 'mv /tmp/recorder /usr/local/bin/recorder'

# prepare for the upgrade
# workaround current state of node IPAM controller
${KOPS} edit cluster \
	'--set=cluster.spec.kubeControllerManager.enableLeaderMigration=false' \
	'--set=cluster.spec.cloudControllerManager.enableLeaderMigration=true' \
	'--set=cluster.spec.cloudControllerManager.controllers=*' \
	'--set=cluster.spec.cloudControllerManager.controllers=-nodeipam' \
	"--set=cluster.spec.kubernetesVersion=${K8S_VERSION_B}"

# perform the upgrade
${KOPS} update cluster
${KOPS} update cluster --admin --yes
${KOPS} update cluster

# perform the rolling upgrade, we only care about the control plane
${KOPS} rolling-update cluster
${KOPS} rolling-update cluster --yes --validation-timeout=30m --instance-group-roles=master

# check recorder status
phase=$(${KUBECTL} get pod -o go-template="{{.status.phase}}" recorder)

# if the recorder fails, which means a conflict is detected, dump log and exit
if [[ ! "$phase" == Running ]]; then
	${KUBECTL} logs recorder
	${KUBECTL} delete -f "${SCENARIO_ROOT}/resources.yaml" # clean up
	echo "upgrade failed"
	exit 1
fi

# prepare for the rollback
${KOPS} edit cluster \
	'--set=cluster.spec.kubeControllerManager.enableLeaderMigration=true' \
	'--unset=cluster.spec.cloudControllerManager' \
	"--set=cluster.spec.kubernetesVersion=${K8S_VERSION_A}"

# perform the rollback
${KOPS} update cluster
${KOPS} update cluster --admin --yes
${KOPS} update cluster

# perform the rolling rollback of the control plane
${KOPS} rolling-update cluster
${KOPS} rolling-update cluster --yes --validation-timeout=30m --instance-group-roles=master

# dump recorder output
${KUBECTL} logs recorder

# check recorder status, again
phase=$(${KUBECTL} get pod -o go-template="{{.status.phase}}" recorder)

# clean up
${KUBECTL} delete -f "${SCENARIO_ROOT}/resources.yaml"

if [[ ! "$phase" == Running ]]; then
	echo "rollback failed"
	exit 1
fi
