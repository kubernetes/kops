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
cd ${KOPS_ROOT}

# Update gobindata to reflect any yaml changes
make kops-gobindata

# Don't override variables that are commonly used in dev, but shouldn't be in our tests
export KOPS_BASE_URL=
export DNSCONTROLLER_IMAGE=

# Run the tests in "autofix mode"
HACK_UPDATE_EXPECTED_IN_PLACE=1 go test ./... -count=1
