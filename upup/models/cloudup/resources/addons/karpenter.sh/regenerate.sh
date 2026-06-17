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

# Regenerate k8s-1.19.yaml.template from the upstream karpenter Helm chart.
# Chart version and kops patches live in kustomization.yaml; kops value
# overrides are in helm-values.yaml.
set -euo pipefail
cd "$(dirname "$0")"

kustomize build --enable-helm . > k8s-1.19.yaml.template
echo "Wrote k8s-1.19.yaml.template"
