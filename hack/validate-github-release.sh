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

# Validates that the assets attached to a GitHub release match the canonical .sha256 checksums
# published on artifacts.k8s.io, using the SHA-256 digests that the GitHub API reports for each
# asset (no binaries are downloaded).
#
# Usage: validate-github-release.sh <version>
#   e.g. validate-github-release.sh 1.30.0

set -o errexit
set -o nounset
set -o pipefail

REPO="kubernetes/kops"
BASE_URL="https://artifacts.k8s.io/binaries/kops"

# Binaries to validate: source-path -> github-name.
# Each binary is checked together with its .sha256 companion asset.
declare -A BINARIES=(
  ["darwin/amd64/kops"]="kops-darwin-amd64"
  ["darwin/arm64/kops"]="kops-darwin-arm64"
  ["linux/amd64/kops"]="kops-linux-amd64"
  ["linux/arm64/kops"]="kops-linux-arm64"
  ["windows/amd64/kops.exe"]="kops-windows-amd64"
  ["linux/amd64/nodeup"]="nodeup-linux-amd64"
  ["linux/arm64/nodeup"]="nodeup-linux-arm64"
  ["linux/amd64/protokube"]="protokube-linux-amd64"
  ["linux/arm64/protokube"]="protokube-linux-arm64"
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

if command -v sha256sum &>/dev/null; then
  SHA256SUM=(sha256sum)
else
  SHA256SUM=(shasum -a 256)
fi

echo "Fetching asset digests for GitHub release ${TAG} ..."

# The by-tag endpoint does not serve draft releases, so search the release list instead.
assets=$(gh api --paginate "repos/${REPO}/releases?per_page=100" \
  --jq ".[] | select(.tag_name == \"${TAG}\") | .assets[] | \"\(.name) \(.digest)\"")

if [[ -z "${assets}" ]]; then
  echo "Error: GitHub release ${TAG} not found or has no assets" >&2
  exit 1
fi

failed=0

# check <asset-name> <expected-sha256>: compare against the digest GitHub reports for the asset
check() {
  local actual
  actual=$(awk -v name="$1" '$1 == name {print $2}' <<<"${assets}")
  if [[ -z "${actual}" ]]; then
    echo "$1: MISSING from GitHub release"
    failed=1
  elif [[ "${actual}" != "sha256:$2" ]]; then
    echo "$1: FAILED (expected sha256:$2, got ${actual})"
    failed=1
  else
    echo "$1: OK"
  fi
}

for source in $(printf '%s\n' "${!BINARIES[@]}" | sort); do
  url="${BASE_URL}/${VERSION}/${source}.sha256"
  check "${BINARIES[$source]}" "$(curl -fsSL --retry 3 "${url}")"
  # The .sha256 asset itself must be byte-identical to the artifacts.k8s.io file.
  check "${BINARIES[$source]}.sha256" "$(curl -fsSL --retry 3 "${url}" | "${SHA256SUM[@]}" | cut -d' ' -f1)"
done

# Anything else attached to the release was not published by promote-to-github.sh.
expected=$(for name in "${BINARIES[@]}"; do printf '%s\n%s.sha256\n' "${name}" "${name}"; done | sort)
while read -r name; do
  echo "${name}: UNEXPECTED asset on GitHub release"
  failed=1
done < <(comm -23 <(awk '{print $1}' <<<"${assets}" | sort) <(echo "${expected}"))

if [[ "${failed}" -ne 0 ]]; then
  echo "Error: GitHub release ${TAG} does not match artifacts.k8s.io" >&2
  exit 1
fi

echo "GitHub release ${TAG} matches artifacts.k8s.io"
