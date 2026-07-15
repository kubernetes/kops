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

# shellcheck source=hack/release-assets.sh
. "$(dirname "${BASH_SOURCE[0]}")/release-assets.sh"

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

echo "Fetching asset digests for GitHub release ${TAG} ..."

# The by-tag endpoint does not serve draft releases, so search the release list instead.
assets=$(gh api --paginate "repos/${REPO}/releases?per_page=100" \
  --jq ".[] | select(.tag_name == \"${TAG}\") | .assets[] | \"\(.name) \(.digest)\"")

if [[ -z "${assets}" ]]; then
  echo "Error: GitHub release ${TAG} not found or has no assets" >&2
  exit 1
fi

declare -A DIGESTS=()
while read -r name digest; do
  DIGESTS[$name]="${digest}"
done <<<"${assets}"

failed=0

# check <asset-name> <expected-sha256>: compare against the digest GitHub reports for the asset
check() {
  local actual="${DIGESTS[$1]:-}"
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
  hash=$(curl -fsSL --retry 3 "${BASE_URL}/${VERSION}/${source}.sha256")
  check "${BINARIES[$source]}" "${hash}"
  # The .sha256 asset must be byte-identical to the artifacts.k8s.io file, which is the bare hash
  # followed by a newline; reconstruct those bytes to compute its expected digest.
  check "${BINARIES[$source]}.sha256" "$(printf '%s\n' "${hash}" | "${SHA256SUM[@]}" | cut -d' ' -f1)"
done

# Anything else attached to the release was not published by promote-to-github.sh.
declare -A EXPECTED=()
for github_name in "${BINARIES[@]}"; do
  EXPECTED[$github_name]=1
  EXPECTED[$github_name.sha256]=1
done

for name in $(printf '%s\n' "${!DIGESTS[@]}" | sort); do
  if [[ -z "${EXPECTED[$name]:-}" ]]; then
    echo "${name}: UNEXPECTED asset on GitHub release"
    failed=1
  fi
done

if [[ "${failed}" -ne 0 ]]; then
  echo "Error: GitHub release ${TAG} does not match artifacts.k8s.io" >&2
  exit 1
fi

echo "GitHub release ${TAG} matches artifacts.k8s.io"
