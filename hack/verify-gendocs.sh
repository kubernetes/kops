#!/bin/bash

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

. $(dirname "${BASH_SOURCE}")/common.sh

TMP_DOCS="${KUBE_ROOT}/.build/docs"
rm -rf $TMP_DOCS
mkdir $TMP_DOCS

command -v uname >/dev/null 2>&1 || { echo >&2 "uname must be installed. Aborting."; exit 1; }
OS=$(uname -s)

platform=""
if [[ $OS == "Darwin" ]]; then
	platform="darwin/amd64/"	
else 
	platform="linux/amd64/"	
fi

locations=(
  "${GOPATH}/bin/kops"
  "${KUBE_ROOT}/.build/dist/${platform}/kops"
)
# List most recently-updated location.
BIN=$( (ls -t "${locations[@]}" 2>/dev/null || true) | head -1 )

command -v $BIN >/dev/null 2>&1 || { echo >&2 "kops must be installed. Please run make.  Aborting."; exit 1; }

KOPS_STATE_STORE= $BIN genhelpdocs --out $TMP_DOCS

if [[ "$(diff $TMP_DOCS ${KUBE_ROOT}/docs/cli)" != "" ]]; then
	  echo "List of generated docs doesn't match a freshly built list. Please run the command make gen-cli-docs."
	  exit 1
fi
