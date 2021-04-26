#!/bin/bash

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

# Outputs the main kops version; as reported in `kops version`.

GITSHA=$(git describe --always 2>/dev/null)

# When we cut a new release we need to increment these accordingly
KOPS_RELEASE_VERSION=$(grep 'KOPS_RELEASE_VERSION\s*=' version.go  | awk '{print $3}' | sed -e 's_"__g')
KOPS_CI_VERSION=$(grep 'KOPS_CI_VERSION\s*=' version.go  |  awk '{print $3}' | sed -e 's_"__g')

if [[ -z "${VERSION}" ]]; then
  if [[ -z "${CI}" ]]; then
    VERSION=${KOPS_RELEASE_VERSION}
  else
    VERSION="${KOPS_CI_VERSION}+${GITSHA}"
  fi
fi

# If we are CI-building something that is exactly a tag, then we use that as the version.
# This let us do release (candidate) builds from our CI pipeline.
if [[ -n "${CI}" ]]; then
    EXACT_TAG=$(git describe --tags --exact-match 2>/dev/null || true)
    if [[ -n "${EXACT_TAG}" ]]; then
        VERSION="${EXACT_TAG#v}" # Remove the v prefix from the git tag
        if [[ "${VERSION}" != "${KOPS_RELEASE_VERSION}" ]]; then
            echo "Build was tagged with ${VERSION}, but version.go had version ${KOPS_RELEASE_VERSION}"
            exit 1
        fi
    fi
fi

echo "VERSION ${VERSION}"
