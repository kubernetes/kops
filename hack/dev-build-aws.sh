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
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export S3_BUCKET_NAME=kops-dev-${ACCOUNT_ID}-${USER}

export UPLOAD_DEST=s3://${S3_BUCKET_NAME}
aws s3 ls "${UPLOAD_DEST}" || aws s3 mb "${UPLOAD_DEST}" || return
make kops-install dev-upload-linux-${KOPS_ARCH} || return

# Set KOPS_BASE_URL
(tools/get_version.sh | grep VERSION | awk '{print $2}') || return
KOPS_VERSION=$(tools/get_version.sh | grep VERSION | awk '{print $2}')
export KOPS_BASE_URL=https://${S3_BUCKET_NAME}.s3.amazonaws.com/kops/${KOPS_VERSION}/

# Create the state-store bucket if it doesn't exist
KOPS_STATE_STORE="s3://kops-state-${ACCOUNT_ID}-${USER}"
export KOPS_STATE_STORE
aws s3 ls "${KOPS_STATE_STORE}" || aws s3 mb "${KOPS_STATE_STORE}" || return

# Set feature flags needed on AWS
# export KOPS_FEATURE_FLAGS=

echo "SUCCESS"
