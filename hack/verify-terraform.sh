#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Terraform versions
TF_TAG=0.13.5

PROVIDER_CACHE="${KOPS_ROOT}/.cache/terraform"

RC=0
while IFS= read -r -d '' -u 3 test_dir; do
  [ -f "${test_dir}/kubernetes.tf" ] || [ -f "${test_dir}/kubernetes.tf.json" ] || continue
  echo -e "${test_dir}\n"

  docker run --rm -e "TF_PLUGIN_CACHE_DIR=${PROVIDER_CACHE}" -v "${PROVIDER_CACHE}:${PROVIDER_CACHE}" -v "${test_dir}":"${test_dir}" -w "${test_dir}" --entrypoint=sh hashicorp/terraform:${TF_TAG} -c '/bin/terraform init >/dev/null && /bin/terraform validate' || RC=$?
done 3< <(find "${KOPS_ROOT}/tests/integration/update_cluster" -maxdepth 1 -type d -print0)

if [ $RC != 0 ]; then
  echo -e "\nTerraform validation failed\n"
  exit $RC
else
  echo -e "\nTerraform validation succeeded\n"
fi
