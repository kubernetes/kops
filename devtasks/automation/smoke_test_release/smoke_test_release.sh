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

WORKDIR=$(mktemp -d -t workflow.XXXXXXXX)
echo "WORKDIR is ${WORKDIR}"

cd "${WORKDIR}"
wget https://artifacts.k8s.io/binaries/kops/${VERSION}/linux/amd64/kops

chmod +x kops

# Verify version works
./kops version

# Verify version matches the version we expect
./kops version | grep ${VERSION}

GCP_PROJECT=$(gcloud config get project)
export KOPS_STATE_STORE="gs://kops-state-${GCP_PROJECT}/"

./kops get cluster || true # ignore error if cluster doesn't exist

# TEMP HACK
./kops delete cluster smoketest.k8s.local --yes || true

./kops create cluster smoketest.k8s.local --zones us-east4-a
./kops update cluster smoketest.k8s.local --yes --admin
./kops validate cluster --wait=10m

./kops delete cluster smoketest.k8s.local --yes
