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

# InstanceGroup.Spec.Role is a comma-separated list of roles, so a group can carry more than one
# (for example "APIServer,Scheduler"). Any code that treats it as a single value will silently
# look at only one of the roles. This check looks for the two shapes that go wrong:
#
#   1. Comparing the role against a role constant with == or !=.
#   2. Switching on the role, where only the first matching case runs.
#
# Use the Has* helpers to ask whether a group carries a role, or PrimaryRole() where exactly one
# role can be represented (naming a cloud resource, selecting an IAM identity, a state-store
# path). A switch that genuinely needs to match the whole value can opt out with a trailing
# "kops:single-role-switch" comment explaining why.

set -o errexit
set -o nounset
set -o pipefail

KOPS_ROOT=$(git rev-parse --show-toplevel)
cd "${KOPS_ROOT}"

errors=0

# Exclude vendor, .build, and pkg/apis/kops/instancegroup.go, which is where the roles and the
# helpers themselves are defined.
files=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.build/*" -not -path "./pkg/apis/kops/instancegroup.go")

roles='(ControlPlane|Node|Bastion|APIServer|Etcd|Scheduler|KubeControllerManager)'
# Role constants, with or without a package qualifier (kops., api., unversioned., ...).
role_const="([a-zA-Z0-9_]+\.)?InstanceGroupRole${roles}\b"

# variable == constant, constant == variable, and the != forms.
equality_regex="(==|!=)[[:space:]]*${role_const}|${role_const}[[:space:]]*(==|!=)"

# switch on the role itself, e.g. "switch ig.Spec.Role {".
switch_regex="switch[[:space:]]+[^{]*\.Spec\.Role[[:space:]]*\{"

optout_marker='kops:single-role-switch'

for file in ${files}; do
    if grep -E "${equality_regex}" "${file}" > /dev/null; then
        echo "Verification failed in ${file}: direct comparison with an InstanceGroupRole constant:"
        grep -n -E "${equality_regex}" "${file}"
        errors=$((errors + 1))
    fi

    if grep -E "${switch_regex}" "${file}" | grep -v "${optout_marker}" > /dev/null; then
        echo "Verification failed in ${file}: switch on a possibly-composite InstanceGroup role:"
        grep -n -E "${switch_regex}" "${file}" | grep -v "${optout_marker}"
        errors=$((errors + 1))
    fi
done

if [ "${errors}" -ne 0 ]; then
  echo
  echo "Error: found ${errors} file(s) treating InstanceGroup.Spec.Role as a single role."
  echo "Spec.Role is a comma-separated list. Use HasControlPlane(), HasNode(), HasBastion(),"
  echo "HasAPIServer(), HasEtcd(), HasScheduler() or HasKubeControllerManager() to test for a"
  echo "role, or PrimaryRole() where only one role can be represented."
  exit 1
fi

echo "InstanceGroupRole comparison verification passed."
exit 0
