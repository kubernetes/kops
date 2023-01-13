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

set -o errexit
set -o nounset
set -o pipefail

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

cd "${KOPS_ROOT}"

KOPS_RELEASE_VERSION=$(grep 'KOPS_RELEASE_VERSION\s*=' kops-version.go  | awk '{print $3}' | sed -e 's_"__g')
"$(dirname "${BASH_SOURCE[0]}")/set-version" "${KOPS_RELEASE_VERSION}"

changed_files=$(git status --porcelain --untracked-files=no || true)
if [ -n "${changed_files}" ]; then
   echo "Detected that version generation is needed"
   echo "changed files:"
   printf "%s" "${changed_files}\n"
   echo "git diff:"
   git --no-pager diff
   echo "To fix: run 'hack/set-version ${KOPS_RELEASE_VERSION}'"
   exit 1
fi

