#!/usr/bin/env bash

# Copyright 2026 The Kubernetes Authors.
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

# Start hubble-relay port-forward in background
kubectl port-forward -n kube-system deployment/hubble-relay 4245:4245 &

# Download cilium-cli
CILIUM_CLI_VERSION="v0.14.8"
WORKSPACE="${WORKSPACE:-$(mktemp -d)}"
wget -qO- "https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-amd64.tar.gz" | tar xz -C "${WORKSPACE}"

# Run connectivity test
"${WORKSPACE}/cilium" connectivity test --all-flows
