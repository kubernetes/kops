#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly DOCKER_ROOT=$(dirname "${BASH_SOURCE}")
readonly GITISH="$(git describe --always)"
readonly ARCH=${BUILD_ARCH:-"linux/amd64"}
readonly NAME=${BUILD_NAME:-"ci-${GITISH}-${ARCH/\//-}"} # e.g. ci-bef7faf-linux-amd64
readonly TMPNAME="${NAME}-$(date +%s)" # e.g. ci-bef7faf-linux-amd64-12345678
readonly TAG=${BUILD_DOCKER_TAG:-"b.gcr.io/kops-ci/kops:${NAME}"}
readonly PUSH_TAG=${BUILD_PUSH_TAG:-"no"}
readonly CLEAN_TAG=${BUILD_CLEAN_TAG:-"yes"}
readonly TMPTAG="${TAG}-$(date +%s)"
readonly LINK=${BUILD_LINK:-} # Also pushes to e.g. ci-{BUILD_LINK}-linux-amd64, i.e. for "latest"
readonly SYMBOLIC_TAG=${BUILD_SYMBOLIC_TAG:-"b.gcr.io/kops-ci/kops:ci-${LINK}-${ARCH/\//-}"}

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
echo "=== Copying src to docker/_src ==="
echo

rsync -a --exclude=/docker/ "${DOCKER_ROOT}/.." "${DOCKER_ROOT}/_src"

echo
echo "=== Building at ${GITISH} for ${ARCH} ==="
echo

# Build -> $TMPTAG
docker build -t "${TMPTAG}" --build-arg "KUBECTL_ARCH=${ARCH}" --force-rm=true --rm=true --pull=true --no-cache=true "${DOCKER_ROOT}"

# Squash -> $TAG
docker create --name="${TMPNAME}" "${TMPTAG}"
docker export "${TMPNAME}" | docker import - "${TAG}"

if [[ "${PUSH_TAG}" == "yes" ]]; then
  echo
  echo "=== Pushing ${TAG} ==="
  echo

  gcloud docker push "${TAG}"
fi

if [[ -n "${LINK}" ]]; then
  echo
  echo "=== Pushing ${SYMBOLIC_TAG} ==="
  echo
  docker tag "${TAG}" "${SYMBOLIC_TAG}"
  gcloud docker push "${SYMBOLIC_TAG}"
fi

echo
echo "=== Cleaning up ==="
echo
docker rm "${TMPNAME}" || true
docker rmi -f "${TMPTAG}" || true
if [[ -n "${LINK}" ]]; then
  docker rmi -f "${SYMBOLIC_TAG}" || true
fi
if [[ "${CLEAN_TAG}" == "yes" ]]; then
  docker rmi -f "${TAG}" || true
else
  echo
  echo "=== ${TAG} leaked (BUILD_CLEAN_TAG not set) ==="
  echo
fi
