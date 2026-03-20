# Copyright 2024 The Kubernetes Authors.
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

#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

set -x

if [[ -z "${VERSION:-}" ]]; then
    echo "Must specify VERSION"
    exit 1
fi

shipbot="go run github.com/kopeio/shipbot/cmd/shipbot@master"

WORKDIR=$(mktemp -d -t workflow.XXXXXXXX)
echo "WORKDIR is ${WORKDIR}"

git clone https://github.com/kubernetes/kops ${WORKDIR}/kops
cd ${WORKDIR}/kops
gh repo set-default kubernetes/kops

rm -rf ${WORKDIR}/releases
mkdir -p ${WORKDIR}/releases/${VERSION}/
gsutil rsync -r  gs://k8s-staging-kops/kops/releases/${VERSION}/ ${WORKDIR}/releases/${VERSION}/

git checkout v$VERSION
${shipbot} -tag v${VERSION} -config .shipbot.yaml -src ${WORKDIR}/releases/${VERSION}/
