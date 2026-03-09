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

set -e
set -x

# We need an instance group with a large single single node for addons such as prometheus, exec-service, etc.
# In kubeup we used to call it heapster due to name of first addon, but now we call it addons.

if [[ "${CLOUD_PROVIDER}" == "gce" ]]; then
  kops create instancegroup addons --edit=false --role node --zone us-east1-b
  kops edit instancegroup addons --set=spec.machineType="${ADDONS_NODE_SIZE:-c3-standard-22}" \
    --set=spec.maxSize=1 --set=spec.minSize=1 --set=spec.rootVolume.type=hyperdisk-balanced \
    --set=spec.image="${INSTANCE_IMAGE:-ubuntu-os-cloud/ubuntu-2404-noble-amd64-v20251001}"
  kops update cluster --yes
  kops validate cluster --wait 10m
fi
