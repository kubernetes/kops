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

make gomod

changes=$(git status --porcelain go.mod go.sum vendor/ tests/e2e/go.mod  tests/e2e/go.sum || true)
if [ -n "${changes}" ]; then
  echo "ERROR: go modules are not up to date; please run: make gomod"
  echo "changed files:"
  printf "%s" "${changes}\n"
  echo "git diff:"
  git --no-pager diff
  exit 1
fi
