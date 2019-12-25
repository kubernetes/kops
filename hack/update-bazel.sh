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

KOPS_ROOT=$(git rev-parse --show-toplevel)
cd ${KOPS_ROOT}

export GOPATH=${KOPS_ROOT}/../../../

TMP_OUT=$(mktemp -d)
trap "{ rm -rf ${TMP_OUT}; }" EXIT

GOBIN="${TMP_OUT}" go install ./vendor/github.com/bazelbuild/bazel-gazelle/cmd/gazelle

# manually remove BUILD file for k8s.io/apimachinery/pkg/util/sets/BUILD if it
# exists; there is a specific set-gen rule that breaks importing
# ref: https://github.com/kubernetes/kubernetes/blob/4e2f5e2212b05a305435ef96f4b49dc0932e1264/staging/src/k8s.io/apimachinery/pkg/util/sets/BUILD#L23-L49
# rm -f ${KOPS_ROOT}/vendor/k8s.io/apimachinery/pkg/util/sets/{BUILD,BUILD.bazel}

"${TMP_OUT}/gazelle" fix \
  -external=vendored \
  -mode=fix \
  -proto=disable \
  -repo_root="${KOPS_ROOT}"
