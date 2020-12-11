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

# Build apimachinery executables from vendor-ed dependencies

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

cd "${KOPS_ROOT}/hack" || exit 1

go build -o "${TOOLS_BIN}/conversion-gen" -v k8s.io/code-generator/cmd/conversion-gen
go build -o "${TOOLS_BIN}/deepcopy-gen" -v k8s.io/code-generator/cmd/deepcopy-gen
go build -o "${TOOLS_BIN}/defaulter-gen" -v k8s.io/code-generator/cmd/defaulter-gen
go build -o "${TOOLS_BIN}/client-gen" -v k8s.io/code-generator/cmd/client-gen
