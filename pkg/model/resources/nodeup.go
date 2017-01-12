package resources

var AWSNodeUpTemplate = `#!/bin/bash
# Copyright 2016 The Kubernetes Authors All rights reserved.
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

set -o errexit
set -o nounset
set -o pipefail

NODEUP_URL={{ NodeUpSource }}
NODEUP_HASH={{ NodeUpSourceHash }}

PRE_INSTALL_SCRIPT_URL={{ PreInstallScriptSource }}
PRE_INSTALL_SCRIPT_HASH={{ PreInstallScriptHash }}

POST_INSTALL_SCRIPT_URL={{ PostInstallScriptSource }}
POST_INSTALL_SCRIPT_HASH={{ PostInstallScriptHash }}

function ensure-install-dir() {
  INSTALL_DIR="/var/cache/kubernetes-install"
  mkdir -p ${INSTALL_DIR}
  cd ${INSTALL_DIR}
}

# Retry a download until we get it. Takes a hash and a set of URLs.
#
# $1 is the sha1 of the URL. Can be "" if the sha1 is unknown.
# $2+ are the URLs to download.
download-or-bust() {
  local -r hash="$1"
  shift 1

  urls=( $* )
  while true; do
    for url in "${urls[@]}"; do
      local file="${url##*/}"
      rm -f "${file}"
      if ! curl -f --ipv4 -Lo "${file}" --connect-timeout 20 --retry 6 --retry-delay 10 "${url}"; then
        echo "== Failed to download ${url}. Retrying. =="
      elif [[ -n "${hash}" ]] && ! validate-hash "${file}" "${hash}"; then
        echo "== Hash validation of ${url} failed. Retrying. =="
      else
        if [[ -n "${hash}" ]]; then
          echo "== Downloaded ${url} (SHA1 = ${hash}) =="
        else
          echo "== Downloaded ${url} =="
        fi
        return
      fi
    done

    echo "All downloads failed; sleeping before retrying"
    sleep 60
  done
}

validate-hash() {
  local -r file="$1"
  local -r expected="$2"
  local actual

  actual=$(sha1sum ${file} | awk '{ print $1 }') || true
  if [[ "${actual}" != "${expected}" ]]; then
    echo "== ${file} corrupted, sha1 ${actual} doesn't match expected ${expected} =="
    return 1
  fi
}

function split-commas() {
  echo $1 | tr "," "\n"
}

# Takes the URL of a file, downloads the file's hash if needed, and then
# downloads the file.
#
# $1 the sha1 of the resource
# $2 the url for the resource
#
# Sets $filename as the filename of the resource
function try-download() {
  local -r hash="$1"
  local -r url_list=( $(split-commas "${url_list}") )

  filename="${url_list[0]##*/}"
  if [[ -n "${hash:-}" ]]; then
    local -r file_hash="${hash}"
  else
    echo "Downloading sha1 (not passed as argument)"
    download-or-bust "" "${url_list[@]/%/.sha1}"
    local -r file_hash=$(cat "${filename}.sha1")
  fi

  echo "Downloading ${filename}"
  download-or-bust "${file_hash}" "${url_list[@]}"

  chmod +x "${filename}" &
}

# Attempts to download a resource
function download() {
  local -r hash="$1"
  local -r url_list="$2"

  until try-download "${hash}" "${url_list}"; do
    sleep 15
    echo "Couldn't download ${url_list}. Retrying..."
  done
}

function run() {
  local -r cmd="$1"

  echo "Running '${cmd}'"
  sleep 1
  ( cd ${INSTALL_DIR}; eval "./${cmd}" )
}
####################################################################################

/bin/systemd-machine-id-setup || echo "failed to set up ensure machine-id configured"
ensure-install-dir

if [[ ! -z ${PRE_INSTALL_SCRIPT_URL} ]]; then
  echo "== pre-install script starting"
  download "${PRE_INSTALL_SCRIPT_HASH}" "${PRE_INSTALL_SCRIPT_URL}"
  run "${filename}"
  echo "== pre-install script done"
fi

echo "== nodeup node config starting =="

cat > kube_env.yaml << __EOF_KUBE_ENV
{{ KubeEnv }}
__EOF_KUBE_ENV

download "${NODEUP_HASH}" "${NODEUP_URL}"
# We run in the background to work around https://github.com/docker/docker/issues/23793
run "${filename} --conf=/var/cache/kubernetes-install/kube_env.yaml --v=8" &
wait

echo "== nodeup node config done =="

if [[ ! -z ${POST_INSTALL_SCRIPT_URL} ]]; then
  echo "== post-install script starting"
  download "${POST_INSTALL_SCRIPT_HASH}" "${POST_INSTALL_SCRIPT_URL}"
  run "${filename}"
  echo "== post-install script done"
fi

`
