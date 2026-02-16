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

# We need to install metrics-server in the cluster and schedule it on the control-plane
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm upgrade --install metrics-server metrics-server/metrics-server \
    -n kube-system --set "args={--kubelet-insecure-tls}" \
    --set addonResizer.enabled=true

# We need an IG that has a single node for addons such as metrics-server, exec-service
# We used to call it heapster IG but its now called addons

if [[ "${CLOUD_PROVIDER}" == "aws" ]]; then
  kops create instancegroup addons --edit=false --role node --zone us-east-2b
  kops edit instancegroup addons --set=spec.machineType="${ADDONS_NODE_SIZE:-c7a.8xlarge}" \
    --set=spec.maxSize=1 --set=spec.minSize=1 --set=spec.image="ssm:/aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id"
elif [[ "${CLOUD_PROVIDER}" == "gce" ]]; then
  kops create instancegroup addons-us-east1-b-y86s --edit=false --role node --zone us-east1-b
  kops edit instancegroup addons-us-east1-b-y86s --set=spec.machineType="${ADDONS_NODE_SIZE:-c3-standard-8}" \
    --set=spec.maxSize=1 --set=spec.minSize=1 --set=spec.rootVolume.type=hyperdisk-balanced
fi

kops update cluster --yes
sleep 120 # it shouldn't take long to have the node up and ready
# To be replaced with kops validate instancegroup instead of the wait
