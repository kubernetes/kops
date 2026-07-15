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

# Concurrent transfers, overridable via environment. Per-connection throughput to both the
# artifacts CDN and GitHub is throttled well below typical link speed, so parallelism is what
# drives wall-clock time; 10 covers all large binaries in one wave, and 4 upload streams saturate
# a ~400 Mbit/s uplink.
DOWNLOAD_PARALLELISM="${DOWNLOAD_PARALLELISM:-10}"
UPLOAD_PARALLELISM="${UPLOAD_PARALLELISM:-4}"

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
# Keep the checksum manifest outside WORKDIR so it is not uploaded with the assets.
CHECKSUMS=$(mktemp)
trap 'rm -rf "${WORKDIR}" "${CHECKSUMS}"' EXIT

echo "Downloading binaries from ${BASE_URL}/${VERSION}/ ..."

curl_args=(-fsSL --retry 3 --parallel --parallel-max "${DOWNLOAD_PARALLELISM}")
for source in "${!BINARIES[@]}"; do
  github_name="${BINARIES[$source]}"
  echo "  ${source} -> ${github_name}"
  curl_args+=(-o "${WORKDIR}/${github_name}" "${BASE_URL}/${VERSION}/${source}")
done

if ! curl "${curl_args[@]}"; then
  echo "Error: failed to download binaries from ${BASE_URL}/${VERSION}/" >&2
  exit 1
fi

echo "Verifying checksums ..."

if command -v sha256sum &>/dev/null; then
  SHA256SUM=(sha256sum)
else
  SHA256SUM=(shasum -a 256)
fi

# The downloaded .sha256 files contain a bare hash, so build a "<hash>  <file>" manifest that the
# tool's check mode can verify in one pass. Pass it as a file argument: BSD sha256sum does not
# read the manifest from stdin.
for checksum_file in "${WORKDIR}"/*.sha256; do
  read -r hash _ <"${checksum_file}"
  echo "${hash}  $(basename "${checksum_file%.sha256}")"
done >"${CHECKSUMS}"

if ! (cd "${WORKDIR}" && "${SHA256SUM[@]}" -c "${CHECKSUMS}"); then
  echo "Error: checksum verification failed" >&2
  exit 1
fi

echo "Uploading binaries to GitHub release ${TAG} ..."

# Upload assets in parallel; --clobber makes reruns after a partial failure succeed.
if ! printf '%s\0' "${WORKDIR}"/* | xargs -0 -n 1 -P "${UPLOAD_PARALLELISM}" gh release upload "${TAG}" --repo "${REPO}" --clobber; then
  echo "Error: failed to upload binaries to GitHub release ${TAG}" >&2
  exit 1
fi
