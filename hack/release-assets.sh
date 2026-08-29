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

# Shared definitions for the release scripts that publish and validate GitHub release assets.
# Sourced by promote-to-github.sh and validate-github-release.sh.
# shellcheck disable=SC2034  # the variables are consumed by the sourcing scripts

REPO="kubernetes/kops"
BASE_URL="https://artifacts.k8s.io/binaries/kops"

# Binaries published as GitHub release assets: source-path -> github-name.
# Each binary is accompanied by a .sha256 asset, which the scripts derive from this table.
# The upstream .sha256 files contain a bare hash followed by a newline.
declare -A BINARIES=(
  ["darwin/amd64/kops"]="kops-darwin-amd64"
  ["darwin/arm64/kops"]="kops-darwin-arm64"
  ["linux/amd64/kops"]="kops-linux-amd64"
  ["linux/arm64/kops"]="kops-linux-arm64"
  ["windows/amd64/kops.exe"]="kops-windows-amd64"
)

# Binaries that only some kOps releases ship: source-path -> github-name.
# channels was dropped from releases in kOps 1.36, protokube in kOps 1.37, and
# uncompressed nodeup after kOps 1.37.0-beta.1. nodeup.xz was added in that release.
declare -A OPTIONAL_BINARIES=(
  ["linux/amd64/channels"]="channels-linux-amd64"
  ["linux/arm64/channels"]="channels-linux-arm64"
  ["linux/amd64/protokube"]="protokube-linux-amd64"
  ["linux/arm64/protokube"]="protokube-linux-arm64"
  ["linux/amd64/nodeup"]="nodeup-linux-amd64"
  ["linux/arm64/nodeup"]="nodeup-linux-arm64"
  ["linux/amd64/nodeup.xz"]="nodeup-linux-amd64.xz"
  ["linux/arm64/nodeup.xz"]="nodeup-linux-arm64.xz"
)

# add_optional_binaries <version>: append the optional binaries that exist upstream for
# the given version to BINARIES, so releases keep full coverage without version-specific logic.
# Only a confirmed 404 counts as absent; any other probe outcome fails the script,
# so a transient error cannot silently drop an artifact from promotion or validation.
add_optional_binaries() {
  local version="$1" source code
  for source in "${!OPTIONAL_BINARIES[@]}"; do
    code=$(curl -sSIL --retry 3 --max-time 60 -o /dev/null -w '%{http_code}' "${BASE_URL}/${version}/${source}.sha256" || true)
    case "${code}" in
      200) BINARIES[$source]="${OPTIONAL_BINARIES[$source]}" ;;
      404) ;;
      *)
        echo "Error: probing ${BASE_URL}/${version}/${source}.sha256 returned HTTP ${code}" >&2
        exit 1
        ;;
    esac
  done
}

if command -v sha256sum &>/dev/null; then
  SHA256SUM=(sha256sum)
else
  SHA256SUM=(shasum -a 256)
fi
