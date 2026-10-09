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

set -x

# Capture kubelet heap profiles from a sample of nodes so a kubelet memory
# change seen on perf-dash can be diagnosed from the job artifacts alone.
# kubetest2 runs this after the test and before teardown with KUBECONFIG and
# ARTIFACTS set. Everything here is best effort: a missing profile must never
# fail the job.

if [[ -z "${ARTIFACTS:-}" ]]; then
  echo "ARTIFACTS is not set, skipping kubelet profile capture"
  exit 0
fi

OUT_DIR="${ARTIFACTS}/kubelet-pprof"
mkdir -p "${OUT_DIR}"

# Control-plane nodes plus a few worker nodes. Files are prefixed with the
# instance group because AWS node names are instance ids.
KUBELET_PPROF_WORKER_NODES="${KUBELET_PPROF_WORKER_NODES:-3}"
jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kops\.k8s\.io/instancegroup}{"\n"}{end}'

nodes=$(kubectl --request-timeout=30s get nodes -l node-role.kubernetes.io/control-plane -o jsonpath="${jsonpath}")
nodes+=$'\n'
nodes+=$(kubectl --request-timeout=30s get nodes -l node-role.kubernetes.io/node -o jsonpath="${jsonpath}" | head -n "${KUBELET_PPROF_WORKER_NODES}")

while read -r node ig; do
  if [[ -z "${node}" ]]; then
    continue
  fi
  prefix="${OUT_DIR}/${ig:-unknown}-${node}"
  for profile in heap allocs; do
    if ! kubectl --request-timeout=60s get --raw "/api/v1/nodes/${node}/proxy/debug/pprof/${profile}" > "${prefix}.${profile}.pb.gz"; then
      rm -f "${prefix}.${profile}.pb.gz"
    fi
  done
done <<< "${nodes}"

ls -la "${OUT_DIR}"
exit 0
