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

# Checks the properties of a split control plane that the integration tests can only assert on
# generated terraform: that the components actually came up where they were placed.

set -o errexit
set -o nounset
set -o pipefail

kubectl get nodes -o wide

fail=0

function expect_label() {
  local ig="$1" label="$2"
  local node
  node=$(kubectl get nodes -l "kops.k8s.io/instancegroup=${ig}" -o name | head -1)
  if [[ -z "${node}" ]]; then
    echo "FAIL: no node found for instance group ${ig}"
    fail=1
    return
  fi
  if kubectl get "${node}" -o jsonpath="{.metadata.labels['${label//./\\.}']}" | grep -q . ; then
    echo "ok: ${ig} has ${label}"
  else
    echo "FAIL: ${ig} is missing ${label}"
    fail=1
  fi
}

function expect_no_label() {
  local ig="$1" label="$2"
  local node
  node=$(kubectl get nodes -l "kops.k8s.io/instancegroup=${ig}" -o name | head -1)
  if [[ -z "${node}" ]]; then
    echo "FAIL: no node found for instance group ${ig}"
    fail=1
    return
  fi
  if kubectl get "${node}" -o jsonpath="{.metadata.labels['${label//./\\.}']}" | grep -q . ; then
    echo "FAIL: ${ig} unexpectedly has ${label}"
    fail=1
  else
    echo "ok: ${ig} does not have ${label}"
  fi
}

ZONE=$(kubectl get nodes -o jsonpath='{.items[0].metadata.labels.topology\.kubernetes\.io/zone}')

# Every role label should land on its own instance group.
expect_label "external-${ZONE}" "node-role.kubernetes.io/api-server"
expect_label "internal-${ZONE}" "node-role.kubernetes.io/api-server"
expect_label "etcd-${ZONE}" "node-role.kubernetes.io/etcd"
expect_label "scheduler-${ZONE}" "node-role.kubernetes.io/scheduler"
expect_label "kcm-${ZONE}" "node-role.kubernetes.io/kube-controller-manager"

# The API servers these two run are reached on localhost only, so they carry no API server
# network tag and so no API server label. That is the same decision that keeps them off the
# public firewall rule.
expect_no_label "scheduler-${ZONE}" "node-role.kubernetes.io/api-server"
expect_no_label "kcm-${ZONE}" "node-role.kubernetes.io/api-server"

# Only the internal API server claims the scheduled components.
expect_label "internal-${ZONE}" "node-role.kops.k8s.io/kops-controller"
expect_label "internal-${ZONE}" "node-role.kops.k8s.io/cloud-controller-manager"
for ig in "external" "scheduler" "kcm" "etcd"; do
  expect_no_label "${ig}-${ZONE}" "node-role.kops.k8s.io/kops-controller"
  expect_no_label "${ig}-${ZONE}" "node-role.kops.k8s.io/cloud-controller-manager"
done

# kops-controller and the cloud-controller-manager should have been scheduled onto the node that
# claimed them, which is the point of the placement.
for app in kops-controller cloud-controller-manager; do
  pods=$(kubectl get pods -n kube-system -l k8s-app="${app}" -o jsonpath='{.items[*].spec.nodeName}' 2>/dev/null || true)
  if [[ -z "${pods}" ]]; then
    echo "FAIL: no ${app} pods found"
    fail=1
    continue
  fi
  for node in ${pods}; do
    ig=$(kubectl get "node/${node}" -o jsonpath='{.metadata.labels.kops\.k8s\.io/instancegroup}')
    if [[ "${ig}" == "internal-${ZONE}" ]]; then
      echo "ok: ${app} is running on ${ig}"
    else
      echo "FAIL: ${app} is running on ${ig}, expected internal-${ZONE}"
      fail=1
    fi
  done
done

# Each split component should be running, as a static pod, only where its role put it.
for ig_component in "scheduler:kube-scheduler" "kcm:kube-controller-manager"; do
  ig="${ig_component%%:*}"
  component="${ig_component##*:}"
  pods=$(kubectl get pods -n kube-system -l k8s-app="${component}" -o jsonpath='{.items[*].spec.nodeName}' 2>/dev/null || true)
  if [[ -z "${pods}" ]]; then
    echo "FAIL: no ${component} pods found"
    fail=1
    continue
  fi
  for node in ${pods}; do
    actual=$(kubectl get "node/${node}" -o jsonpath='{.metadata.labels.kops\.k8s\.io/instancegroup}')
    if [[ "${actual}" == "${ig}-${ZONE}" ]]; then
      echo "ok: ${component} is running on ${actual}"
    else
      echo "FAIL: ${component} is running on ${actual}, expected ${ig}-${ZONE}"
      fail=1
    fi
  done
done

kubectl get pods -n kube-system -o wide

if [[ "${fail}" != "0" ]]; then
  echo "split control plane checks failed"
  exit 1
fi
echo "split control plane checks passed"
