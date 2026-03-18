#!/usr/bin/env bash

# Copyright The Kubernetes Authors.
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

# Downloads release binaries from artifacts.k8s.io and uploads them
# to the corresponding GitHub release, replacing the shipbot tool.
#
# Usage: promote-to-github.sh <version>
#   e.g. promote-to-github.sh 1.30.0

set -o errexit
set -o nounset
set -o pipefail

REPO="kubernetes/kops"
BASE_URL="https://artifacts.k8s.io/binaries/kops"

# Binaries to upload: source-path -> github-name
declare -A BINARIES=(
  ["darwin/amd64/kops"]="kops-darwin-amd64"
  ["darwin/amd64/kops.sha256"]="kops-darwin-amd64.sha256"
  ["darwin/arm64/kops"]="kops-darwin-arm64"
  ["darwin/arm64/kops.sha256"]="kops-darwin-arm64.sha256"
  ["linux/amd64/kops"]="kops-linux-amd64"
  ["linux/amd64/kops.sha256"]="kops-linux-amd64.sha256"
  ["linux/arm64/kops"]="kops-linux-arm64"
  ["linux/arm64/kops.sha256"]="kops-linux-arm64.sha256"
  ["windows/amd64/kops.exe"]="kops-windows-amd64"
  ["windows/amd64/kops.exe.sha256"]="kops-windows-amd64.sha256"
  ["linux/amd64/nodeup"]="nodeup-linux-amd64"
  ["linux/amd64/nodeup.sha256"]="nodeup-linux-amd64.sha256"
  ["linux/arm64/nodeup"]="nodeup-linux-arm64"
  ["linux/arm64/nodeup.sha256"]="nodeup-linux-arm64.sha256"
  ["linux/amd64/protokube"]="protokube-linux-amd64"
  ["linux/amd64/protokube.sha256"]="protokube-linux-amd64.sha256"
  ["linux/arm64/protokube"]="protokube-linux-arm64"
  ["linux/arm64/protokube.sha256"]="protokube-linux-arm64.sha256"
  ["linux/amd64/channels"]="channels-linux-amd64"
  ["linux/amd64/channels.sha256"]="channels-linux-amd64.sha256"
  ["linux/arm64/channels"]="channels-linux-arm64"
  ["linux/arm64/channels.sha256"]="channels-linux-arm64.sha256"
)

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <version>" >&2
  echo "  e.g. $0 1.30.0" >&2
  exit 1
fi

VERSION="$1"
# Strip leading 'v' if provided
VERSION="${VERSION#v}"
TAG="v${VERSION}"

if ! command -v gh &>/dev/null; then
  echo "Error: gh (GitHub CLI) is required but not found in PATH" >&2
  exit 1
fi

WORKDIR=$(mktemp -d)
trap 'rm -rf "${WORKDIR}"' EXIT

echo "Downloading binaries from ${BASE_URL}/${VERSION}/ ..."

for source in "${!BINARIES[@]}"; do
  github_name="${BINARIES[$source]}"
  dest="${WORKDIR}/${github_name}"
  url="${BASE_URL}/${VERSION}/${source}"

  echo "  ${source} -> ${github_name}"
  if ! curl -fsSL --retry 3 -o "${dest}" "${url}"; then
    echo "Error: failed to download ${url}" >&2
    exit 1
  fi
done

echo "Uploading binaries to GitHub release ${TAG} ..."

if ! gh release upload "${TAG}" --repo "${REPO}" "${WORKDIR}"/*; then
  echo "Error: failed to upload binaries to GitHub release ${TAG}" >&2
  exit 1
fi
