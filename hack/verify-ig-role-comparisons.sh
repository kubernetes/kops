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

set -o errexit
set -o nounset
set -o pipefail

KOPS_ROOT=$(git rev-parse --show-toplevel)
cd "${KOPS_ROOT}"

errors=0

# Find all .go files, excluding vendor, .build, and the file where roles are defined.
# We also exclude pkg/apis/kops/instancegroup.go
files=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.build/*" -not -path "./pkg/apis/kops/instancegroup.go")

# Regex to match == or != with InstanceGroupRole constants
# We check for constants with or without package prefixes (e.g. kops., api., unversioned.)
# We match both:
#   variable == constant
#   constant == variable
#   variable != constant
#   constant != variable
REGEX='==[[:space:]]*([a-zA-Z0-9_]+\.)?InstanceGroupRole(ControlPlane|Node|Bastion|APIServer|Etcd|Scheduler|CloudControllerManager|KubeControllerManager)\b|\b([a-zA-Z0-9_]+\.)?InstanceGroupRole(ControlPlane|Node|Bastion|APIServer)[[:space:]]*==|!=[[:space:]]*([a-zA-Z0-9_]+\.)?InstanceGroupRole(ControlPlane|Node|Bastion|APIServer)\b|\b([a-zA-Z0-9_]+\.)?InstanceGroupRole(ControlPlane|Node|Bastion|APIServer)[[:space:]]*!='

for file in $files; do
    if grep -E "${REGEX}" "${file}" > /dev/null; then
      echo "Verification failed in ${file}: direct comparison with InstanceGroupRole constant found:"
        grep -n -E "${REGEX}" "${file}"
        errors=$((errors + 1))
    fi
done

if [ "${errors}" -ne 0 ]; then
  echo "Error: Found ${errors} files with direct InstanceGroupRole comparisons (== or !=). Use HasControlPlane(), HasNode(), HasBastion(), HasAPIServer(), HasEtcd(), HasScheduler(), HasCloudControllerManager() or HasKubeControllerManager() instead."
    exit 1
fi

echo "InstanceGroupRole comparison verification passed."
exit 0
