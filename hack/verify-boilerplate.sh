#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
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

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

boiler="${KOPS_ROOT}/hack/boilerplate/boilerplate.py $@"

files_need_boilerplate=( `${boiler}` )

if [[ -z ${files_need_boilerplate+x} ]]; then
    exit
fi

TO_REMOVE=(${KOPS_ROOT}/federation/model/bindata.go ${KOPS_ROOT}/upup/models/bindata.go)
TEMP_ARRAY=()

for pkg in "${files_need_boilerplate[@]}"; do
    for remove in "${TO_REMOVE[@]}"; do
        KEEP=true
        if [[ ${pkg} == ${remove} ]]; then
            KEEP=false
            break
        fi
    done
    if ${KEEP}; then
        TEMP_ARRAY+=(${pkg})
    fi
done

if [[ ${#TEMP_ARRAY[@]} -gt 0 ]]; then
  for file in "${TEMP_ARRAY[@]}"; do
    echo "FAIL: Boilerplate header is wrong for: ${file}"
  done
  echo "FAIL: Please execute ./hack/update-header.sh"
  exit 1
fi
