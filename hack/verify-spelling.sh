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

#!/bin/bash

set -e

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"

OUTPUT_GOBIN="${REPO_ROOT}/_output/bin"
GOBIN="${OUTPUT_GOBIN}" go install ./vendor/github.com/client9/misspell/cmd/misspell

mkdir -p .build/docs
find . -type f \( -name "*.go*" -o -name "*.md*" \) -a -path "./docs/releases/*" -exec basename {} \; | \
	xargs -I{} sh -c 'sed -e "/^\* .*github.com\/kubernetes\/kops\/pull/d" docs/releases/{} > .build/docs/$(basename {})'
find . -type f \( -name "*.go*" -o -name "*.md*" \) -a \( -not -path "./vendor/*" -not -path "./docs/releases/*" \) | \
  sed -e /README-ES.md/d -e /node_modules/d |
		xargs ${OUTPUT_GOBIN}/misspell -error


