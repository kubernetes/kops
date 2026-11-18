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
    --set=spec.image="${INSTANCE_IMAGE:-ubuntu-os-cloud/ubuntu-2404-noble-amd64-v20261016}"
elif [[ "${CLOUD_PROVIDER}" == "aws" ]]; then
  kops create instancegroup addons --edit=false --role node --subnet us-east-2a
  kops edit instancegroup addons --set=spec.machineType="${ADDONS_NODE_SIZE:-c7a.8xlarge}" \
    --set=spec.maxSize=1 --set=spec.minSize=1 \
    --set=spec.image="${INSTANCE_IMAGE:-ssm:/aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id}"
  # The ec2 scale jobs set PROMETHEUS_PVC_STORAGE_CLASS=io2. CL2 only puts that name on the PVC of Prometheus,
  # it does not create the storage class, and kOps does not install one with this name.
  kubectl apply -f - <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: io2
parameters:
  type: io2
  iops: "${KOPS_EBS_CSI_DRIVER_IO2_IOPS:-5000}"
  encrypted: "true"
provisioner: ebs.csi.aws.com
allowVolumeExpansion: true
volumeBindingMode: WaitForFirstConsumer
EOF
fi

kops update cluster --yes
# TODO: Replaced with `kops validate instancegroup --wait 10m` when it's fixed for 5k node clusters.
sleep 120 # should take around 2 minutes to get node read
