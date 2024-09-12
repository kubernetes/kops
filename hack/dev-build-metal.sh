#!/bin/bash

# Copyright 2024 The Kubernetes Authors.
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

# This is a convenience script for developing kOps on Metal.
# It builds the code, including nodeup, and uploads to our fake S3 storage server.
# It also sets KOPS_BASE_URL to point to that storage server.
# To use, source the script.  For example `. hack/dev-build-metal.sh` (note the initial `.`)

# Can't use set -e in a script we want to source
#set -e

#set -x

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}" || return

# Dev environments typically do not need to test multiple architectures
KOPS_ARCH=amd64
export KOPS_ARCH

# Configure aws cli to talk to local storage
aws configure --profile metal set aws_access_key_id accesskey
aws configure --profile metal set aws_secret_access_key secret
aws configure --profile metal set aws_region us-east-1
aws configure --profile metal set endpoint_url http://10.123.45.1:8443
export AWS_ENDPOINT_URL=http://10.123.45.1:8443
export AWS_PROFILE=metal
export AWS_REGION=us-east-1

# Avoid chunking in S3 uploads (not supported by our mock yet)
aws configure --profile metal set s3.multipart_threshold 64GB

export UPLOAD_DEST=s3://kops-dev-build/
aws --version
aws s3 ls "${UPLOAD_DEST}" || aws s3 mb "${UPLOAD_DEST}" || return
make kops-install dev-version-dist-${KOPS_ARCH} || return

hack/upload .build/upload/ "${UPLOAD_DEST}" || return

# Set KOPS_BASE_URL
(tools/get_version.sh | grep VERSION | awk '{print $2}') || return
KOPS_VERSION=$(tools/get_version.sh | grep VERSION | awk '{print $2}')
export KOPS_BASE_URL=http://10.123.45.1:8443/kops-dev-build/kops/${KOPS_VERSION}/
echo "set KOPS_BASE_URL=${KOPS_BASE_URL}"

# Set feature flags needed on Metal
# export KOPS_FEATURE_FLAGS=

echo "SUCCESS"
