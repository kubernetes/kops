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

GOBIN="${TOOLS_BIN}" go install ./vendor/github.com/bazelbuild/bazel-gazelle/cmd/gazelle

"${TOOLS_BIN}/gazelle" fix \
  -external=vendored \
  -exclude=tests/e2e \
  -mode=fix \
  -proto=disable \
  -repo_root="${KOPS_ROOT}"
