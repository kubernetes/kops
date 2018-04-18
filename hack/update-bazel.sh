#!/usr/bin/env bash
# Copyright 2016 The Kubernetes Authors.
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
# https://github.com/kubernetes/test-infra/issues/5699#issuecomment-348350792
cd ${KOPS_ROOT}
TMP_GOPATH=$(mktemp -d)

# manually remove BUILD file for k8s.io/apimachinery/pkg/util/sets/BUILD if it
# exists; there is a specific set-gen rule that breaks importing
# ref: https://github.com/kubernetes/kubernetes/blob/4e2f5e2212b05a305435ef96f4b49dc0932e1264/staging/src/k8s.io/apimachinery/pkg/util/sets/BUILD#L23-L49
# rm -f ${KOPS_ROOT}/vendor/k8s.io/apimachinery/pkg/util/sets/{BUILD,BUILD.bazel}

"${KOPS_ROOT}/hack/go_install_from_commit.sh" \
  github.com/bazelbuild/bazel-gazelle/cmd/gazelle \
  8bc6a862933eaa0d7431e15b308ceadc5729a6f9 \
  "${TMP_GOPATH}"

"${TMP_GOPATH}/bin/gazelle" fix \
  -external=vendored \
  -mode=fix \
  -proto=disable \
  -repo_root="${KOPS_ROOT}"
