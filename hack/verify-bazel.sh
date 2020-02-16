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

cd "${KOPS_ROOT}"

TMP_OUT=$(mktemp -d)
trap "{ rm -rf ${TMP_OUT}; }" EXIT

GOBIN="${TMP_OUT}" go install ./vendor/github.com/bazelbuild/bazel-gazelle/cmd/gazelle

gazelle_diff=$("${TMP_OUT}/gazelle" fix \
  -external=vendored \
  -mode=diff \
  -proto=disable \
  -repo_root="${KOPS_ROOT}")

if [[ -n "${gazelle_diff}" ]]; then
  echo "${gazelle_diff}" >&2
  echo >&2
  echo "FAIL: ./hack/verify-bazel.sh failed, as the bazel files are not up to date" >&2
  echo "FAIL: Please execute the following command: ./hack/update-bazel.sh" >&2
  exit 1
fi

# Make sure there are no BUILD files outside vendor - we should only have
# BUILD.bazel files.
old_build_files=$(find . -name BUILD \( -type f -o -type l \) \
  -not -path './vendor/*' | sort)
if [[ -n "${old_build_files}" ]]; then
  echo "One or more BUILD files found in the tree:" >&2
  echo "${old_build_files}" >&2
  echo >&2
  echo "FAIL: Only bazel files named BUILD.bazel are allowed." >&2
  echo "FAIL: Please move incorrectly named files to BUILD.bazel" >&2
  exit 1
fi
