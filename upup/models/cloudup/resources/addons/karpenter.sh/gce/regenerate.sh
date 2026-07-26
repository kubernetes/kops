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

# Regenerate ../k8s-1.19-gce.yaml.template from the karpenter-provider-gcp Helm chart. Chart version
# and kops patches live in kustomization.yaml; kops value overrides are in helm-values.yaml. The
# kops-managed GCENodeClass/NodePool template (instancegroups.yaml.template) is appended so both
# ship as one addon.
#
# Self-hosted mode (GCENodeClass.spec.startupScript) requires the chart and controller release
# containing cloudpilot-ai/karpenter-provider-gcp#540; bump the chart version in kustomization.yaml
# (and the controller image default in pkg/model/components/karpenter.go) to the first release that
# includes it.
set -euo pipefail
cd "$(dirname "$0")"

kustomize build --enable-helm . > ../k8s-1.19-gce.yaml.template
cat instancegroups.yaml.template >> ../k8s-1.19-gce.yaml.template
echo "Wrote ../k8s-1.19-gce.yaml.template"
