#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly DOCKER_ROOT=$(dirname "${BASH_SOURCE}")
cd "${DOCKER_ROOT}" # To ensure we're in the git tree

readonly GITISH=${BUILD_GITISH:-"$(git describe --always)"}
readonly ARCH=${BUILD_ARCH:-"linux/amd64"}
readonly NAME=${BUILD_NAME:-"ci-${GITISH}-${ARCH/\//-}"} # e.g. ci-bef7faf-linux-amd64
readonly TMPNAME="${NAME}-$(date +%s)" # e.g. ci-bef7faf-linux-amd64-12345678
readonly TAG=${BUILD_TAG:-"b.gcr.io/kops-ci/kops:${NAME}"}
readonly TMPTAG="${TAG}-$(date +%s)"
readonly LINK=${BUILD_LINK:-} # Also pushes to e.g. ci-{BUILD_LINK}-linux-amd64, i.e. for "latest"
readonly SYMBOLIC_TAG=${BUILD_TAG:-"b.gcr.io/kops-ci/kops:ci-${LINK}-${ARCH/\//-}"}

if [[ "${ARCH}" != "linux/amd64" ]]; then
  echo "!!! Alternate architecture build not supported yet. !!!"
  exit 1
fi

if [[ -z "${GITISH}" ]]; then
  echo "!!! git hash not found, are you sure you're in a git tree and git is installed? !!!"
  git config -l
  exit 1
fi

echo
echo "=== Building at ${GITISH} for ${ARCH} (note: unable to build unpushed changes) ==="
echo

# Build -> $TMPTAG
docker build -t "${TMPTAG}" --build-arg "KOPS_GITISH=${GITISH}" --build-arg "KUBECTL_ARCH=${ARCH}" --force-rm=true --rm=true --pull=true --no-cache=true "${DOCKER_ROOT}"

# Squash -> $TAG
docker create --name="${TMPNAME}" "${TMPTAG}"
docker export "${TMPNAME}" | docker import - "${TAG}"
gcloud docker push "${TAG}"

echo
echo "=== Pushed ${TAG} ==="
echo

if [[ -n "${LINK}" ]]; then
  docker tag "${TAG}" "${SYMBOLIC_TAG}"
  gcloud docker push "${SYMBOLIC_TAG}"
  echo
  echo "=== Pushed ${SYMBOLIC_TAG} ==="
  echo
fi

echo "=== Cleaning up ==="
echo
docker rm "${TMPNAME}"
docker rmi -f "${TAG}" "${TMPTAG}"
if [[ -n "${LINK}" ]]; then
  docker rmi -f "${SYMBOLIC_TAG}"
fi
