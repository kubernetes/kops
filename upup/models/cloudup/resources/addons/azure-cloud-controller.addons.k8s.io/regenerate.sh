#!/usr/bin/env bash

# Copyright The Kubernetes Authors.
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

# Regenerate k8s-1.31.yaml.template from the upstream cloud-provider-azure
# Helm chart. Chart version and kops patches live in kustomization.yaml.
set -euo pipefail
cd "$(dirname "$0")"

# replicas is a Go template expression, which kustomize can only carry as a quoted string.
# Unquote it so the rendered manifest has an integer.
kustomize build --enable-helm . \
  | sed "s/^\(  replicas: \)'\(.*\)'$/\1\2/" \
  > k8s-1.31.yaml.template
echo "Wrote k8s-1.31.yaml.template"
