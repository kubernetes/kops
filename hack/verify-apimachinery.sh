#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
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

KOPS_ROOT=$(git rev-parse --show-toplevel)

export API_OPTIONS="--verify-only"
if make apimachinery-codegen; then
	echo "apimachinery is up to date"
else
	echo "\n FAIL: - the verify-apimachinery.sh test failed, apimachinery is not up to date"
	echo "\n FAIL: - please run the command 'make apimachinery'"
	exit 1
fi
