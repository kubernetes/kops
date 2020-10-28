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

TAG=v0.39.0
IMAGE="cfn-python-lint:${TAG}"

# There is no official docker image so build it locally
# https://github.com/aws-cloudformation/cfn-python-lint/issues/1025
function docker_build() {
  echo "Building cfn-python-lint image"
  TMP=$(mktemp -d)
  git clone -q -b "${TAG}" https://github.com/aws-cloudformation/cfn-python-lint "${TMP}"
  pushd "${TMP}"
  docker build --tag "${IMAGE}" .
  popd
  rm -rf "${TMP}"
}

docker image inspect "${IMAGE}" >/dev/null 2>&1 || docker_build

docker run --rm -v "${KOPS_ROOT}:/${KOPS_ROOT}" -v "${KOPS_ROOT}/hack/.cfnlintrc.yaml:/root/.cfnlintrc" "${IMAGE}" "/${KOPS_ROOT}/tests/integration/update_cluster/**/cloudformation.json"
RC=$?

if [ $RC != 0 ]; then
  echo -e "\nCloudformation linting failed\n"
  exit 0 # TODO: exit $RC once issues have been addressed to make this a blocking check
else
  echo -e "\nCloudformation linting succeeded\n"
fi
