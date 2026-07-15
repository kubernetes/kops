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

# shellcheck source=hack/release-assets.sh
. "$(dirname "${BASH_SOURCE[0]}")/release-assets.sh"

# Concurrent transfers, overridable via environment. Per-connection throughput to both the
# artifacts CDN and GitHub is throttled well below typical link speed, so parallelism is what
# drives wall-clock time; 10 covers all large binaries in one wave, and 4 upload streams saturate
# a ~400 Mbit/s uplink.
DOWNLOAD_PARALLELISM="${DOWNLOAD_PARALLELISM:-10}"
UPLOAD_PARALLELISM="${UPLOAD_PARALLELISM:-4}"

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

curl_args=(-fsSL --retry 3 --parallel --parallel-max "${DOWNLOAD_PARALLELISM}")
for source in "${!BINARIES[@]}"; do
  github_name="${BINARIES[$source]}"
  echo "  ${source} -> ${github_name}"
  curl_args+=(-o "${WORKDIR}/${github_name}" "${BASE_URL}/${VERSION}/${source}")
  curl_args+=(-o "${WORKDIR}/${github_name}.sha256" "${BASE_URL}/${VERSION}/${source}.sha256")
done

if ! curl "${curl_args[@]}"; then
  echo "Error: failed to download binaries from ${BASE_URL}/${VERSION}/" >&2
  exit 1
fi

echo "Verifying checksums ..."

# The downloaded .sha256 files contain a bare hash, so build a "<hash>  <file>" manifest that the
# tool's check mode can verify in one pass.
for checksum_file in "${WORKDIR}"/*.sha256; do
  read -r hash _ <"${checksum_file}"
  github_name="${checksum_file##*/}"
  echo "${hash}  ${github_name%.sha256}"
done >"${WORKDIR}/SHA256SUMS"

if ! (cd "${WORKDIR}" && "${SHA256SUM[@]}" -c SHA256SUMS); then
  echo "Error: checksum verification failed" >&2
  exit 1
fi

# Remove the manifest so it is not uploaded with the assets.
rm "${WORKDIR}/SHA256SUMS"

echo "Uploading binaries to GitHub release ${TAG} ..."

# Upload assets in parallel; --clobber makes reruns after a partial failure succeed.
if ! printf '%s\0' "${WORKDIR}"/* | xargs -0 -n 1 -P "${UPLOAD_PARALLELISM}" gh release upload "${TAG}" --repo "${REPO}" --clobber; then
  echo "Error: failed to upload binaries to GitHub release ${TAG}" >&2
  exit 1
fi
