#!/bin/bash

# Copyright 2021 The Kubernetes Authors.
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

#set -x

node=$1

# bazel caching doesn't work when when we switch architectures
KOPS_ARCH=amd64
export KOPS_ARCH

make bazel-build-nodeup-linux-amd64

#echo "Uploading to ${node}"
#gcloud compute ssh ${node} -- rm -f /tmp/nodeup
#gcloud compute scp .bazel-bin/cmd/nodeup/linux-${KOPS_ARCH}/nodeup ${node}:/tmp/nodeup
#gcloud compute ssh ${node} -- sudo /tmp/nodeup --conf=/opt/kops/conf/kube_env.yaml --v=8

# Build and upload to bucket
UPLOAD_DEST_BUCKET="kops-dev-$(gcloud config get-value project)-${USER}"
export UPLOAD_DEST=gs://${UPLOAD_DEST_BUCKET}
gsutil ls "${UPLOAD_DEST}" || gsutil mb "${UPLOAD_DEST}" || return
make kops-install dev-upload-linux-${KOPS_ARCH} || return

echo "SUCCESS"
