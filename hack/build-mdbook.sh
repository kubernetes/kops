#!/usr/bin/env bash

# Copyright 2026 The Kubernetes Authors.
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

# Installs a pinned mdBook binary (if missing) and builds the docs site.
# Usage: ./hack/build-mdbook.sh [build|serve]   (default: build)

set -o errexit
set -o nounset
set -o pipefail

MDBOOK_VERSION="${MDBOOK_VERSION:-0.5.2}"
CMD="${1:-build}"

KOPS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${KOPS_ROOT}/.build/bin"
MDBOOK="${BIN_DIR}/mdbook"

# Map host OS/arch to upstream release asset names.
case "$(uname -m)" in
  x86_64|amd64) ARCH="x86_64" ;;
  arm64|aarch64) ARCH="aarch64" ;;
  *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac
case "$(uname -s)" in
  Darwin) OS="apple-darwin" ;;
  Linux)
    # Upstream ships musl for aarch64 and gnu for x86_64.
    if [[ "${ARCH}" == "aarch64" ]]; then OS="unknown-linux-musl"; else OS="unknown-linux-gnu"; fi
    ;;
  *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

if [[ ! -x "${MDBOOK}" ]] || ! "${MDBOOK}" --version 2>/dev/null | grep -q "${MDBOOK_VERSION}"; then
  mkdir -p "${BIN_DIR}"
  TARBALL="mdbook-v${MDBOOK_VERSION}-${ARCH}-${OS}.tar.gz"
  URL="https://github.com/rust-lang/mdBook/releases/download/v${MDBOOK_VERSION}/${TARBALL}"
  echo "Downloading ${URL}"
  curl -fsSL "${URL}" | tar -xz -C "${BIN_DIR}"
fi

cd "${KOPS_ROOT}"
case "${CMD}" in
  build) "${MDBOOK}" build ;;
  serve) "${MDBOOK}" serve --hostname 0.0.0.0 --port 3000 --open ;;
  *) echo "unknown command: ${CMD}" >&2; exit 1 ;;
esac
