#!/bin/bash

# Copyright 2022 The Kubernetes Authors.
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

# This is a convenience script for developing kOps on Scaleway.
# It builds the code, including nodeup, and uploads to a custom Scaleway object bucket.
# It also sets KOPS_STATE_STORE and KOPS_BASE_URL to project-isolated values.
# To use, source the script.  For example `. hack/dev-build-scaleway.sh` (note the initial `.`)

# Can't use set -e in a script we want to source
#set -e

#set -x

# Dev environments typically do not need to test multiple architectures
KOPS_ARCH=amd64
export KOPS_ARCH

# We use the scaleway-cli
# https://github.com/scaleway/scaleway-cli
SCW_DEFAULT_PROJECT_ID=$(scw config get default-project-id)
SCW_DEFAULT_REGION=$(scw config get default-region)

# Build and upload to bucket
UPLOAD_DEST_BUCKET="kops-dev-${SCW_DEFAULT_PROJECT_ID}-${USER}"
export UPLOAD_DEST=s3://${UPLOAD_DEST_BUCKET}

# Scaleway object bucket creation that requires that the aws command line is installed
# https://www.scaleway.com/en/docs/storage/object/api-cli/object-storage-aws-cli/
aws s3 ls "${UPLOAD_DEST}" || aws s3 mb "${UPLOAD_DEST}" || return
make kops-install dev-upload-linux-${KOPS_ARCH} || return

# Set KOPS_BASE_URL
(tools/get_version.sh | grep VERSION | awk '{print $2}') || return
KOPS_VERSION=$(tools/get_version.sh | grep VERSION | awk '{print $2}')
export KOPS_BASE_URL=https://s3.${SCW_DEFAULT_REGION}.scw.cloud/${UPLOAD_DEST_BUCKET}/kops/${KOPS_VERSION}/

# Create the state-store bucket if it doesn't exist
KOPS_STATE_STORE="s3://kops-state-$(scw config get default-project-id)"
export KOPS_STATE_STORE
aws s3 ls "${KOPS_STATE_STORE}" || aws s3 mb "${KOPS_STATE_STORE}" || return

echo "SUCCESS"
