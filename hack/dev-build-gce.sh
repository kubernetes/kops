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

# This is a convenience script for developing kOps on GCE.
# It builds the code, including nodeup, and uploads to a custom GCS bucket.
# It also sets KOPS_STATE_STORE and KOPS_BASE_URL to project-isolated values.
# To use, source the script.  For example `. hack/dev-build-gce.sh` (note the initial `.`)

# Can't use set -e in a script we want to source
#set -e

#set -x

# Dev environments typically do not need to test multiple architectures
KOPS_ARCH=amd64
export KOPS_ARCH

# Build and upload to bucket
UPLOAD_DEST_BUCKET="kops-dev-$(gcloud config get-value project)-${USER}"
export UPLOAD_DEST=gs://${UPLOAD_DEST_BUCKET}
gsutil ls "${UPLOAD_DEST}" || gsutil mb "${UPLOAD_DEST}" || return
make kops-install dev-upload-linux-${KOPS_ARCH} || return

# Set KOPS_BASE_URL
(tools/get_version.sh | grep VERSION | awk '{print $2}') || return
KOPS_VERSION=$(tools/get_version.sh | grep VERSION | awk '{print $2}')
export KOPS_BASE_URL=https://storage.googleapis.com/${UPLOAD_DEST_BUCKET}/kops/${KOPS_VERSION}/

# Create the state-store bucket if it doesn't exist
KOPS_STATE_STORE="gs://kops-state-$(gcloud config get-value project)"
export KOPS_STATE_STORE
gsutil ls "${KOPS_STATE_STORE}" || gsutil mb "${KOPS_STATE_STORE}" || return

echo "SUCCESS"
