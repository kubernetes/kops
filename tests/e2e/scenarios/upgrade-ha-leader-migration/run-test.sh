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

REPO_ROOT=$(git rev-parse --show-toplevel)
SCENARIO_ROOT="${REPO_ROOT}/tests/e2e/scenarios/upgrade-ha-leader-migration"

source "${REPO_ROOT}"/tests/e2e/scenarios/lib/common.sh

OVERRIDES=("--node-count=1" "--master-count=3")

case "${CLOUD_PROVIDER}" in
gce)
  export KOPS_FEATURE_FLAGS=AlphaAllowGCE,SpecOverrideFlag
  OVERRIDES+=("--zones=us-central1-a,us-central1-b,us-central1-c" "--master-zones=us-central1-a,us-central1-b,us-central1-c")
  ;;
*) ;;

esac

kops-acquire-latest

# the migration in this test case is KCM to KCM+CCM, which should happen
# during the upgrade from 1.23 to 1.24
K8S_VERSION_A=$(curl https://storage.googleapis.com/kubernetes-release/release/stable-1.23.txt)
K8S_VERSION_B=$(curl https://storage.googleapis.com/kubernetes-release/release/latest-1.24.txt)

# spin up the 1.23 cluster
OVERRIDES="${OVERRIDES[*]}" K8S_VERSION="${K8S_VERSION_A}" kops-up

# create the recorder pod
kubectl create -f "${SCENARIO_ROOT}/resources.yaml"
(cd "${SCENARIO_ROOT}/cmd/recorder/" && go build -o "${WORKSPACE}/recorder")
kubectl wait --for=condition=ready pod/recorder
kubectl cp "${WORKSPACE}/recorder" recorder:/tmp/recorder
rm "${WORKSPACE}/recorder"
kubectl exec recorder -- /usr/bin/env sh -c 'mv /tmp/recorder /usr/local/bin/recorder'

# prepare for the upgrade
# workaround current state of node IPAM controller
"${KOPS}" edit cluster \
  '--set=cluster.spec.kubeControllerManager.enableLeaderMigration=false' \
  '--set=cluster.spec.cloudControllerManager.enableLeaderMigration=true' \
  '--set=cluster.spec.cloudControllerManager.controllers=*' \
  '--set=cluster.spec.cloudControllerManager.controllers=-nodeipam' \
  "--set=cluster.spec.kubernetesVersion=${K8S_VERSION_B}"

# perform the upgrade
"${KOPS}" update cluster
"${KOPS}" update cluster --admin --yes
"${KOPS}" update cluster

# perform the rolling upgrade, we only care about the control plane
"${KOPS}" rolling-update cluster
"${KOPS}" rolling-update cluster --yes --validation-timeout=30m --instance-group-roles=master

# check recorder status
phase=$(kubectl get pod -o go-template="{{.status.phase}}" recorder)

# if the recorder fails, which means a conflict is detected, dump log and exit
if [[ ! "$phase" == Running ]]; then
    kubectl logs recorder
    kubectl delete -f "${SCENARIO_ROOT}/resources.yaml" # clean up
    echo "upgrade failed"
    exit 1
fi

# prepare for the rollback
"${KOPS}" edit cluster \
  '--set=cluster.spec.kubeControllerManager.enableLeaderMigration=true' \
  '--unset=cluster.spec.cloudControllerManager' \
  "--set=cluster.spec.kubernetesVersion=${K8S_VERSION_A}"

# perform the rollback
"${KOPS}" update cluster
"${KOPS}" update cluster --admin --yes
"${KOPS}" update cluster

# perform the rolling rollback of the control plane
"${KOPS}" rolling-update cluster
"${KOPS}" rolling-update cluster --yes --validation-timeout=30m --instance-group-roles=master

# dump recorder output
kubectl logs recorder

# check recorder status, again
phase=$(kubectl get pod -o go-template="{{.status.phase}}" recorder)

# clean up
kubectl delete -f "${SCENARIO_ROOT}/resources.yaml"

if [[ ! "$phase" == Running ]]; then
  echo "rollback failed"
  exit 1
fi
