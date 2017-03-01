# Copyright 2016 The Kubernetes Authors.
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

# required:
# KUBE_ROOT: path of the root of the Kubernetes repository

: "${KUBE_ROOT?Must set KUBE_ROOT env var}"

source "${KUBE_ROOT}/cluster/common.sh"

# Creates the required certificates for Heapster apiserver.
# $1: The public IP for the master.
#
# Assumed vars
#   KUBE_TEMP
#   MASTER_NAME
#
# Set vars
#   HEAPSTER_APISERVER_CA_CERT_BASE64
#   HEAPSTER_APISERVER_CERT_BASE64
#   HEAPSTER_APISERVER_KEY_BASE64
#
function create-apiserver-certs() {
  local -r primary_cn="${1}"
  local sans="IP:${1},DNS:${MASTER_NAME}"

  echo "Generating certs for alternate-names: ${sans}"

  local kube_temp="${KUBE_TEMP}/heapster"
  mkdir -p "${kube_temp}"
  KUBE_TEMP="${kube_temp}" PRIMARY_CN="${primary_cn}" SANS="${sans}" generate-certs

  local cert_dir="${kube_temp}/easy-rsa-master/easyrsa3"
  # By default, linux wraps base64 output every 76 cols, so we use 'tr -d' to remove whitespaces.
  # Note 'base64 -w0' doesn't work on Mac OS X, which has different flags.
  export HEAPSTER_APISERVER_CA_CERT_BASE64=$(cat "${cert_dir}/pki/ca.crt" | base64 | tr -d '\r\n')
  export HEAPSTER_APISERVER_CERT_BASE64=$(cat "${cert_dir}/pki/issued/${MASTER_NAME}.crt" | base64 | tr -d '\r\n')
  export HEAPSTER_APISERVER_KEY_BASE64=$(cat "${cert_dir}/pki/private/${MASTER_NAME}.key" | base64 | tr -d '\r\n')
}

# Creates token and basic auth credentials for Heapster apiserver.
#
# Set vars
#   HEAPSTER_API_KNOWN_TOKENS
#   HEAPSTER_API_BASIC_AUTH
#   KUBE_USER
#   KUBE_PASSWORD
#   HEAPSTER_API_TOKEN
#
function create-auth-config() {
  # Generate token
  HEAPSTER_API_TOKEN="$(dd if=/dev/urandom bs=128 count=1 2>/dev/null | base64 | tr -d "=+/" | dd bs=32 count=1 2>/dev/null)"
  export HEAPSTER_API_KNOWN_TOKENS="${HEAPSTER_API_TOKEN},admin,admin"
  # Generate basic auth credentials
  gen-kube-basicauth
  export HEAPSTER_API_BASIC_AUTH="${KUBE_PASSWORD},${KUBE_USER},admin"

  export KUBE_USER
  export KUBE_PASSWORD
  export HEAPSTER_API_TOKEN
}

# Creates kubeconfig for Heapster apiserver.
#
# Assumed vars
#   CONTEXT
#   KUBECONFIG
#   HEAPSTER_API_HOST
#   HEAPSTER_API_TOKEN
#   KUBE_USER
#   KUBE_PASSWORD
#
function create-heapster-kubeconfig() {
  KUBE_MASTER_IP="${HEAPSTER_API_HOST}:443" \
    CONTEXT="${CONTEXT}" \
    KUBE_BEARER_TOKEN="$HEAPSTER_API_TOKEN" \
    KUBE_USER="${KUBE_USER}" \
    KUBE_PASSWORD="${KUBE_PASSWORD}" \
    KUBECONFIG="${KUBECONFIG}" \
    create-kubeconfig
}
