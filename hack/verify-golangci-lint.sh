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

# This script checks coding style for go language files
# Usage: `hack/verify-golangci-lint.sh`.

set -o errexit
set -o nounset
set -o pipefail

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# Ensure that we find the binaries we build before anything else.
export GOBIN="${TOOLS_BIN}"
PATH="${GOBIN}:${PATH}"

# Install golangci-lint
if ! command -v golangci-lint &> /dev/null; then
  cd "${KOPS_ROOT}/hack" || exit 1
  go install github.com/golangci/golangci-lint/cmd/golangci-lint
fi

cd "${KOPS_ROOT}"

# The config is in ${KOPS_ROOT}/.golangci.yaml
echo 'running golangci-lint ' >&2
res=0
if [[ "$#" -gt 0 ]]; then
    golangci-lint run "$@" >&2 || res=$?
else
    golangci-lint run ./... >&2 || res=$?
fi

# print a message based on the result
if [[ "$res" -eq 0 ]]; then
  echo 'Congratulations! All files are passing lint :-)'
else
  {
    echo
    echo 'Please review the above warnings. You can test via "./hack/verify-golangci-lint.sh"'
    echo 'If the above warnings do not make sense, you can exempt this warning with a comment'
    echo ' (if your reviewer is okay with it).'
    echo 'In general please prefer to fix the error, we have already disabled specific lints'
    echo ' that the project chooses to ignore.'
    echo 'See: https://golangci-lint.run/usage/false-positives/'
    echo
  } >&2
  exit 1
fi

# preserve the result
exit "$res"
