#!/bin/bash -xe

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

VERSION=$(grep 'KOPS_RELEASE_VERSION\s*=' kops-version.go  | awk '{print $3}' | sed -e 's_"__g')

if [[ ! "${VERSION}" =~ ^([0-9]+[.][0-9]+)[.]([0-9]+)(-(alpha|beta)[.]([0-9]+))?$ ]]; then
  echo "Version ${VERSION} must be 'X.Y.Z', 'X.Y.Z-alpha.N', or 'X.Y.Z-beta.N'"
  exit 1
fi

MINOR=${BASH_REMATCH[1]}
RELEASE_BRANCH="release-${MINOR}"

if [ "$(git tag -l "v${VERSION}")" ]; then
  echo "Tag v${VERSION} already exists"
  exit 0
fi

git tag -a -m "Release ${VERSION}" "v${VERSION}"
git push origin "v${VERSION}"

if [[ ! "${VERSION}" =~ .0-beta.1$ ]]; then
  exit 0
fi

git branch "${RELEASE_BRANCH}"
git push origin "${RELEASE_BRANCH}"
