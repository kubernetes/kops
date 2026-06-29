#!/bin/bash

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


set -o errexit
set -o nounset
set -o pipefail

set -x

if [[ -z "${VERSION:-}" ]]; then
    echo "Must specify VERSION"
    exit 1
fi

GITHUB_USER=$(gh api user | jq -r '.login')

MAJOR=$(echo $VERSION | cut -d. -f1)
MINOR=$(echo $VERSION | cut -d. -f2)
PATCH=$(echo $VERSION | cut -d. -f3-)

RELEASE_BRANCH="release-${MAJOR}.${MINOR}"

PR_BRANCH_NAME=create_release_${VERSION}

# If first beta (.0-beta.1)
if [[ "${PATCH}" == "0-beta.1" ]]; then
  RELEASE_BRANCH="master"
fi

WORKDIR=$(mktemp -d -t workflow.XXXXXXXX)
echo "WORKDIR is ${WORKDIR}"

git clone https://github.com/kubernetes/kops ${WORKDIR}/kops -b ${RELEASE_BRANCH} --depth=10
cd ${WORKDIR}/kops
gh repo set-default kubernetes/kops

git checkout ${RELEASE_BRANCH}
git checkout -b ${PR_BRANCH_NAME}

hack/set-version ${VERSION}

hack/update-expected.sh || true # Expected to fail first time because there will be updates
find . -name "*.bak" -delete
hack/update-expected.sh 2>&1 # Expected to succeed second time because there should be no updates
find . -name "*.bak" -delete

VERSION=$(tools/get_version.sh | grep VERSION | awk '{print $2}')
git add . && git commit -m "Release ${VERSION}"

git remote add fork https://github.com/${GITHUB_USER}/kops
git push fork --force

pwd
gh pr create -l tide/merge-method-squash \
  --base ${RELEASE_BRANCH} --head ${GITHUB_USER}:${PR_BRANCH_NAME} \
  --title "Release ${VERSION}" \
  --body "Release ${VERSION}"