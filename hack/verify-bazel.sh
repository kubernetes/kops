#!/usr/bin/env bash
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

set -o errexit
set -o nounset
set -o pipefail

export KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..

# Example:  kube::util::trap_add 'echo "in trap DEBUG"' DEBUG
# See: http://stackoverflow.com/questions/3338030/multiple-bash-traps-for-the-same-signal
trap_add() {
  local trap_add_cmd
  trap_add_cmd=$1
  shift

  for trap_add_name in "$@"; do
    local existing_cmd
    local new_cmd

    # Grab the currently defined trap commands for this trap
    existing_cmd=`trap -p "${trap_add_name}" |  awk -F"'" '{print $2}'`

    if [[ -z "${existing_cmd}" ]]; then
      new_cmd="${trap_add_cmd}"
    else
      new_cmd="${trap_add_cmd};${existing_cmd}"
    fi

    # Assign the test
    trap "${new_cmd}" "${trap_add_name}"
  done
}

_tmpdir="$(mktemp -d -t verify-bazel.XXXXXX)"
trap_add "rm -rf ${_tmpdir}" EXIT

_tmp_gopath="${_tmpdir}/go"
_tmp_kuberoot="${_tmp_gopath}/src/k8s.io/kops"
mkdir -p "${_tmp_kuberoot}/.."
cp -a "${KUBE_ROOT}" "${_tmp_kuberoot}/.."

cd "${_tmp_kuberoot}"
GOPATH="${_tmp_gopath}" bazel run //:gazelle

diff=$(diff -Naupr "${KUBE_ROOT}" "${_tmp_kuberoot}" || true)

if [[ -n "${diff}" ]]; then
  echo "${diff}"
  echo
  echo "Run make bazel-gazelle"
  exit 1
fi
