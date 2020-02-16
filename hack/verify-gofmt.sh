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

GOFMT="bazel run //:gofmt -- -s -w"

bad_files=$(git ls-files "*.go" | grep -v vendor | xargs $GOFMT -l)
if [[ -n "${bad_files}" ]]; then
  echo "FAIL: '$GOFMT' needs to be run on the following files: "
  echo "${bad_files}"
  echo "FAIL: please execute make gofmt"
  exit 1
fi
